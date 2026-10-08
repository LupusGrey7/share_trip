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
	contracts "job4j.ru/share_trip/internal/clients/http/contract"
	"job4j.ru/share_trip/internal/observability/logctx"
	"job4j.ru/share_trip/internal/trip/domain"
	"job4j.ru/share_trip/internal/trip/usecase"
)

func (s *TripService) MoveTripFromPublishedToStarted(
	ctx context.Context,
	req domain.MoveTripFromPublishedToStartedInput,
) (res *domain.MoveTripFromPublishedToStartedOutput, err error) {
	ctx, span := otel.Tracer("TripService").Start(ctx, "TripService.MoveTripFromPublishedToStarted")

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

	logger := logctx.Logger(ctx).With(
		slog.String("service", "TripService"),
		slog.String("operation", "MoveTripFromPublishedToStarted"),
		slog.String("trip_id", req.ID),
		slog.String("company_id", req.CompanyID),
		slog.String("service_code", string(req.ServiceCode)),
		slog.String("client_id", req.ClientID.String()),
	)

	contractResult, err := s.contractClient.CheckAvailableService(
		ctx,
		contracts.CheckServiceRequest{
			CompanyID:   req.CompanyID,
			ServiceCode: string(req.ServiceCode),
		},
	)
	if err != nil {
		return nil, err
	}

	if !contractResult.IsAllowed() {
		businessDenyReason := contractResult.Reason
		if businessDenyReason == "" {
			businessDenyReason = "service is not allowed"
		}

		return nil, fmt.Errorf("%w: %s", usecase.ErrConflict, businessDenyReason)
	}
	req.ContractCheck = &contractResult

	txCtx, txSpan := otel.Tracer("database").Start(ctx, "DB.Transaction")
	defer txSpan.End()

	res, err = tx(txCtx, s.pool, func(tx pgx.Tx) (*domain.MoveTripFromPublishedToStartedOutput, error) {
		txLogger := logger.With(slog.String("layer", "transaction"))
		txLogger.Debug("move trip from published to started transaction started", slog.String("trip_id", req.ID))

		resp, err := s.useCase.MoveTripFromPublishedToStarted(txCtx, tx, s.repo, req)
		if err != nil {
			return nil, err
		}

		txLogger.Debug("move trip from published to started transaction completed", slog.String("trip_id", req.ID))
		return resp, nil
	})

	if err != nil {
		return nil, err
	}

	logger.Debug("move trip from published to started completed", slog.String("trip_id", req.ID))
	return res, nil
}
