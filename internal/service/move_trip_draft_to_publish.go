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
	"job4j.ru/share_trip/internal/domain/trip/usecase"
	"job4j.ru/share_trip/internal/observability/logctx"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"job4j.ru/share_trip/internal/clients/kafka"
)

func (s *TripService) MoveTripDraftToPublish(
	ctx context.Context,
	req model.MoveTripDraftToPublishModel,
) (res *model.MoveTripDraftToPublishModelResponse, err error) {
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

	// 4. Contract Service Client — outside transaction, check if service is allowed
	contractResult, err := s.useCase.CheckServiceAllowed(ctxSpc, req.CompanyID, string(req.ServiceCode))
	if err != nil {
		logger.Error("contract check failed", slog.Any("error", err))
		return nil, err
	}

	if !contractResult.IsAllowed() {
		businessDenyReason := contractResult.Reason
		if businessDenyReason == "" {
			businessDenyReason = "service is not allowed"
		}

		logger.Error("contract denied trip start", slog.String("reason", businessDenyReason))
		return nil, fmt.Errorf("%w: %s", usecase.ErrConflict, businessDenyReason)
	}
	eventID := uuid.New().String()
	createdAt := time.Now()

	// 5. Open a database transaction
	// Mandatory to create a sub-span for the transaction to measure its clean duration
	txCtx, txSpan := otel.Tracer("database").Start(ctxSpc, "DB.Transaction")
	defer txSpan.End()

	res, err = tx(txCtx, s.pool, func(tx pgx.Tx) (*model.MoveTripDraftToPublishModelResponse, error) {
		txLogger := logger.With(slog.String("layer", "transaction")) //added new key/value to logger
		txLogger.Debug("move trip draft to publish transaction execution started")

		resp, err := s.useCase.MoveTripDraftToPublishTx(txCtx, tx, s.repo, req)
		if err != nil {
			txLogger.Error("move trip draft to publish usecase failed", slog.Any("error", err))
			return nil, err
		}

		// 6. Create outbox event
		// outbox
		payload := model2.PayloadEvent{TripID: resp.ID}
		event := model2.Entity{
			EventID:     eventID,
			EventName:   string(model2.EventPublished),
			AggregateId: resp.ID,
			Payload:     payload,
			CreatedAt:   time.Now(),
		}

		err = s.outboxRepo.CreateOutboxEventTripPublishTx(ctxSpc, tx, &event)
		if err != nil {
			txLogger.Error("move trip draft to publish outbox create event failed", slog.Any("error", err))
			return nil, fmt.Errorf("error while MoveTripDraftToPublish create Outbox Event: %w", err)
		}

		txLogger.Debug("transaction execution completed", slog.String("trip_id", resp.ID.String()))
		return resp, nil
	})

	// 8. Fix the transaction error (if COMMIT failed or logic inside failed) using txSpan
	if err != nil {
		logger.Error("move trip draft to publish failed", slog.Any("error", err))

		txSpan.RecordError(err)
		txSpan.SetStatus(codes.Error, err.Error())
		// Span will be closed automatically through defer txSpan.End() above
		return nil, err
	}

	// 7. Publish event
	// 7.1. Publish event to Kafka (trip_id is key)
	// When trip goes from draft to published, ShareTrip publishes the event: -> TripPublished
	// IMPORTANT: event must be published after successful change of the aggregate state.
	err = s.kafka.PublishTripPublished(
		ctxSpc,
		kafka.TripPublished{
			TripID:     req.ID,
			DriverID:   res.DriverID.String(),
			CompanyID:  req.CompanyID,
			EventType:  kafka.EventTypePublished,
			EventID:    eventID,
			OccurredAt: createdAt,
		})
	if err != nil {
		// dual-write gap: trip is published in DB but event not sent
		// TODO: outbox poller will resend; for now log + return 200
		logger.Error("kafka publish failed after commit, event lost until poller",
			slog.String("trip_id", res.ID.String()),
			slog.Any("error", err),
		)
	}

	logger.Debug("move trip draft to publish completed", slog.String("trip_id", res.ID.String()))
	return res, nil
}
