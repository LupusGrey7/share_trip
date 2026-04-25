package service

import (
	"context"
	"github.com/gofiber/fiber/v2/log"
	"github.com/jackc/pgx/v5"
	"job4j.ru/share_trip/internal/domain/trip"
)

func (s *TripService) GetTripByID(
	ctx context.Context,
	req trip.GetByIDModelRequest,
) (*trip.GetTripByIDModelResponse, error) {
	res, err := tx(ctx, s.pool, func(tx pgx.Tx) (*trip.GetTripByIDModelResponse, error) {
		resp, err := s.useCase.GetTripById(ctx, tx, s.repo, req)

		if err != nil {
			return nil, err
		}

		return resp, nil
	})

	if err != nil {
		log.Error("error Get By ID: ", err)
		return nil, err
	}

	return res, nil
}
