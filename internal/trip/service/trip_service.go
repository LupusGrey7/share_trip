// Orchestration + tx[] wrapper
// service - this layer is responsible for preparing the data, orchestrating the data and passing it further to the usecase
package service

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"job4j.ru/share_trip/internal/clients/kafka"
	"job4j.ru/share_trip/internal/observability/metrics"
	"job4j.ru/share_trip/internal/storage"
	"job4j.ru/share_trip/internal/trip/domain"
	"job4j.ru/share_trip/internal/trip/usecase"
)

type Service interface {
	CreateTripDraft(ctx context.Context, req domain.CreateTripInput) (*domain.CreateTripOutput, error)
	MoveTripDraftToPublish(ctx context.Context, req domain.MoveTripDraftToPublishInput) (*domain.MoveTripDraftToPublishOutput, error)
	GetTripByID(ctx context.Context, input domain.GetByIDInput) (*domain.GetTripByIDOutput, error)
	MoveTripPublishedToStarted(ctx context.Context, req domain.MoveTripPublishedToStartedInput) (*domain.MoveTripPublishedToStartedOutput, error)
}

type TripService struct {
	metrics    *metrics.Metrics
	pool       *pgxpool.Pool
	kafka      kafka.TripEventProducer
	repo       storage.BaseTxTripRepository
	outboxRepo storage.OutboxRepository
	useCase    usecase.BaseTripUseCase
}

func NewTripService(
	m *metrics.Metrics,
	pool *pgxpool.Pool,
	kafkaProducer kafka.TripEventProducer,
	r storage.BaseTxTripRepository,
	outbox storage.OutboxRepository,
	uc usecase.BaseTripUseCase,
) *TripService {
	return &TripService{
		metrics:    m,
		pool:       pool,
		kafka:      kafkaProducer,
		repo:       r,
		outboxRepo: outbox,
		useCase:    uc,
	}
}
