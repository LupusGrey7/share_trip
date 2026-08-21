package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"

	"github.com/jackc/pgx/v5"
	"job4j.ru/share_trip/internal/observability/logctx"
	"job4j.ru/share_trip/internal/trip/domain"
)

func (s *TripService) GetTripByID(ctx context.Context, req domain.GetByIDModelRequest) (res *domain.GetTripByIDModelResponse, err error) {
	//1. Integration with Jaeger: create a child span for this layer
	ctx, span := otel.Tracer("TripService").Start(ctx, "TripService.GetTripByID")

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
		s.metrics.TripGetByIDTotal.WithLabelValues(result).Inc()
		s.metrics.TripGetByIDDuration.WithLabelValues(result).
			Observe(time.Since(started).Seconds())

		span.End() // span always ends in the end. Jaeger will measure the time between Start and End!
	}()

	// getting custom logger context for logging
	logger := logctx.Logger(ctx).With(
		slog.String("service", "TripService"),
		slog.String("operation", "GetTripByID"),
		slog.String("client_id", req.ID),
	)
	logger.Debug("get trip by ID started")

	txCtx, txSpan := otel.Tracer("database").Start(ctx, "DB.Transaction")
	defer txSpan.End()

	res, err = tx(txCtx, s.pool, func(tx pgx.Tx) (*domain.GetTripByIDModelResponse, error) {
		txLogger := logger.With(
			slog.String("layer", "transaction"),
		)
		txLogger.Debug("transaction execution started")

		resp, err := s.useCase.GetTripByIDTx(txCtx, tx, s.repo, req)
		if err != nil {
			txLogger.Error("get trip by ID useCase failed", slog.Any("error", err))
			return nil, fmt.Errorf("useCase.GetTripByID: %w", err)
		}

		txLogger.Debug("transaction get trip by ID completed", slog.String("trip_id", resp.ID.String()))
		return resp, nil
	})

	// 5. Use txSpan to fix the transaction error (if COMMIT failed or logic inside failed)
	if err != nil {
		logger.Error("get trip by ID failed", slog.Any("error", err))
		txSpan.RecordError(err)
		txSpan.SetStatus(codes.Error, err.Error())
		// Span will be closed automatically through defer txSpan.End() above
		return nil, err
	}

	logger.Debug("get trip by ID completed", slog.String("trip_id", res.ID.String()))
	return res, nil
}
