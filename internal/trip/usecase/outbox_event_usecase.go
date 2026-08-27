package usecase

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"job4j.ru/share_trip/internal/storage"
)

type BaseOutboxEventUseCase interface {
	CreateEventWhenTripToPublish(ctx context.Context, tx pgx.Tx, repo storage.OutboxRepository, id uuid.UUID) error
}

type OutboxEventUseCase struct {
}

func NewOutboxEventUseCase() *OutboxEventUseCase {
	return &OutboxEventUseCase{}
}
