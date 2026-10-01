package usecase

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (u *OutboxUseCase) MarkFailed(ctx context.Context, tx pgx.Tx, ids []uuid.UUID, err error) error {
	return u.repo.MarkFailed(ctx, tx, ids, err)
}
