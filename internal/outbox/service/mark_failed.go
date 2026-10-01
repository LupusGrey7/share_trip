package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (o *OutboxService) MarkFailed(ctx context.Context, tx pgx.Tx, ids []uuid.UUID, err error) error {
	return o.useCase.MarkFailed(ctx, tx, ids, err)
}
