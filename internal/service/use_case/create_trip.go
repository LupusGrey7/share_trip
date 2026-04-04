package use_case

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"job4j.ru/share_trip/internal/domain/trip"
	"job4j.ru/share_trip/internal/repository"
)

func (t *TripUsecase) CreateTrip(
	ctx context.Context,
	tx pgx.Tx,
	repo repository.BaseTxTripRepository,
	req trip.CreateTripRequest,
) (*trip.CreateTripResponse, error) {
	entity := req.ToEntity()
	entity.Status = trip.StatusDraft
	entity.Seats = 1

	resp, err := repo.CreateTripTx(ctx, tx, entity)
	if err != nil {
		return nil, fmt.Errorf("err use_case.CreateTrip: %w", err)
	}

	return resp.ToCreateResponse(), nil
}
