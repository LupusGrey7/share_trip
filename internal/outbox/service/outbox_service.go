package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"job4j.ru/share_trip/internal/observability/metrics"
	"job4j.ru/share_trip/internal/outbox/domain"
	"job4j.ru/share_trip/internal/outbox/usecase"
)

// BaseOutboxService — publisher calls this; Lock/Mark must share one TX.
type BaseOutboxService interface {
	LockPendingTx(ctx context.Context, tx pgx.Tx, limit int) ([]*domain.Entity, error)
	MarkSentTx(ctx context.Context, tx pgx.Tx, id uuid.UUID) error
	MarkFailedTx(ctx context.Context, tx pgx.Tx, id uuid.UUID, err error) error
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

func (o *OutboxService) Pool() *pgxpool.Pool {
	return o.pool
}
