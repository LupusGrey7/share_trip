package usecase

import (
	"context"

	"github.com/jackc/pgx/v5"
	"job4j.ru/share_trip/internal/outbox/domain"
)

// LockPending - lock pending events in one transaction
func (u *OutboxUseCase) LockPending(ctx context.Context, tx pgx.Tx, limit int) ([]*domain.Entity, error) {
	return u.repo.LockPending(ctx, tx, limit)
}
