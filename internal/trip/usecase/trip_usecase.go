package usecase

import (
	"context"

	"github.com/jackc/pgx/v5"
	"job4j.ru/share_trip/internal/client/contracts"
	contractusecase "job4j.ru/share_trip/internal/client/contracts/usecase"
	"job4j.ru/share_trip/internal/storage"
	"job4j.ru/share_trip/internal/trip/domain"
)

type BaseTripUseCase interface {
	CreateTripDraftTx(ctx context.Context, tx pgx.Tx, repo storage.BaseTxTripRepository, req domain.CreateTripRequestModel) (*domain.CreateTripDraftResponse, error)
	MoveTripDraftToPublishTx(ctx context.Context, tx pgx.Tx, repo storage.BaseTxTripRepository, req domain.MoveTripDraftToPublishModel) (*domain.MoveTripDraftToPublishModelResponse, error)
	GetTripByIDTx(ctx context.Context, tx pgx.Tx, repo storage.BaseTxTripRepository, req domain.GetByIDInput) (*domain.GetTripByIDOutput, error)
	CheckServiceAllowed(ctx context.Context, companyID string, serviceCode string) (contracts.CheckResult, error)
	MoveTripPublishedToStartedTx(ctx context.Context, tx pgx.Tx, repo storage.BaseTxTripRepository, req domain.MoveTripPublishedToStartedInput) (*domain.MoveTripPublishedToStartedModelResponse, error)
}

type TripUseCase struct {
	contractUsecase contractusecase.BaseContractUsecase
}

func NewTripUseCase(contractUsecase contractusecase.BaseContractUsecase) *TripUseCase {
	return &TripUseCase{
		contractUsecase: contractUsecase,
	}
}
