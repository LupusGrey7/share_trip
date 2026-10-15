package service

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"job4j.ru/share_trip/internal/clients/kafka"
	"job4j.ru/share_trip/internal/observability/metrics"
	"job4j.ru/share_trip/internal/outbox/usecase"
)

// BaseOutboxService — publisher calls this; lock + produce + mark run in one TX inside the service.
type BaseOutboxService interface {
	PublishPending(ctx context.Context, limit int) error
}

// OutboxService orchestrates outbox for the publisher (service → usecase → repo).
type OutboxService struct {
	metrics  *metrics.Metrics
	pool     *pgxpool.Pool
	useCase  *usecase.OutboxUseCase
	producer kafka.TripEventProducer
}

func NewOutboxService(
	m *metrics.Metrics,
	pool *pgxpool.Pool,
	useCase *usecase.OutboxUseCase,
	producer kafka.TripEventProducer,
) *OutboxService {
	return &OutboxService{
		metrics:  m,
		pool:     pool,
		useCase:  useCase,
		producer: producer,
	}
}
