// repository/outbox_repo.go

package storage

import (
	"context"
	"fmt"
	"time"

	"job4j.ru/share_trip/internal/outbox/domain"

	"github.com/jackc/pgx/v5"
)



const (
	createEvent = `
insert into outbox_event(event_id, event_name, aggregate_id, payload, created_at)
values($1, $2, $3, $4, $5)`
)

type OutboxRepository interface {
	CreateOutboxEventTripPublishTx(ctx context.Context, tx pgx.Tx, o *domain.Entity) error
}

type OutboxEventRepository struct {
}

func NewOutboxEventRepository() *OutboxEventRepository {
	return &OutboxEventRepository{}
}

func (r *OutboxEventRepository) CreateOutboxEventTripPublishTx(ctx context.Context, tx pgx.Tx, o *domain.Entity) error {
	createdAt := o.CreatedAt
	if createdAt.IsZero() {
		createdAt = time.Now()
	}

	query := createEvent
	args := []interface{}{o.EventID, o.EventName, o.AggregateId, o.Payload, createdAt}

	rows, err := tx.Query(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("err when create outbox: %w", err)
	}
	defer rows.Close()

	return nil
}
