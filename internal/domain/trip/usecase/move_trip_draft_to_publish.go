package usecase

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/gofiber/fiber/v2/log"
	"job4j.ru/share_trip/internal/domain/trip/model"
	"job4j.ru/share_trip/internal/observability/logctx"

	"github.com/jackc/pgx/v5"
	"job4j.ru/share_trip/internal/repository"
)

func (t *TripUseCase) MoveTripDraftToPublishTx(
	ctx context.Context,
	tx pgx.Tx,
	repo repository.BaseTxTripRepository,
	req model.MoveTripDraftToPublishModel,
) (*model.MoveTripDraftToPublishModelResponse, error) {
	//getting custom logger context
	logger := logctx.Logger(ctx).With(
		slog.String("layer", "usecase"),
		slog.String("usecase", "TripUsecase.GetTripByID"),
		slog.String("client_id", req.ID),
	)
	logger.Info("move draft to publish usecase started")

	resp, err := repo.GetForUpdateByIDTx(ctx, tx, req.ID)
	if err != nil {
		if errors.Is(err, repository.ErrTripNotFound) {
			logger.Error(
				"repository get trip failed",
				slog.Any("error", err),
			)
			return nil, ErrTripNotFound
		}
		// If this is not a ErrEntityNotFound, This means it's a system failure (500 error)
		return nil, err
	}

	if resp.DriverID != req.ClientID {
		logger.Error(
			"move draft to publish usecase failed",
			slog.Any("error", err),
		)
		return nil, fmt.Errorf("%w: client %s is not driver of trip %s", ErrForbidden, req.ClientID, req.ID)
	}

	if resp.Status == model.StatusPublished {
		return &model.MoveTripDraftToPublishModelResponse{
			ID: resp.ID,
		}, nil
	}

	if resp.Status != model.StatusDraft {
		logger.Error(
			"move draft to publish usecase failed",
			slog.Any("error", err),
		)
		return nil, fmt.Errorf("%w: invalid entity status: expected %s", ErrConflict, model.StatusDraft)
	}

	resp.Status = model.StatusPublished

	updatedTrip, err := repo.UpdateTripTx(ctx, tx, resp)
	if err != nil {
		logger.Error(
			"repository update trip failed",
			slog.Any("error", err),
		)
		return nil, err
	}

	log.Info("move draft to publish completed",
		slog.String("trip_id", resp.ID.String()),
	)
	return updatedTrip.UpdateToPublishModelResponse(), nil
}
