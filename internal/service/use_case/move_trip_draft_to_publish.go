package use_case

import (
	"context"
	"fmt"

	"github.com/gofiber/fiber/v2/log"
	"github.com/jackc/pgx/v5"
	"job4j.ru/share_trip/internal/domain/trip"
	"job4j.ru/share_trip/internal/repository"
)

func (t *TripUsecase) MoveTripDraftToPublishTx(
	ctx context.Context,
	tx pgx.Tx,
	repo repository.BaseTxTripRepository,
	req trip.MoveTripDraftToPublishModel,
) (*trip.MoveTripDraftToPublishModelResponse, error) {
	log.Info("update trip tx", req)

	resp, err := repo.GetForUpdateByIDTx(ctx, tx, req.ID)

	if err != nil {
		return nil, fmt.Errorf("tripRepository.GetForUpdateByID: %w", err)
	}

	if resp.DriverID != req.ClientID {
		return nil, fmt.Errorf("forbidden: client %s is not driver of trip %s", req.ClientID, req.ID)
	}

	if resp.Status == trip.StatusPublished {
		return &trip.MoveTripDraftToPublishModelResponse{
			ID: resp.ID,
		}, nil
	}

	if resp.Status != trip.StatusDraft {
		return nil, fmt.Errorf("invalid entity status: expected %s, got %s", trip.StatusDraft, resp.Status)
	}

	resp.Status = trip.StatusPublished

	updatedTrip, err := repo.UpdateTripTx(ctx, tx, resp)
	if err != nil {
		return nil, fmt.Errorf("error tripRepository.Update: %w", err)
	}

	return updatedTrip.UpdateToPublishModelResponse(), nil
}
