// Orchestration + tx[] wrapper
// service - this layer is responsible for preparing the data, orchestrating the data and passing it further to the usecase
package service

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"job4j.ru/share_trip/internal/domain/trip/model"
	"job4j.ru/share_trip/internal/domain/trip/usecase"
	"job4j.ru/share_trip/internal/observability/metrics"
	"job4j.ru/share_trip/internal/repository"
	"job4j.ru/share_trip/internal/clients/kafka"
)

type Service interface {
	CreateTripWithTx(context.Context, model.CreateTripRequestModel) (*model.CreateTripDraftResponse, error)
	MoveTripDraftToPublish(ctx context.Context, req model.MoveTripDraftToPublishModel) (*model.MoveTripDraftToPublishModelResponse, error)
	GetTripByID(ctx context.Context, req model.GetByIDModelRequest) (*model.GetTripByIDModelResponse, error)
	MoveTripPublishedToStarted(ctx context.Context, req model.MoveTripPublishedToStartedModel) (*model.MoveTripPublishedToStartedModelResponse, error)
}

type TripService struct {
	metrics    *metrics.Metrics
	pool       *pgxpool.Pool
	kafka      kafka.TripEventProducer
	repo       repository.BaseTxTripRepository
	outboxRepo repository.OutboxRepository
	useCase    usecase.BaseTripUseCase
}

func NewTripService(
	m *metrics.Metrics,
	pool *pgxpool.Pool,
	kafka kafka.TripEventProducer,
	r repository.BaseTxTripRepository,
	outbox repository.OutboxRepository,
	uc usecase.BaseTripUseCase,
	) *TripService {
	return &TripService{
		metrics: m,
		pool:    pool,
		kafka:   kafka,
		repo:    r,
		outboxRepo: outbox,
		useCase:    uc,
	}
