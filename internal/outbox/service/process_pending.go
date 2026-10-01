package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"job4j.ru/share_trip/internal/outbox/domain"
)

// ProcessPending - process pending events
func (o *OutboxService) ProcessPending(
	ctx context.Context,
	limit int,
	handle func(ctx context.Context, e *domain.Entity) error, // Lambda function (lambda expression) to handle the event
) error {
	tx, err := o.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin outbox batch tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// lock pending events in one transaction
	events, err := o.LockPending(ctx, tx, limit)
	if err != nil {
		return err
	}
	sentIDs, failedIDs, lastErr := splitByHandle(ctx, events, handle)

	if len(failedIDs) != 0 {
		if err := o.MarkFailed(ctx, tx, failedIDs, fmt.Errorf("handle event: %w", lastErr)); err != nil {
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

// splitByHandle calls handle for each event and splits IDs into sent / failed.
// lastErr is the last handle error (one last_error for the whole failed batch).
func splitByHandle(
	ctx context.Context,
	events []*domain.Entity,
	handle func(ctx context.Context, e *domain.Entity) error,
) (sentIDs, failedIDs []uuid.UUID, lastErr error) {
	sentIDs = make([]uuid.UUID, 0, len(events))
	failedIDs = make([]uuid.UUID, 0, len(events))
	for _, e := range events {
		if err := handle(ctx, e); err != nil {
			failedIDs = append(failedIDs, e.ID)
			lastErr = err
			continue
		}
		sentIDs = append(sentIDs, e.ID)
	}
	return sentIDs, failedIDs, lastErr
}
