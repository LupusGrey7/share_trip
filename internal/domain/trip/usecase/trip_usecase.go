package usecase

import (
	"context"

	"github.com/jackc/pgx/v5"
	contractusecase "job4j.ru/share_trip/internal/client/contracts/usecase"
	"job4j.ru/share_trip/internal/domain/trip/model"
	"job4j.ru/share_trip/internal/repository"
)

type BaseTripUseCase interface {
	CreateTripDraftTx(ctx context.Context, tx pgx.Tx, repo repository.BaseTxTripRepository, req model.CreateTripRequestModel) (*model.CreateTripDraftResponse, error)
	MoveTripDraftToPublishTx(ctx context.Context, tx pgx.Tx, repo repository.BaseTxTripRepository, req model.MoveTripDraftToPublishModel) (*model.MoveTripDraftToPublishModelResponse, error)
	GetTripByIDTx(ctx context.Context, tx pgx.Tx, repo repository.BaseTxTripRepository, req model.GetByIDModelRequest) (*model.GetTripByIDModelResponse, error)
	//Добавьте use case StartTrip.
	MoveTripPublishedToStartTx(ctx context.Context, tx pgx.Tx, repo repository.BaseTxTripRepository, req model.MoveTripPublishedToStartModel) (*model.MoveTripPublishedToStartModelResponse, error)
}
type TripUseCase struct {
	contractUsecase contractusecase.BaseContractUsecase
}

func NewTripUseCase(contractUsecase contractusecase.BaseContractUsecase) *TripUseCase {
	return &TripUseCase{
		contractUsecase: contractUsecase,
	}
}
