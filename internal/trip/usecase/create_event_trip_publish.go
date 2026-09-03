package usecase

import (
	"context"
	"fmt"

	"job4j.ru/share_trip/internal/outbox/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"job4j.ru/share_trip/internal/storage"
)

func (c *OutboxEventUseCase) CreateEventWhenTripToPublish(
	ctx context.Context,
	tx pgx.Tx,
	repo storage.OutboxRepository,
	id uuid.UUID,
) error {
	//outbox
	payload := domain.PayloadEvent{TripID: id}
	event := domain.Entity{
		EventName:   string(domain.EventPublished),
		AggregateId: id,
		Payload:     payload,
	}

	err := repo.CreateNotificationTripPublishTx(ctx, tx, &event)
	if err != nil {
		return fmt.Errorf("error outboxRepository.Create: %w", err)
	}
	return nil
}
