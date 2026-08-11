// Orchestration + tx[] wrapper

package service

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"job4j.ru/share_trip/internal/domain/trip/model"
	"job4j.ru/share_trip/internal/domain/trip/usecase"
	"job4j.ru/share_trip/internal/observability/metrics"
	"job4j.ru/share_trip/internal/repository"
)

type Service interface {
	CreateTripWithTx(context.Context, model.CreateTripRequestModel) (*model.CreateTripDraftResponse, error)
	MoveTripDraftToPublish(ctx context.Context, req model.MoveTripDraftToPublishModel) (*model.MoveTripDraftToPublishModelResponse, error)
	GetTripByID(ctx context.Context, req model.GetByIDModelRequest) (*model.GetTripByIDModelResponse, error)
}

type TripService struct {
	metrics    *metrics.Metrics
	pool       *pgxpool.Pool
	repo       repository.BaseTxTripRepository
	outboxRepo repository.OutboxRepository
	useCase    usecase.BaseUseCase
}

func NewTripService(
	m *metrics.Metrics,
	pool *pgxpool.Pool,
	r repository.BaseTxTripRepository,
	outbox repository.OutboxRepository,
	uc usecase.BaseUseCase,
) *TripService {
	return &TripService{
		metrics:    m,
		pool:       pool,
		repo:       r,
		outboxRepo: outbox,
		useCase:    uc,
	}
}
