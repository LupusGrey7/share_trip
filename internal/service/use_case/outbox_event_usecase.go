package use_case

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"job4j.ru/share_trip/internal/repository"
)

type BaseOutboxEventUseCase interface {
	CreateEventWhenTripToPublish(ctx context.Context, tx pgx.Tx, repo repository.OutboxRepository, id uuid.UUID) error
}

type OutboxEventUseCase struct {
}

func NewOutboxEventUseCase() *OutboxEventUseCase {
	return &OutboxEventUseCase{}
}
