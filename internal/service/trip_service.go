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
	GetTripByID(ctx context.Context, req model.GetByIDModelRequest) (*model.GetTripByIDModelResponse, error)
}

type TripService struct {
	pool       *pgxpool.Pool
	repo       repository.BaseTxTripRepository
	outboxRepo repository.OutboxRepository
	useCase    usecase.BaseUseCase
}

func NewTripService(
	pool *pgxpool.Pool,
	r repository.BaseTxTripRepository,
	outbox repository.OutboxRepository,
	uc usecase.BaseUseCase,
) *TripService {
	return &TripService{
		pool:       pool,
		repo:       r,
		outboxRepo: outbox,
		useCase:    uc,
	}
}
