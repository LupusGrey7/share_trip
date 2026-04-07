package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"job4j.ru/share_trip/internal/domain/outbox"
)

// repository/outbox_repo.go

const (
	createEvent = `
insert into outbox_event(event_name, aggregate_id, payload, created_at)
values($1, $2, $3, $4)`
)

type OutboxRepository interface {
	CreateNotificationTripPublishTx(ctx context.Context, tx pgx.Tx, o *outbox.Entity) error
}

type OutboxEventRepository struct {
}

func NewOutboxEventRepository() *OutboxEventRepository {
	return &OutboxEventRepository{}
}

func (r *OutboxEventRepository) CreateNotificationTripPublishTx(ctx context.Context, tx pgx.Tx, o *outbox.Entity) error {

	query := createEvent
	args := []interface{}{o.EventName, o.AggregateId, o.Payload, time.Now()}

	rows, err := tx.Query(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("err when create outbox: %w", err)
	}
	defer rows.Close() // обработать rows

	return nil
}
