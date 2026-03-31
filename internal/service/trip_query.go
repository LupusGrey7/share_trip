package service

import (
	"context"
	"github.com/gofiber/fiber/v2/log"
	"github.com/google/uuid"
	"job4j.ru/share_trip/internal/domain/errs"
	"job4j.ru/share_trip/internal/domain/trip"
	"job4j.ru/share_trip/internal/repository"
)

type QueryTripReaderService interface {
	GetById(ctx context.Context, id string) (trip.CreateTripResponse, error)
}

type QueryTripService struct {
	repo repository.BaseTripRepository
}

func NewQueryTripService(repo repository.BaseTripRepository) *QueryTripService {
	return &QueryTripService{
		repo: repo,
	}
}

func (s *QueryTripService) GetById(ctx context.Context, id string) (trip.CreateTripResponse, error) {
	_, err := uuid.Parse(id)
	if err != nil {
		log.Error("uuid parse errors: %v\n", err)
		return trip.CreateTripResponse{}, errs.RequestValidationError{Message: err.Error()}
	}

	tr, err := s.repo.GetById(ctx, id)
	if err != nil {
		log.Debug("error when get by ID: ", err)
		return trip.CreateTripResponse{}, err
	}

	return tr.ToResponse(), nil
}
