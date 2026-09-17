package outbox

import (
	"context"
	"fmt"
)

// PublishBatch locks a pending batch in one TX, produces to Kafka, then marks sent/failed.
// Produce runs while rows are locked (SKIP LOCKED); COMMIT releases them.
func (p *Publisher) PublishBatch(ctx context.Context) error {
	tx, err := p.outbox.Pool().Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin outbox batch tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	events, err := p.outbox.LockPendingTx(ctx, tx, p.batchSize)
	if err != nil {
		return err
	}
	if len(events) == 0 {
		return tx.Commit(ctx)
	}

	for _, event := range events {
		evt := event.ToTripPublished()
		if err := p.producer.PublishTripPublished(ctx, evt); err != nil {
			if markErr := p.outbox.MarkFailedTx(ctx, tx, event.ID, err); markErr != nil {
				return markErr
			}
			continue
		}

		if err := p.outbox.MarkSentTx(ctx, tx, event.ID); err != nil {
			return err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit outbox batch tx: %w", err)
	}
	return nil
}
