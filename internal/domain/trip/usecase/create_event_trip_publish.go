package usecase

import (
	"context"
	"fmt"
	"job4j.ru/share_trip/internal/domain/outbox/model"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"job4j.ru/share_trip/internal/repository"
)

func (c *OutboxEventUseCase) CreateEventWhenTripToPublish(
	ctx context.Context,
	tx pgx.Tx,
	repo repository.OutboxRepository,
	id uuid.UUID,
) error {
	//outbox
	payload := model.PayloadEvent{TripID: id}
	event := model.Entity{
		EventName:   string(model.EventPublished),
		AggregateId: id,
		Payload:     payload,
	}

	err := repo.CreateNotificationTripPublishTx(ctx, tx, &event)
	if err != nil {
		return fmt.Errorf("error outboxRepository.Create: %w", err)
	}
	return nil
}
