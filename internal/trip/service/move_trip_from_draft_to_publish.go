package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"job4j.ru/share_trip/internal/clients/kafka"
	"job4j.ru/share_trip/internal/observability/logctx"
	outboxdomain "job4j.ru/share_trip/internal/outbox/domain"
	"job4j.ru/share_trip/internal/trip/domain"
	"job4j.ru/share_trip/internal/trip/usecase"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (s *TripService) MoveTripFromDraftToPublish(
	ctx context.Context,
	req domain.MoveTripFromDraftToPublishInput,
) (res *domain.MoveTripFromDraftToPublishOutput, err error) {
	ctxSpc, span := otel.Tracer("TripService").Start(ctx, "TripService.MoveTripFromDraftToPublish")

	started := time.Now()
	result := "success"

	defer func() {
		if err != nil && !errors.Is(err, usecase.ErrAlreadyDone) {
			result = "error"
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}

		s.metrics.TripDraftToPublishTotal.WithLabelValues(result).Inc()
		s.metrics.TripDraftToPublishDuration.WithLabelValues(result).
			Observe(time.Since(started).Seconds())

		span.End()
	}()

	logger := logctx.Logger(ctxSpc).With(
		slog.String("service", "TripService"),
		slog.String("operation", "MoveTripFromDraftToPublish"),
		slog.String("client_id", req.ID),
	)
	logger.Debug("move trip from draft to publish started")

	eventID := uuid.New().String()
	occurredAt := time.Now()

	txCtx, txSpan := otel.Tracer("database").Start(ctxSpc, "DB.Transaction")
	defer txSpan.End()

	res, err = tx(txCtx, s.pool, func(tx pgx.Tx) (*domain.MoveTripFromDraftToPublishOutput, error) {
		txLogger := logger.With(slog.String("layer", "transaction"))
		txLogger.Debug("move trip from draft to publish transaction execution started")

		resp, err := s.useCase.MoveTripFromDraftToPublish(txCtx, tx, s.repo, req)
		if err != nil {
			if !errors.Is(err, usecase.ErrAlreadyDone) {
				txLogger.Error("move trip from draft to publish usecase failed", slog.Any("error", err))
			}
			return nil, err
		}

		payload := outboxdomain.PayloadEvent{TripID: resp.ID}
		event := outboxdomain.Entity{
			EventID:     eventID,
			EventName:   string(outboxdomain.EventPublished),
			AggregateId: resp.ID,
			Payload:     payload,
			CreatedAt:   occurredAt,
		}

		err = s.outboxRepo.CreateEventWhenTripMovesFromDraftToPublishedTx(ctxSpc, tx, &event)
		if err != nil {
			txLogger.Error("move trip from draft to publish outbox create event failed", slog.Any("error", err))
			return nil, fmt.Errorf("error while MoveTripFromDraftToPublish create Outbox Event: %w", err)
		}

		txLogger.Debug("transaction execution completed", slog.String("trip_id", resp.ID.String()))
		return resp, nil
	})

	if err != nil {
		if !errors.Is(err, usecase.ErrAlreadyDone) {
			logger.Error("move trip from draft to publish failed", slog.Any("error", err))
			txSpan.RecordError(err)
			txSpan.SetStatus(codes.Error, err.Error())
		}
		return nil, err
	}

	if pubErr := s.kafka.PublishTripPublished(
		ctxSpc,
		kafka.TripPublished{
			EventID:    eventID,
			EventType:  kafka.EventTypePublished,
			OccurredAt: occurredAt,
			Payload: kafka.TripPublishedPayload{
				TripID:    res.ID.String(),
				DriverID:  res.DriverID.String(),
				CompanyID: req.CompanyID,
			},
		}); pubErr != nil {
		logger.Error("kafka publish failed after commit, event lost until poller",
			slog.String("trip_id", res.ID.String()),
			slog.String("event_id", eventID),
			slog.Any("error", pubErr),
		)
	}

	logger.Debug("move trip from draft to publish completed", slog.String("trip_id", res.ID.String()))
	return res, nil
}
