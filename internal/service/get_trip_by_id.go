package service

import (
	"context"
	"fmt"
	"github.com/gofiber/fiber/v2/log"
	"github.com/jackc/pgx/v5"
	"job4j.ru/share_trip/internal/domain/trip/model"
)

func (s *TripService) GetTripByID(
	ctx context.Context,
	req model.GetByIdModelRequest,
) (*model.GetTripByIdModelResponse, error) {
	res, err := tx(ctx, s.pool, func(tx pgx.Tx) (*model.GetTripByIdModelResponse, error) {
		resp, err := s.useCase.GetTripById(ctx, tx, s.repo, req)

		if err != nil {
			return nil, fmt.Errorf("err trip UseCaseGetTrip By ID: %w", err)
		}

		return resp, nil

	})

	if err != nil {
		log.Error("error moving trip Draft to Publish: ", err)
		return nil, err
	}

	return res, nil
}
