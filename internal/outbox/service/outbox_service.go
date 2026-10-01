package service

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"job4j.ru/share_trip/internal/observability/metrics"
	"job4j.ru/share_trip/internal/outbox/domain"
	"job4j.ru/share_trip/internal/outbox/usecase"
)

// BaseOutboxService — publisher calls this; lock + mark run in one TX inside the service.
type BaseOutboxService interface {
	ProcessPending(ctx context.Context, limit int, handle func(ctx context.Context, e *domain.Entity) error) error
}

// OutboxService orchestrates outbox for the publisher (service → usecase → repo).
type OutboxService struct {
	metrics *metrics.Metrics
	pool    *pgxpool.Pool
	useCase *usecase.OutboxUseCase
}

func NewOutboxService(
	m *metrics.Metrics,
	pool *pgxpool.Pool,
	useCase *usecase.OutboxUseCase,
) *OutboxService {
	return &OutboxService{
		metrics: m,
		pool:    pool,
		useCase: useCase,
	}
}
