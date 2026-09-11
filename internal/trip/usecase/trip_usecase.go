package usecase

import (
	"context"

	"github.com/jackc/pgx/v5"
	contracts "job4j.ru/share_trip/internal/clients/http/contract"
	contractusecase "job4j.ru/share_trip/internal/clients/http/contract/usecase"
	"job4j.ru/share_trip/internal/storage"
	"job4j.ru/share_trip/internal/trip/domain"
)

type BaseTripUseCase interface {
	CreateTripDraft(ctx context.Context, tx pgx.Tx, repo storage.BaseTxTripRepository, req domain.CreateTripInput) (*domain.CreateTripOutput, error)
	MoveTripFromDraftToPublish(ctx context.Context, tx pgx.Tx, repo storage.BaseTxTripRepository, req domain.MoveTripFromDraftToPublishInput) (*domain.MoveTripFromDraftToPublishOutput, error)
	GetTripByID(ctx context.Context, tx pgx.Tx, repo storage.BaseTxTripRepository, req *domain.GetByIDInput) (*domain.GetTripByIDOutput, error)
	CheckServiceIsAllowed(ctx context.Context, companyID string, serviceCode string) (contracts.CheckResult, error)
	MoveTripFromPublishedToStarted(ctx context.Context, tx pgx.Tx, repo storage.BaseTxTripRepository, req domain.MoveTripFromPublishedToStartedInput) (*domain.MoveTripFromPublishedToStartedOutput, error)
}

type TripUseCase struct {
	contractUseCase contractusecase.BaseContractUsecase
}

func NewTripUseCase(contractUsecase contractusecase.BaseContractUsecase) *TripUseCase {
	return &TripUseCase{
		contractUseCase: contractUsecase,
	}
}
