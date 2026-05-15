package usecase

import (
	"context"
	"job4j.ru/share_trip/internal/domain/trip/model"

	"github.com/jackc/pgx/v5"
	"job4j.ru/share_trip/internal/repository"
)

type BaseUseCase interface {
	CreateTrip(ctx context.Context, tx pgx.Tx, repo repository.BaseTxTripRepository, req model.CreateTripRequestModel) (*model.CreateTripResponse, error)
	MoveTripDraftToPublishTx(ctx context.Context, tx pgx.Tx, repo repository.BaseTxTripRepository, req model.MoveTripDraftToPublishModel) (*model.MoveTripDraftToPublishModelResponse, error)
	GetTripByID(ctx context.Context, tx pgx.Tx, repo repository.BaseTxTripRepository, req model.GetByIDModelRequest) (*model.GetTripByIDModelResponse, error)
}

type TripUseCase struct {
}

func NewTripUseCase() *TripUseCase {
	return &TripUseCase{}
}
