package usecase

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"job4j.ru/share_trip/internal/domain/trip/model"
	"job4j.ru/share_trip/internal/repository"
)

func (t *TripUsecase) GetTripById(
	ctx context.Context,
	tx pgx.Tx,
	repo repository.BaseTxTripRepository,
	req model.GetByIdModelRequest,
) (*model.GetTripByIdModelResponse, error) {
	entity, err := repo.GetByID(ctx, tx, req.ID)

	if err != nil {
		if errors.Is(err, repository.ErrTripNotFound) {
			return nil, ErrTripNotFound
		}
		// Если это не ErrEntityNotFound, значит это системный сбой (500 ошибка)
		return nil, err
	}

	return entity.ToGetByIdModelResponse(), nil
}
