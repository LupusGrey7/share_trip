package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (o *OutboxService) MarkSent(ctx context.Context, tx pgx.Tx, ids []uuid.UUID) error {
	return o.useCase.MarkSent(ctx, tx, ids)
}
