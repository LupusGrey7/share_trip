package usecase

import (
	"context"
	"errors"
	"fmt"
	"github.com/gofiber/fiber/v2/log"
	"job4j.ru/share_trip/internal/domain/trip/model"

	"github.com/jackc/pgx/v5"
	"job4j.ru/share_trip/internal/repository"
)

func (t *TripUsecase) MoveTripDraftToPublishTx(
	ctx context.Context,
	tx pgx.Tx,
	repo repository.BaseTxTripRepository,
	req model.MoveTripDraftToPublishModel,
) (*model.MoveTripDraftToPublishModelResponse, error) {

	resp, err := repo.GetForUpdateByIDTx(ctx, tx, req.ID)

	if err != nil {
		if errors.Is(err, repository.ErrTripNotFound) {
			return nil, ErrTripNotFound
		}
		// Если это не ErrEntityNotFound, значит это системный сбой (500 ошибка)
		return nil, fmt.Errorf("failed to get entity: %w", err)
	}

	if resp.DriverID != req.ClientID {
		return nil, fmt.Errorf("%w: client %s is not driver of trip %s", ErrForbidden, req.ClientID, req.ID)
	}

	if resp.Status == model.StatusPublished {
		return &model.MoveTripDraftToPublishModelResponse{
			ID: resp.ID,
		}, nil
	}

	if resp.Status != model.StatusDraft {
		return nil, fmt.Errorf("%w: invalid entity status: expected %s", ErrConflict, model.StatusDraft)
	}

	resp.Status = model.StatusPublished

	updatedTrip, err := repo.UpdateTripTx(ctx, tx, resp)
	if err != nil {
		return nil, fmt.Errorf("error tripRepository.Update: %w", err)
	}

	log.Info("move draft to publish: ", req)
	return updatedTrip.UpdateToPublishModelResponse(), nil
}
