package service

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"job4j.ru/share_trip/internal/domain/trip/model"
	"job4j.ru/share_trip/internal/observability/logctx"
	"log/slog"
)

func (s *TripService) CreateTripWithTx(ctx context.Context, req model.CreateTripRequestModel) (*model.CreateTripResponse, error) {
	//metrics
	started := time.Now()
	result := "success"

	defer func() {
		s.metrics.TripCreateTotal.WithLabelValues(result).Inc()
		s.metrics.TripCreateDuration.WithLabelValues(result).
			Observe(time.Since(started).Seconds())
	}()

	//getting custom logger context
	logger := logctx.Logger(ctx).With(
		slog.String("service", "TripService"),
		slog.String("operation", "CreateTrip"),
		slog.String("client_id", req.DriverID.String()),
	)
	logger.Info("create trip started")

	res, err := tx(ctx, s.pool, func(tx pgx.Tx) (*model.CreateTripResponse, error) {
		txLogger := logger.With(
			slog.String("layer", "transaction"),
		)
		txLogger.Info("transaction started")

		resp, err := s.useCase.CreateTrip(ctx, tx, s.repo, req)

		if err != nil {
			txLogger.Error(
				"create trip usecase failed",
				slog.Any("error", err),
			)
			return nil, fmt.Errorf("usecase.CreateTrip: %w", err)
		}

		txLogger.Info(
			"transaction completed",
			slog.String("trip_id", resp.ID.String()),
		)
		return resp, nil
	})

	if err != nil {
		logger.Error(
			"create trip failed",
			slog.Any("error", err),
		)
		return nil, err
	}

	logger.Info(
		"create trip completed",
		slog.String("trip_id", res.ID.String()),
	)
	return res, nil
}
