package usecase

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"go.opentelemetry.io/otel"
	"job4j.ru/share_trip/internal/observability/logctx"

	"github.com/jackc/pgx/v5"
	"job4j.ru/share_trip/internal/domain/trip/model"
	"job4j.ru/share_trip/internal/repository"
)

func (t *TripUseCase) MoveTripPublishedToStartTx(
	ctx context.Context,
	tx pgx.Tx,
	repo repository.BaseTxTripRepository,
	req model.MoveTripPublishedToStartModel,
) (*model.MoveTripPublishedToStartModelResponse, error) {
	//tracing Jaeger
	ctxSpc, span := otel.Tracer("TripUseCase").Start(ctx, "TripUseCase.MoveTripPublishedToStartTx")
	defer span.End()

	//getting custom logger context
	logger := logctx.Logger(ctxSpc).With(
		slog.String("layer", "useCase"),
		slog.String("useCase", "TripUseCase.MoveTripPublishedToStartTx"),
		slog.String("trip_id", req.ID),
		slog.String("company_id", req.CompanyID),
		slog.String("service_code", string(req.ServiceCode)),
	)
	logger.Debug("move trip published to start transaction useCase started")

	//Before moving from draft to published ShareTrip should check if the company has access to the trip_publication service.
	//If the Contract Service allows the action, ShareTrip moves the trip to the published state.
	//If the Contract Service prohibits the action, the trip status remains unchanged.

	//Check if the company has access to the trip_publication service.
	contractResult, err := t.contractUsecase.CheckAvailableService(ctxSpc, req.CompanyID, string(req.ServiceCode))
	if err != nil {
		logger.Error("check service useCase failed", slog.Any("error", err))
		return nil, err
	}

	if !contractResult.IsAllowed() {
		logger.Error("check service useCase failed", slog.Any("error", err))
		return nil, fmt.Errorf("%w: service is not allowed", ErrConflict)
	}

	// If the permission is present, we move the trip to the published status and update the publication date
	resp, err := repo.GetForUpdateByIDTx(ctxSpc, tx, req.ID)
	if err != nil {
		if errors.Is(err, repository.ErrTripNotFound) {
			logger.Error("move trip published to start transaction get trip repository failed", slog.Any("error", err))
			return nil, ErrTripNotFound
		}
		// If this is not a ErrEntityNotFound, This means it's a system failure (500 error)
		logger.Error("move trip published to start transaction get trip repository failed", slog.String("error", err.Error()))
		return nil, err
	}

	if resp.DriverID != req.ClientID {
		logger.Error("move trip published to start useCase failed", slog.Any("error", err))
		return nil, fmt.Errorf("%w: client %s is not driver of trip %s", ErrForbidden, req.ClientID, req.ID)
	}

	if resp.Status == model.StatusPublished {
		logger.Debug("move trip published to start transaction useCase completed", slog.String("trip_id", resp.ID.String()))
		return &model.MoveTripPublishedToStartModelResponse{ID: resp.ID}, nil
	}

	if resp.Status != model.StatusDraft {
		logger.Error(
			"move published to start useCase failed",
			slog.Any("error", err),
		)
		return nil, fmt.Errorf("%w: invalid entity status: expected %s", ErrConflict, model.StatusDraft)
	}

	resp.Status = model.StatusPublished

	updatedTrip, err := repo.UpdateTripTx(ctxSpc, tx, resp)
	if err != nil {
		logger.Error("update trip repository failed", slog.Any("error", err))
		return nil, err
	}

	logger.Debug("move published to start completed", slog.String("trip_id", resp.ID.String()))
	return updatedTrip.ToPublishedStartModelResponse(contractResult.Allowed, contractResult.Reason), nil
}
