package usecase

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"job4j.ru/share_trip/internal/outbox/domain"
	"job4j.ru/share_trip/internal/storage"
)

// BaseOutboxUseCase — DB operations for outbox (calls repo).
type BaseOutboxUseCase interface {
	LockPending(ctx context.Context, tx pgx.Tx, limit int) ([]*domain.Entity, error)
	MarkSent(ctx context.Context, tx pgx.Tx, ids []uuid.UUID) error
	MarkFailed(ctx context.Context, tx pgx.Tx, ids []uuid.UUID, err error) error
}

type OutboxUseCase struct {
	repo storage.OutboxRepository
}

func NewOutboxUseCase(repo storage.OutboxRepository) *OutboxUseCase {
	return &OutboxUseCase{repo: repo}
}
