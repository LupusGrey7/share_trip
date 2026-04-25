package service

import (
	"context"
	"fmt"
	model2 "job4j.ru/share_trip/internal/domain/outbox/model"
	"job4j.ru/share_trip/internal/domain/trip/model"

	"github.com/gofiber/fiber/v2/log"
	"github.com/jackc/pgx/v5"
)

func (s *TripService) MoveTripDraftToPublish(
	ctx context.Context,
	req model.MoveTripDraftToPublishModel,
) (*model.MoveTripDraftToPublishModelResponse, error) {
	res, err := tx(ctx, s.pool, func(tx pgx.Tx) (*model.MoveTripDraftToPublishModelResponse, error) {
		resp, err := s.useCase.MoveTripDraftToPublishTx(ctx, tx, s.repo, req)

		if err != nil {
			return nil, err
		}

		//outbox
		payload := model2.PayloadEvent{TripID: resp.ID}
		event := model2.Entity{
			EventName:   string(model2.EventPublished),
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
