package use_case

import (
	"context"

	"github.com/gofiber/fiber/v2/log"
	"github.com/jackc/pgx/v5"
	"job4j.ru/share_trip/internal/domain/trip"
	"job4j.ru/share_trip/internal/repository"
)

func (t *TripUsecase) GetTripById(
	ctx context.Context,
	tx pgx.Tx,
	repo repository.BaseTxTripRepository,
	req trip.GetByIdModelRequest,
) (*trip.GetTripByIdModelResponse, error) {
	entity, err := repo.GetById(ctx, tx, req.ID)
	if err != nil {
		log.Debug("error create entity: ", err)
		return nil, err
	}

	return entity.ToGetByIdModelResponse(), nil
}
