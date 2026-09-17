package service

import (
	"context"

	"github.com/jackc/pgx/v5"
	"job4j.ru/share_trip/internal/outbox/domain"
)

func (o *OutboxService) LockPendingTx(
	ctx context.Context,
	tx pgx.Tx,
	limit int,
) ([]*domain.Entity, error) {
	return o.useCase.LockPendingTx(ctx, tx, limit)
}
