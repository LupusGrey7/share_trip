// Publisher — outbox poller entry (Java @Scheduled analogue).
// Chain: publisher → service → usecase → repo.

package outbox

import (
	"context"
	"time"

	"job4j.ru/share_trip/internal/clients/kafka"
	"job4j.ru/share_trip/internal/observability/metrics"
	"job4j.ru/share_trip/internal/outbox/service"
)

type TripEventPublisher interface {
	Run(ctx context.Context) error
	PublishBatch(ctx context.Context) error
}

type Publisher struct {
	producer  kafka.TripEventProducer
	outbox    *service.OutboxService
	interval  time.Duration
	batchSize int
}

func NewPublisher(
	_ *metrics.Metrics,
	producer kafka.TripEventProducer,
	outbox *service.OutboxService,
	interval time.Duration,
	batchSize int,
) *Publisher {
	if interval <= 0 {
		interval = time.Second
	}
	if batchSize <= 0 {
		batchSize = 50
	}
	return &Publisher{
		producer:  producer,
		outbox:    outbox,
		interval:  interval,
		batchSize: batchSize,
	}
}
