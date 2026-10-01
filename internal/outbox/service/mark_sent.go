package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (o *OutboxService) MarkSentTx(ctx context.Context, tx pgx.Tx, id uuid.UUID) error {
	return o.useCase.MarkSentTx(ctx, tx, id)
}
