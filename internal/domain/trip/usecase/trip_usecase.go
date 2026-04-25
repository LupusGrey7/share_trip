package usecase

import (
	"context"
	"job4j.ru/share_trip/internal/domain/trip/model"

	"github.com/jackc/pgx/v5"
	"job4j.ru/share_trip/internal/repository"
)

type BaseUsecase interface {
	CreateTrip(ctx context.Context, tx pgx.Tx, repo repository.BaseTxTripRepository, req model.CreateTripRequest) (*model.CreateTripResponse, error)
	MoveTripDraftToPublishTx(ctx context.Context, tx pgx.Tx, repo repository.BaseTxTripRepository, req model.MoveTripDraftToPublishModel) (*model.MoveTripDraftToPublishModelResponse, error)
	GetTripById(ctx context.Context, tx pgx.Tx, repo repository.BaseTxTripRepository, req model.GetByIdModelRequest) (*model.GetTripByIdModelResponse, error)
}

type TripUsecase struct {
}

func NewTripUsecase() *TripUsecase {
	return &TripUsecase{}
}
