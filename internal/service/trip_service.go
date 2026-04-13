package service

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"job4j.ru/share_trip/internal/domain/trip"
	"job4j.ru/share_trip/internal/repository"
	"job4j.ru/share_trip/internal/service/use_case"
)

// Оркестрация + tx[] wrapper

type Service interface {
	CreateTripWithTx(context.Context, trip.CreateTripRequest) (*trip.CreateTripResponse, error)
	MoveTripDraftToPublish(ctx context.Context, req trip.MoveTripDraftToPublishModel) (*trip.MoveTripDraftToPublishModelResponse, error)
	GetTripByID(ctx context.Context, req trip.GetByIdModelRequest) (*trip.GetTripByIdModelResponse, error)
}

type TripService struct {
	pool       *pgxpool.Pool
	repo       repository.BaseTxTripRepository
	outboxRepo repository.OutboxRepository
	useCase    use_case.BaseUsecase
}

func NewTripService(
	pool *pgxpool.Pool,
	r repository.BaseTxTripRepository,
	outbox repository.OutboxRepository,
	uc use_case.BaseUsecase,
) *TripService {
	return &TripService{
		pool:       pool,
		repo:       r,
		outboxRepo: outbox,
		useCase:    uc,
	}
}
