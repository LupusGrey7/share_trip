package service

import (
	"context"
	"fmt"

	"github.com/gofiber/fiber/v2/log"
	"github.com/jackc/pgx/v5"
	"job4j.ru/share_trip/internal/domain/outbox"
	"job4j.ru/share_trip/internal/domain/trip"
)

func (s *TripService) MoveTripDraftToPublish(
	ctx context.Context,
	req trip.MoveTripDraftToPublishModel,
) (*trip.MoveTripDraftToPublishModelResponse, error) {
	res, err := tx(ctx, s.pool, func(tx pgx.Tx) (*trip.MoveTripDraftToPublishModelResponse, error) {
		resp, err := s.useCase.MoveTripDraftToPublishTx(ctx, tx, s.repo, req)

		if err != nil {
			return nil, fmt.Errorf("err trip UseCase MoveTripDraftToPublishTx: %w", err)
		}

		//outbox
		payload := outbox.PayloadEvent{TripID: resp.ID}
		event := outbox.Entity{
			EventName:   string(outbox.EventPublished),
			AggregateId: resp.ID,
			Payload:     payload,
		}

		err = s.outboxRepo.CreateNotificationTripPublishTx(ctx, tx, &event)
		if err != nil {
			return nil, fmt.Errorf("error create Event Notification: %w", err)
		}

		return resp, nil

	})

	if err != nil {
		log.Error("error moving trip Draft to Publish: ", err)
		return nil, err
	}

	return res, nil
}
