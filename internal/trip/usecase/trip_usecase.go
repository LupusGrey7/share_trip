package usecase

import (
	"context"

	"github.com/jackc/pgx/v5"
	"job4j.ru/share_trip/internal/clients/http/contract"
	contractusecase "job4j.ru/share_trip/internal/clients/http/contract/usecase"
	"job4j.ru/share_trip/internal/storage"
	"job4j.ru/share_trip/internal/trip/domain"
)

type BaseTripUseCase interface {
	CreateTripDraftTx(ctx context.Context, tx pgx.Tx, repo storage.BaseTxTripRepository, req domain.CreateTripInput) (*domain.CreateTripOutput, error)
	MoveTripDraftToPublishTx(ctx context.Context, tx pgx.Tx, repo storage.BaseTxTripRepository, req domain.MoveTripDraftToPublishInput) (*domain.MoveTripDraftToPublishOutput, error)
	GetTripByIDTx(ctx context.Context, tx pgx.Tx, repo storage.BaseTxTripRepository, req *domain.GetByIDInput) (*domain.GetTripByIDOutput, error)
	CheckServiceAllowed(ctx context.Context, companyID string, serviceCode string) (contracts.CheckResult, error)
	MoveTripPublishedToStartedTx(ctx context.Context, tx pgx.Tx, repo storage.BaseTxTripRepository, req domain.MoveTripPublishedToStartedInput) (*domain.MoveTripPublishedToStartedOutput, error)
}
type TripUseCase struct {
	contractUseCase contractusecase.BaseContractUsecase
}

func NewTripUseCase(contractUsecase contractusecase.BaseContractUsecase) *TripUseCase {
	return &TripUseCase{
		contractUseCase: contractUsecase,
	}
}
