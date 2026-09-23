// scenario: MoveTripFromPublishedToStarted — Contract check outside tx, then short DB transaction.
package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"job4j.ru/share_trip/internal/observability/logctx"
	"job4j.ru/share_trip/internal/trip/domain"
	"job4j.ru/share_trip/internal/trip/usecase"
)

func (s *TripService) MoveTripFromPublishedToStarted(
	ctx context.Context,
	req domain.MoveTripFromPublishedToStartedInput,
) (res *domain.MoveTripFromPublishedToStartedOutput, err error) {
	ctxSpc, span := otel.Tracer("TripService").Start(ctx, "TripService.MoveTripFromPublishedToStarted")

	started := time.Now()
	result := "success"

	defer func() {
		if err != nil && !errors.Is(err, usecase.ErrAlreadyDone) {
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
		slog.String("operation", "MoveTripFromPublishedToStarted"),
		slog.String("trip_id", req.ID),
		slog.String("company_id", req.CompanyID),
		slog.String("service_code", string(req.ServiceCode)),
		slog.String("client_id", req.ClientID.String()),
	)

	contractResult, err := s.useCase.CheckServiceIsAllowed(ctxSpc, req.CompanyID, string(req.ServiceCode))
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
	req.ContractCheck = &contractResult

	txCtx, txSpan := otel.Tracer("database").Start(ctxSpc, "DB.Transaction")
	defer txSpan.End()

	res, err = tx(txCtx, s.pool, func(tx pgx.Tx) (*domain.MoveTripFromPublishedToStartedOutput, error) {
		txLogger := logger.With(slog.String("layer", "transaction"))

		resp, err := s.useCase.MoveTripFromPublishedToStarted(txCtx, tx, s.repo, req)
		if err != nil {
			if !errors.Is(err, usecase.ErrAlreadyDone) {
				txLogger.Error("move trip from published to started usecase failed", slog.Any("error", err))
			}
			return nil, err
		}

		return resp, nil
	})

	if err != nil {
		if !errors.Is(err, usecase.ErrAlreadyDone) {
			logger.Error("move trip from published to started failed", slog.Any("error", err))
			txSpan.RecordError(err)
			txSpan.SetStatus(codes.Error, err.Error())
		}
		return nil, err
	}

	return res, nil
}
