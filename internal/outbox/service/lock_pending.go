package service

import (
	"context"

	"github.com/jackc/pgx/v5"
	"job4j.ru/share_trip/internal/outbox/domain"
)

// LockPending - lock pending events in one transaction
func (o *OutboxService) LockPending(ctx context.Context, tx pgx.Tx, limit int) ([]*domain.Entity, error) {
	return o.useCase.LockPending(ctx, tx, limit)
}
