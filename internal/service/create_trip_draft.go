package service

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"job4j.ru/share_trip/internal/domain/trip"
)

func (s *TripService) CreateTripWithTx(ctx context.Context, req trip.CreateTripRequest) (*trip.CreateTripResponse, error) {
	res, err := tx(ctx, s.pool, func(tx pgx.Tx) (*trip.CreateTripResponse, error) {

		resp, err := s.useCase.CreateTrip(ctx, tx, s.repo, req)
		if err != nil {
			return nil, fmt.Errorf("err use_case.CreateTripWithxTx: %w", err)
		}

		return resp, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed in transaction CreateTripWithxTx: %w", err)
	}

	return res, nil
}
