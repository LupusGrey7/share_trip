package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	model2 "job4j.ru/share_trip/internal/domain/outbox/model"
	"job4j.ru/share_trip/internal/domain/trip/model"
	"job4j.ru/share_trip/internal/observability/logctx"

	"github.com/jackc/pgx/v5"
)

func (s *TripService) MoveTripDraftToPublish(
	ctx context.Context,
	req model.MoveTripDraftToPublishModel,
) (*model.MoveTripDraftToPublishModelResponse, error) {
	//metrics
	started := time.Now()
	result := "success"

	defer func() {
		s.metrics.TripPublishTotal.WithLabelValues(result).Inc()
		s.metrics.TripPublishDuration.WithLabelValues(result).
			Observe(time.Since(started).Seconds())
	}()

	//getting custom logger context
	logger := logctx.Logger(ctx).With(
		slog.String("service", "TripService"),
		slog.String("operation", "CreateTrip"),
		slog.String("client_id", req.ID),
	)
	logger.Info("move trip to publish started")

	res, err := tx(ctx, s.pool, func(tx pgx.Tx) (*model.MoveTripDraftToPublishModelResponse, error) {
		txLogger := logger.With(
			slog.String("layer", "transaction"), //added new key/value to logger
		)
		txLogger.Info("transaction started")

		resp, err := s.useCase.MoveTripDraftToPublishTx(ctx, tx, s.repo, req)
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
			return nil, fmt.Errorf("err create Event Notification: %w", err)
		}

		txLogger.Info(
			"transaction completed",
			slog.String("trip_id", resp.ID.String()),
		)
		return resp, nil
	})

	if err != nil {
		logger.Error(
			"move trip to publish failed",
			slog.Any("error", err),
		)
		return nil, err
	}

	//success
	logger.Info(
		"get trip completed",
		slog.String("trip_id", res.ID.String()),
	)
	return res, nil
}
