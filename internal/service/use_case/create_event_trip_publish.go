package use_case

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"job4j.ru/share_trip/internal/domain/outbox"
	"job4j.ru/share_trip/internal/repository"
)

func (c *OutboxEventUseCase) CreateEventWhenTripToPublish(
	ctx context.Context,
	tx pgx.Tx,
	repo repository.OutboxRepository,
	id uuid.UUID,
) error {
	//outbox
	payload := outbox.PayloadEvent{TripID: id}
	event := outbox.Entity{
		EventName:   string(outbox.EventPublished),
		AggregateId: id,
		Payload:     payload,
	}

	err := repo.CreateNotificationTripPublishTx(ctx, tx, &event)
	if err != nil {
		return fmt.Errorf("error outboxRepository.Create: %w", err)
	}
	return nil
}
