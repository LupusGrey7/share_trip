package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"job4j.ru/share_trip/internal/observability/logctx"
	outboxdomain "job4j.ru/share_trip/internal/outbox/domain"
	"job4j.ru/share_trip/internal/trip/domain"

	"github.com/jackc/pgx/v5"
)

func (s *TripService) MoveTripDraftToPublish(
	ctx context.Context,
	req domain.MoveTripDraftToPublishModel,
) (res *domain.MoveTripDraftToPublishModelResponse, err error) {
	// 1. Integration with Jaeger: create a child span for this layer (ctx now contains the ID of this span)
	ctxSpc, span := otel.Tracer("TripService").Start(ctx, "TripService.MoveTripDraftToPublishTx")

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

	res, err = tx(txCtx, s.pool, func(tx pgx.Tx) (*domain.MoveTripDraftToPublishModelResponse, error) {
		txLogger := logger.With(slog.String("layer", "transaction")) //added new key/value to logger
		txLogger.Debug("move trip draft to publish transaction execution started")

		resp, err := s.useCase.MoveTripDraftToPublishTx(txCtx, tx, s.repo, req)
		if err != nil {
			txLogger.Error("move trip draft to publish usecase failed", slog.Any("error", err))
			return nil, err
		}

		// outbox
		payload := outboxdomain.PayloadEvent{TripID: resp.ID}
		event := outboxdomain.Entity{
			EventName:   string(outboxdomain.EventPublished),
			AggregateId: resp.ID,
			Payload:     payload,
		}

		err = s.outboxRepo.CreateNotificationTripPublishTx(ctxSpc, tx, &event)
		if err != nil {
			txLogger.Error("move trip draft to publish outbox create notification failed", slog.Any("error", err))
			return nil, fmt.Errorf("error while MoveTripDraftToPublish create Notification: %w", err)
		}

		txLogger.Debug("transaction execution completed", slog.String("trip_id", resp.ID.String()))
		return resp, nil
	})

	// 6. Fix the transaction error (if COMMIT failed or logic inside failed) using txSpan
	if err != nil {
		logger.Error("move trip draft to publish failed", slog.Any("error", err))

		txSpan.RecordError(err)
		txSpan.SetStatus(codes.Error, err.Error())
		// Span will be closed automatically through defer txSpan.End() above
		return nil, err
	}

	logger.Debug("move trip draft to publish completed", slog.String("trip_id", res.ID.String()))
	return res, nil
}
