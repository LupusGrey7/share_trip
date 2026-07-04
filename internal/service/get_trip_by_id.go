package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"

	"github.com/jackc/pgx/v5"
	"job4j.ru/share_trip/internal/domain/trip/model"
	"job4j.ru/share_trip/internal/observability/logctx"
)

func (s *TripService) GetTripByID(ctx context.Context, req model.GetByIDModelRequest) (res *model.GetTripByIDModelResponse, err error) {
	//1. Интеграция с Jaeger: создаем дочерний спан для этого слоя
	ctx, span := otel.Tracer("TripService").Start(ctx, "TripService.GetTripByID")

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

    		span.End() // Закрываем спан ВСЕГДА в самом конце. Егерь сам замерит время между Start и End!
    	}()

	//getting custom logger context
	logger := logctx.Logger(ctx).With(
		slog.String("service", "TripService"),
		slog.String("operation", "GetTripByID"),
		slog.String("client_id", req.ID),
	)
	logger.Debug("get trip started")

	txCtx, txSpan := otel.Tracer("database").Start(ctx, "DB.Transaction")
	defer txSpan.End()

	res, err = tx(txCtx, s.pool, func(tx pgx.Tx) (*model.GetTripByIDModelResponse, error) {
		txLogger := logger.With(
			slog.String("layer", "transaction"),
		)
		txLogger.Debug("transaction execution started")

		resp, err := s.useCase.GetTripByID(ctx, tx, s.repo, req)
		if err != nil {
			return nil, fmt.Errorf("useCase.GetTripByID: %w", err)
		}

		txLogger.Debug(
			"transaction execution completed",
			slog.String("trip_id", resp.ID.String()),
		)
		return resp, nil
	})

	if err != nil {
		logger.Error(
			"get trip failed",
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

	//success
	logger.Debug(
		"get trip by ID completed",
		slog.String("trip_id", res.ID.String()),
	)
	return res, nil
}
