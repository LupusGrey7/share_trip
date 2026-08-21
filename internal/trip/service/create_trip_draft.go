package service

import (
	"context"
	"fmt"
	"time"

	"log/slog"

	"github.com/jackc/pgx/v5"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"job4j.ru/share_trip/internal/observability/logctx"
	"job4j.ru/share_trip/internal/trip/domain"
)

func (s *TripService) CreateTripDraft(
	ctx context.Context,
	req domain.CreateTripRequestModel,
) (res *domain.CreateTripDraftResponse, err error) {
	// 1. Integration with Jaeger: create a child span for this layer (ctx now contains the ID of this span)
	ctxSpc, span := otel.Tracer("TripService").Start(ctx, "TripService.CreateTripDraft")

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
		s.metrics.TripCreateTotal.WithLabelValues(result).
			Inc() // Increment the counter for the result
		s.metrics.TripCreateDuration.WithLabelValues(result).
			Observe(time.Since(started).Seconds()) // Observe the duration of the operation

		span.End() // Span always ends in the end. Jaeger will measure the time between Start and End!
	}()

	//getting custom logger context
	logger := logctx.Logger(ctxSpc).With(
		slog.String("service", "TripService"),
		slog.String("operation", "CreateTripDraft"),
		slog.String("client_id", req.DriverID.String()),
	)
	logger.Debug("create trip draft started")

	// 4. Open a database transaction
	// Mandatory to create a sub-span for the transaction to measure its clean duration
	txCtx, txSpan := otel.Tracer("database").Start(ctxSpc, "DB.Transaction")
	defer txSpan.End()

	res, err = tx(txCtx, s.pool, func(tx pgx.Tx) (*domain.CreateTripDraftResponse, error) {
		txLogger := logger.With(slog.String("layer", "transaction"))
		txLogger.Debug("transaction create trip draft execution started")

		resp, err := s.useCase.CreateTripDraftTx(txCtx, tx, s.repo, req)

		if err != nil {
			txLogger.Error("create trip draft usecase failed", slog.Any("error", err))
			return nil, fmt.Errorf("usecase.CreateTripDraft: %w", err)
		}

		txLogger.Debug("transaction create trip draft completed", slog.String("trip_id", resp.ID.String()))
		return resp, nil
	})

	if err != nil {
		logger.Error("create trip draft failed", slog.Any("error", err))

		txSpan.RecordError(err)
		txSpan.SetStatus(codes.Error, err.Error())
		// Span will be closed automatically through defer txSpan.End() above
		return nil, err

	}

	logger.Debug("create trip draft completed", slog.String("trip_id", res.ID.String()))
	return res, nil
}
