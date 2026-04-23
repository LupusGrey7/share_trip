package usecase

import (
	"context"

	"github.com/jackc/pgx/v5"
	"job4j.ru/share_trip/internal/domain/trip"
	"job4j.ru/share_trip/internal/repository"
)

type BaseUsecase interface {
	CreateTrip(ctx context.Context, tx pgx.Tx, repo repository.BaseTxTripRepository, req trip.CreateTripRequest) (*trip.CreateTripResponse, error)
	MoveTripDraftToPublishTx(ctx context.Context, tx pgx.Tx, repo repository.BaseTxTripRepository, req trip.MoveTripDraftToPublishModel) (*trip.MoveTripDraftToPublishModelResponse, error)
	GetTripById(ctx context.Context, tx pgx.Tx, repo repository.BaseTxTripRepository, req trip.GetByIdModelRequest) (*trip.GetTripByIdModelResponse, error)
}

type TripUsecase struct {
}

func NewTripUsecase() *TripUsecase {
	return &TripUsecase{}
}
