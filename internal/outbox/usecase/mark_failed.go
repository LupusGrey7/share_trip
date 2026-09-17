package usecase

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (u *OutboxUseCase) MarkFailedTx(ctx context.Context, tx pgx.Tx, id uuid.UUID, err error) error {
	return u.repo.MarkFailedTx(ctx, tx, id, err)
}
