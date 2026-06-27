package service

import (
	"context"
	"fmt"
	"go.opentelemetry.io/otel"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"job4j.ru/share_trip/internal/domain/trip/model"
	"job4j.ru/share_trip/internal/observability/logctx"
)

func (s *TripService) GetTripByID(
	ctx context.Context,
	req model.GetByIDModelRequest,
) (*model.GetTripByIDModelResponse, error) {
	//tracing
	ctx, span := otel.Tracer("TripService").Start(ctx, "TripService.GetTripByID")
	defer span.End()

	//getting custom logger context
	logger := logctx.Logger(ctx).With(
		slog.String("service", "TripService"),
		slog.String("operation", "GetTripByID"),
		slog.String("client_id", req.ID),
	)
	logger.Debug("get trip started")

	res, err := tx(ctx, s.pool, func(tx pgx.Tx) (*model.GetTripByIDModelResponse, error) {
		txLogger := logger.With(
			slog.String("layer", "transaction"),
		)
		txLogger.Debug("transaction started")

		resp, err := s.useCase.GetTripByID(ctx, tx, s.repo, req)
		if err != nil {
			return nil, fmt.Errorf("usecase.GetTripByID: %w", err)
		}

		txLogger.Debug(
			"transaction completed",
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

	//success
	logger.Debug(
		"get trip completed",
		slog.String("trip_id", res.ID.String()),
	)
	return res, nil
}
