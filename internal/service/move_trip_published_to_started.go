package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"job4j.ru/share_trip/internal/domain/trip/model"
	"job4j.ru/share_trip/internal/domain/trip/usecase"
	"job4j.ru/share_trip/internal/observability/logctx"
)

// MoveTripPublishedToStarted: Contract outside tx (lead), then short DB transaction.
func (s *TripService) MoveTripPublishedToStarted(
	ctx context.Context,
	req model.MoveTripPublishedToStartedModel,
) (res *model.MoveTripPublishedToStartedModelResponse, err error) {
	ctxSpc, span := otel.Tracer("TripService").Start(ctx, "TripService.MoveTripPublishedToStarted")

	started := time.Now()
	result := "success"

	defer func() {
		if err != nil {
			result = "error"
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}
		s.metrics.TripPublishedToStartTotal.WithLabelValues(result).Inc()
		s.metrics.TripPublishedToStartDuration.WithLabelValues(result).
			Observe(time.Since(started).Seconds())
		span.End()
	}()

	logger := logctx.Logger(ctxSpc).With(
		slog.String("service", "TripService"),
		slog.String("operation", "MoveTripPublishedToStarted"),
		slog.String("trip_id", req.ID),
		slog.String("company_id", req.CompanyID),
		slog.String("service_code", string(req.ServiceCode)),
		slog.String("client_id", req.ClientID.String()),
	)
	logger.Debug("move trip published to started started")

	// 1) Contract Service — outside transaction (lead sequence)
	contractResult, err := s.useCase.CheckStartAllowed(ctxSpc, req.CompanyID, string(req.ServiceCode))
	if err != nil {
		logger.Error("contract check failed", slog.Any("error", err))
		return nil, err
	}
	if !contractResult.IsAllowed() {
		logger.Error("contract denied trip start", slog.String("reason", contractResult.Reason))
		return nil, fmt.Errorf("%w: service is not allowed", usecase.ErrConflict)
	}

	// 2) Short DB transaction: FOR UPDATE → status → commit
	txCtx, txSpan := otel.Tracer("database").Start(ctxSpc, "DB.Transaction")
	defer txSpan.End()

	res, err = tx(txCtx, s.pool, func(tx pgx.Tx) (*model.MoveTripPublishedToStartedModelResponse, error) {
		txLogger := logger.With(slog.String("layer", "transaction"))
		txLogger.Debug("move trip published to started transaction execution started")

		resp, err := s.useCase.MoveTripPublishedToStartedTx(txCtx, tx, s.repo, req, contractResult)
		if err != nil {
			txLogger.Error("move trip published to started usecase failed", slog.Any("error", err))
			return nil, err
		}

		txLogger.Debug("transaction execution completed", slog.String("trip_id", resp.ID.String()))
		return resp, nil
	})

	if err != nil {
		logger.Error("move trip published to started failed", slog.Any("error", err))
		txSpan.RecordError(err)
		txSpan.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	logger.Debug("move trip published to started completed", slog.String("trip_id", res.ID.String()))
	return res, nil
}
