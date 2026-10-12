package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

// PublishPending locks a pending batch in one TX, produces each event to Kafka, then marks sent/failed.
func (o *OutboxService) PublishPending(ctx context.Context, limit int) error {
	tx, err := o.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin outbox batch tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	events, err := o.LockPending(ctx, tx, limit)
	if err != nil {
		return err
	}

	sentIDs := make([]uuid.UUID, 0, len(events))
	failedIDs := make([]uuid.UUID, 0, len(events))
	var lastErr error

	for _, e := range events {
		if err := o.producer.PublishTripPublished(ctx, e.ToTripPublished()); err != nil {
			failedIDs = append(failedIDs, e.ID)
			lastErr = err
			continue
		}
		sentIDs = append(sentIDs, e.ID)
	}

	if len(failedIDs) != 0 {
		if err := o.MarkFailed(ctx, tx, failedIDs, fmt.Errorf("publish event: %w", lastErr)); err != nil {
			return err
		}
	}

	if len(sentIDs) != 0 {
		if err := o.MarkSent(ctx, tx, sentIDs); err != nil {
			return err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit outbox batch tx: %w", err)
	}
	return nil
}
