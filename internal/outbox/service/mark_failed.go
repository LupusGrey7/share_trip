package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (o *OutboxService) MarkFailedTx(ctx context.Context, tx pgx.Tx, id uuid.UUID, err error) error {
	return o.useCase.MarkFailedTx(ctx, tx, id, err)
}
