package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"job4j.ru/share_trip/internal/observability/logctx"
	"job4j.ru/share_trip/internal/trip/domain"
)

func (s *TripService) GetTripByID(
	ctx context.Context,
	input *domain.GetByIDInput,
) (res *domain.GetTripByIDOutput, err error) {
	ctx, span := otel.Tracer("TripService").Start(ctx, "TripService.GetTripByID")

	started := time.Now()
	result := "success"

	defer func() {
		if err != nil {
			result = "error"
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}
		s.metrics.TripGetByIDTotal.WithLabelValues(result).Inc()
		s.metrics.TripGetByIDDuration.WithLabelValues(result).
			Observe(time.Since(started).Seconds())
		span.End()
	}()

	logger := logctx.Logger(ctx).With(
		slog.String("service", "TripService"),
		slog.String("operation", "GetTripByID"),
		slog.String("trip_id", input.ID),
	)
	logger.Debug("get trip by ID started")

	txCtx, txSpan := otel.Tracer("database").Start(ctx, "DB.Transaction")
	defer txSpan.End()

	res, err = tx(txCtx, s.pool, func(tx pgx.Tx) (*domain.GetTripByIDOutput, error) {
		txLogger := logger.With(slog.String("layer", "transaction"))
		txLogger.Debug("transaction execution started")

		resp, err := s.useCase.GetTripByIDTx(txCtx, tx, s.repo, input)
		if err != nil {
			txLogger.Error("get trip by ID useCase failed", slog.Any("error", err))
			return nil, fmt.Errorf("useCase.GetTripByID: %w", err)
		}

		txLogger.Debug("transaction get trip by ID completed", slog.String("trip_id", resp.ID.String()))
		return resp, nil
	})

	if err != nil {
		logger.Error("get trip by ID failed", slog.Any("error", err))
		txSpan.RecordError(err)
		txSpan.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	logger.Debug("get trip by ID completed", slog.String("trip_id", res.ID.String()))
	return res, nil
}
