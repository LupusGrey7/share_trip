package usecase

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (u *OutboxUseCase) MarkSentTx(ctx context.Context, tx pgx.Tx, id uuid.UUID) error {
	return u.repo.MarkSentTx(ctx, tx, id)
}
