package usecase

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"go.opentelemetry.io/otel"
	"job4j.ru/share_trip/internal/client/http/contracts"
	"job4j.ru/share_trip/internal/domain/trip/model"
	"job4j.ru/share_trip/internal/observability/logctx"
	"job4j.ru/share_trip/internal/repository"
)

// MoveTripPublishedToStartedTx applies published→started under an already-open short transaction.
// Contract must already be checked by the service (outside tx).
func (t *TripUseCase) MoveTripPublishedToStartedTx(
	ctx context.Context,
	tx pgx.Tx,
	repo repository.BaseTxTripRepository,
	req model.MoveTripPublishedToStartedModel,
	contractResult contracts.CheckResult,
) (*model.MoveTripPublishedToStartedModelResponse, error) {
	ctxSpc, span := otel.Tracer("TripUseCase").Start(ctx, "TripUseCase.MoveTripPublishedToStartedTx")
	defer span.End()

	logger := logctx.Logger(ctxSpc).With(
		slog.String("layer", "useCase"),
		slog.String("useCase", "TripUseCase.MoveTripPublishedToStartedTx"),
		slog.String("trip_id", req.ID),
		slog.String("company_id", req.CompanyID),
		slog.String("service_code", string(req.ServiceCode)),
		slog.String("client_id", req.ClientID.String()),
	)
	logger.Debug("move trip published to started transaction useCase started")

	if !contractResult.IsAllowed() {
		logger.Error("move trip published to started denied by contract",
			slog.String("reason", contractResult.Reason),
		)
		return nil, fmt.Errorf("%w: service is not allowed", ErrConflict)
	}

	resp, err := repo.GetForUpdateByIDTx(ctxSpc, tx, req.ID)
	if err != nil {
		if errors.Is(err, repository.ErrTripNotFound) {
			logger.Error("get trip for update failed", slog.Any("error", err))
			return nil, ErrTripNotFound
		}
		logger.Error("get trip for update failed", slog.String("error", err.Error()))
		return nil, err
	}

	if resp.DriverID != req.ClientID {
		logger.Error("client is not driver of trip",
			slog.String("client_id", req.ClientID.String()),
			slog.String("driver_id", resp.DriverID.String()),
		)
		return nil, fmt.Errorf("%w: client %s is not driver of trip %s", ErrForbidden, req.ClientID, req.ID)
	}

	// Idempotent: already started → empty DriverID for handler 204
	if resp.Status == model.StatusStarted {
		logger.Debug("trip already started", slog.String("trip_id", resp.ID.String()))
		return &model.MoveTripPublishedToStartedModelResponse{ID: resp.ID}, nil
	}

	if resp.Status != model.StatusPublished {
		logger.Error("invalid trip status for start",
			slog.String("status", string(resp.Status)),
		)
		return nil, fmt.Errorf("%w: invalid entity status: expected %s", ErrConflict, model.StatusPublished)
	}

	resp.Status = model.StatusStarted

	updatedTrip, err := repo.UpdateTripTx(ctxSpc, tx, resp)
	if err != nil {
		logger.Error("update trip to started failed", slog.Any("error", err))
		return nil, err
	}

	logger.Debug("move published to started completed", slog.String("trip_id", resp.ID.String()))
	return updatedTrip.ToPublishedStartModelResponse(contractResult.Allowed, contractResult.Reason), nil
}
