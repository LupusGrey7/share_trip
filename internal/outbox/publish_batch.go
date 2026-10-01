package outbox

import (
	"context"

	"job4j.ru/share_trip/internal/outbox/domain"
)

// PublishBatch locks a pending batch in one TX, produces to Kafka, then marks sent/failed.
func (p *OutboxPublisher) PublishBatch(ctx context.Context) error {
	return p.outbox.ProcessPending(ctx, p.batchSize, func(ctx context.Context, e *domain.Entity) error {
		return p.producer.PublishTripPublished(ctx, e.ToTripPublished())
	})
}
