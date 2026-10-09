// Publisher — outbox poller entry (Java @Scheduled analogue).
// Chain: publisher → service → usecase → repo.

package outbox

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"job4j.ru/share_trip/internal/clients/kafka"
	"job4j.ru/share_trip/internal/observability/metrics"
	"job4j.ru/share_trip/internal/outbox/service"
	"job4j.ru/share_trip/internal/outbox/usecase"
	"job4j.ru/share_trip/internal/storage"
)

const (
	defaultInterval  = time.Second
	defaultBatchSize = 50
)

type TripEventPublisher interface {
	Run(ctx context.Context) error
	PublishBatch(ctx context.Context) error
}

type OutboxPublisher struct {
	outbox    *service.OutboxService
	interval  time.Duration
	batchSize int
}

func NewOutboxPublisher(
	m *metrics.Metrics,
	producer kafka.TripEventProducer,
	pool *pgxpool.Pool,
	interval time.Duration,
	batchSize int,
) *OutboxPublisher {
	if interval <= 0 {
		interval = defaultInterval
	}
	if batchSize <= 0 {
		batchSize = defaultBatchSize
	}

	outboxRepo := storage.NewOutboxEventRepository(m)
	outboxUC := usecase.NewOutboxUseCase(outboxRepo)
	outboxSvc := service.NewOutboxService(m, pool, outboxUC, producer)

	return &OutboxPublisher{
		outbox:    outboxSvc,
		interval:  interval,
		batchSize: batchSize,
	}
}
