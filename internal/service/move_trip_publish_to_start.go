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

func (s *TripService) MoveTripPublishedToStartedTx(
	ctx context.Context,
	req model.MoveTripPublishedToStartedModel,
) (res *model.MoveTripPublishedToStartedModelResponse, err error) {
	// 1. Integration with Jaeger: create a child span for this layer (ctx now contains the ID of this span)
	ctxSpc, span := otel.Tracer("TripService").Start(ctx, "TripService.MoveTripPublishedToStartedTx")

	//2. metrics - time fo Prometheus + Grafana (Integration with Prometheus)
	started := time.Now()
	result := "success"

	defer func() {
		// Now 'err' is taken from the return of the function. If an error occurred below in the stack, err != nil
		if err != nil {
			result = "error"
			span.RecordError(err)                    // Log error in Jaeger
			span.SetStatus(codes.Error, err.Error()) // Color the span in Jaeger in red color
		}

		// metrics - time for Prometheus + Grafana (Integration with Prometheus)
		s.metrics.TripDraftToPublishTotal.WithLabelValues(result).Inc()
		s.metrics.TripDraftToPublishDuration.WithLabelValues(result).
			Observe(time.Since(started).Seconds())

		span.End() // span always ends in the end. Jaeger will measure the time between Start and End!
	}()

	// 3. getting custom logger context
	logger := logctx.Logger(ctxSpc).With(
		slog.String("service", "TripService"),
		slog.String("operation", "MoveTripDraftToPublishTx"),
		slog.String("client_id", req.ID),
	)
	logger.Debug("move trip draft to publish transaction started")

	// 4. Open a database transaction
	// Mandatory to create a sub-span for the transaction to measure its clean duration
	txCtx, txSpan := otel.Tracer("database").Start(ctxSpc, "DB.Transaction")
	defer txSpan.End()

	res, err = tx(txCtx, s.pool, func(tx pgx.Tx) (*model.MoveTripPublishedToStartedModelResponse, error) {
		txLogger := logger.With(slog.String("layer", "transaction")) //added new key/value to logger
		txLogger.Debug("move trip published to started transaction execution started")

		resp, err := s.useCase.MoveTripPublishedToStartedTx(
			txCtx,
			tx,
			s.repo,
			model.MoveTripPublishedToStartedModel{
				ID:          req.ID,
				CompanyID:   req.CompanyID,
				ServiceCode: req.ServiceCode,
			})
		if err != nil {
			txLogger.Error("move trip published to started usecase failed", slog.Any("error", err))
			return nil, err
		}

		// outbox
		payload := model2.PayloadEvent{TripID: resp.ID}
		event := model2.Entity{
			EventName:   string(model2.EventPublished),
			AggregateId: resp.ID,
			Payload:     payload,
		}

		err = s.outboxRepo.CreateNotificationTripPublishTx(ctxSpc, tx, &event)
		if err != nil {
			txLogger.Error("move trip published to started outbox create notification failed", slog.Any("error", err))
			return nil, fmt.Errorf("error while MoveTripPublishedToStarted create Notification: %w", err)
		}

		txLogger.Debug("transaction execution completed", slog.String("trip_id", resp.ID.String()))
		return resp, nil
	})

	// 6. Fix the transaction error (if COMMIT failed or logic inside failed) using txSpan
	if err != nil {
		logger.Error("move trip published to started failed", slog.Any("error", err))

		txSpan.RecordError(err)
		txSpan.SetStatus(codes.Error, err.Error())
		// Span will be closed automatically through defer txSpan.End() above
		return nil, err
	}

	logger.Debug("move trip published to started completed", slog.String("trip_id", res.ID.String()))
	return res, nil
}
