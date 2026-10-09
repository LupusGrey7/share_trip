package outbox

import (
	"context"
)

// PublishBatch publishes one batch of pending outbox events (see service.PublishPending).
func (p *OutboxPublisher) PublishBatch(ctx context.Context) error {
	return p.outbox.PublishPending(ctx, p.batchSize)
}
