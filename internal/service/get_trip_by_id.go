package service

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"job4j.ru/share_trip/internal/domain/trip"
)

func (s *TripService) GetTripByID(
	ctx context.Context,
	req trip.GetByIdModelRequest,
) (*trip.GetTripByIdModelResponse, error) {
	res, err := tx(ctx, s.pool, func(tx pgx.Tx) (*trip.GetTripByIdModelResponse, error) {
		resp, err := s.useCase.GetTripById(ctx, tx, s.repo, req)
		if err != nil {
			return nil, fmt.Errorf("err trip UseCaseGetTrip By ID: %w", err)
		}

		return resp, nil
	})

	if err != nil {
		return nil, err
	}

	return res, nil
}
