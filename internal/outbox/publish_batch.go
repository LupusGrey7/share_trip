package outbox

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"job4j.ru/share_trip/internal/clients/kafka"
	"job4j.ru/share_trip/internal/outbox/domain"
)

// deliverBatch produces each event and marks sent/failed. Used by PublishBatch (after lock)
// and covered by unit tests without Postgres.
func deliverBatch(
	ctx context.Context,
	events []*domain.Entity,
	producer kafka.TripEventProducer,
	markSent func(context.Context, uuid.UUID) error,
	markFailed func(context.Context, uuid.UUID, error) error,
) error {
	for _, event := range events {
		evt := event.ToTripPublished()
		if err := producer.PublishTripPublished(ctx, evt); err != nil {
			if markErr := markFailed(ctx, event.ID, err); markErr != nil {
				return markErr
			}
			continue
		}
		if err := markSent(ctx, event.ID); err != nil {
			return err
		}
	}
	return nil
}

// PublishBatch locks a pending batch in one TX, produces to Kafka, then marks sent/failed.
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

	if err := deliverBatch(
		ctx,
		events,
		p.producer,
		func(ctx context.Context, id uuid.UUID) error {
			return p.outbox.MarkSentTx(ctx, tx, id)
		},
		func(ctx context.Context, id uuid.UUID, produceErr error) error {
			return p.outbox.MarkFailedTx(ctx, tx, id, produceErr)
		},
	); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit outbox batch tx: %w", err)
	}
	return nil
}
