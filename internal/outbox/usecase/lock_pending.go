package usecase

import (
	"context"

	"github.com/jackc/pgx/v5"
	"job4j.ru/share_trip/internal/outbox/domain"
)

func (u *OutboxUseCase) LockPendingTx(
	ctx context.Context,
	tx pgx.Tx,
	limit int,
) ([]*domain.Entity, error) {
	return u.repo.LockPendingTx(ctx, tx, limit)
}
