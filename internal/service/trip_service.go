// Оркестрация + tx[] wrapper

package service

import (
	"context"
	"job4j.ru/share_trip/internal/domain/trip/model"
	"job4j.ru/share_trip/internal/domain/trip/usecase"

	"github.com/jackc/pgx/v5/pgxpool"
	"job4j.ru/share_trip/internal/repository"
)

type Service interface {
	CreateTripWithTx(context.Context, model.CreateTripRequest) (*model.CreateTripResponse, error)
	MoveTripDraftToPublish(ctx context.Context, req model.MoveTripDraftToPublishModel) (*model.MoveTripDraftToPublishModelResponse, error)
	GetTripByID(ctx context.Context, req model.GetByIdModelRequest) (*model.GetTripByIdModelResponse, error)
}

type TripService struct {
	pool       *pgxpool.Pool
	repo       repository.BaseTxTripRepository
	outboxRepo repository.OutboxRepository
	useCase    usecase.BaseUsecase
}

func NewTripService(
	pool *pgxpool.Pool,
	r repository.BaseTxTripRepository,
	outbox repository.OutboxRepository,
	uc usecase.BaseUsecase,
) *TripService {
	return &TripService{
		pool:       pool,
		repo:       r,
		outboxRepo: outbox,
		useCase:    uc,
	}
}
