package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	model2 "job4j.ru/share_trip/internal/domain/outbox/model"
	"job4j.ru/share_trip/internal/domain/trip/model"
	"job4j.ru/share_trip/internal/observability/logctx"

	"github.com/jackc/pgx/v5"
)

func (s *TripService) MoveTripDraftToPublish(
	cx context.Context,
	req model.MoveTripDraftToPublishModel,
) (res *model.MoveTripDraftToPublishModelResponse, err error) {
	// 1. Интеграция с Jaeger: создаем дочерний спан для этого слоя(ctx теперь содержит ID этого спана)
	ctx, span := otel.Tracer("trip-service").Start(cx, "TripService.PublishTrip")

	//2. metrics - time fo Prometheus + Grafana (Интеграция с Prometheus)
	started := time.Now()
	result := "success"

	defer func() {
		// Теперь 'err' берется из возврата функции. Если ниже по стеку упала ошибка, err != nil
		if err != nil {
			result = "error"
			span.RecordError(err)                    // Логируем ошибку в Егерь
			span.SetStatus(codes.Error, err.Error()) // Красим спан в Егере в красный цвет
		}

		// Фиксируем метрики в Prometheus
		s.metrics.TripPublishTotal.WithLabelValues(result).Inc()
		s.metrics.TripPublishDuration.WithLabelValues(result).
			Observe(time.Since(started).Seconds())

		span.End() // Закрываем спан ВСЕГДА в самом конце
	}()

	// 3. getting custom logger context
	logger := logctx.Logger(ctx).With(
		slog.String("service", "TripService"),
		slog.String("operation", "MoveTripDraftToPublish"),
		slog.String("client_id", req.ID),
	)
	logger.Debug("move trip draft to publish started")

	// 4. Открываем транзакцию базы данных
	// Обязательно создаем под-спан для транзакции, чтобы замерить ее чистую длительность
	txCtx, txSpan := otel.Tracer("database").Start(ctx, "DB.Transaction")
	defer txSpan.End()

	res, err = tx(txCtx, s.pool, func(tx pgx.Tx) (*model.MoveTripDraftToPublishModelResponse, error) {
		txLogger := logger.With(
			slog.String("layer", "transaction"), //added new key/value to logger
		)
		txLogger.Debug("transaction execution started")

		resp, err := s.useCase.MoveTripDraftToPublishTx(txCtx, tx, s.repo, req)
		if err != nil {
			return nil, err
		}

		//outbox
		payload := model2.PayloadEvent{TripID: resp.ID}
		event := model2.Entity{
			EventName:   string(model2.EventPublished),
			AggregateId: resp.ID,
			Payload:     payload,
		}

		err = s.outboxRepo.CreateNotificationTripPublishTx(ctx, tx, &event)
		if err != nil {
			return nil, fmt.Errorf("err MoveTripDraftToPublish create Notification: %w", err)
		}

		txLogger.Debug(
			"transaction execution completed",
			slog.String("trip_id", resp.ID.String()),
		)
		return resp, nil
	})

	// business
	if err != nil {
		logger.Error(
			"MoveTripDraftToPublish failed",
			slog.Any("error", err),
		)
		return nil, err
	}

	// 5. Используем txSpan для фиксации ошибки транзакции (если упал COMMIT или логика внутри)
	if err != nil {
		txSpan.RecordError(err)
		txSpan.SetStatus(codes.Error, err.Error())
		// Span закроется автоматически через defer txSpan.End() выше
		return nil, fmt.Errorf("transaction failed: %w", err)
	}

	logger.Debug(
		"MoveTripDraftToPublish completed",
		slog.String("trip_id", res.ID.String()),
	)
	return res, nil
}
