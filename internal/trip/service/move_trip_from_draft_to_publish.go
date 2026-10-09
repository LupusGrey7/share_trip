// scenario: MoveTripFromDraftToPublish — Contract check outside tx, then short DB transaction.

package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
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
	ctx, span := otel.Tracer("TripService").Start(ctx, "TripService.MoveTripFromDraftToPublish")

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

	logger := logctx.Logger(ctx).With(
		slog.String("service", "TripService"),
		slog.String("operation", "MoveTripFromDraftToPublish"),
		slog.String("client_id", req.ID),
	)

	eventID := uuid.New()
	occurredAt := time.Now()

	txCtx, txSpan := otel.Tracer("database").Start(ctx, "DB.Transaction")
	defer txSpan.End()

	res, err = tx(txCtx, s.pool, func(tx pgx.Tx) (*domain.MoveTripFromDraftToPublishOutput, error) {
		txLogger := logger.With(slog.String("layer", "transaction"))
		txLogger.Debug("move trip from draft to publish transaction started", slog.String("trip_id", req.ID))

		resp, err := s.useCase.MoveTripFromDraftToPublish(txCtx, tx, s.repo, req)
		if err != nil {
			return nil, err
		}

		event := outboxdomain.Entity{
			ID:            eventID,
			AggregateType: outboxdomain.AggregateTypeTrip,
			AggregateID:   resp.ID,
			EventType:     string(outboxdomain.EventPublished),
			Payload: outboxdomain.PayloadEvent{
				TripID:    resp.ID.String(),
				DriverID:  resp.DriverID.String(),
				CompanyID: req.CompanyID,
			},
			Status:    outboxdomain.StatusPending,
			Attempts:  0,
			CreatedAt: occurredAt,
		}

		err = s.outboxRepo.CreateEvent(txCtx, tx, &event)
		if err != nil {
			return nil, fmt.Errorf("error while MoveTripFromDraftToPublish create outbox event: %w", err)
		}

		txLogger.Debug("move trip from draft to publish transaction completed", slog.String("trip_id", req.ID))
		return resp, nil
	})

	if err != nil {
		return nil, err
	}

	logger.Debug("move trip from draft to publish completed", slog.String("trip_id", req.ID))
	return res, nil
}
