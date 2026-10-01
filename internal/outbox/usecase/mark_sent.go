package usecase

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (u *OutboxUseCase) MarkSent(ctx context.Context, tx pgx.Tx, ids []uuid.UUID) error {
	return u.repo.MarkSent(ctx, tx, ids)
}
