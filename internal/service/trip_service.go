package service

import (
	"context"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2/log"
	"job4j.ru/share_trip/internal/domain/errs"
	"job4j.ru/share_trip/internal/domain/trip"
	"job4j.ru/share_trip/internal/repository"
)

type BaseTripService interface {
	CreateTrip(ctx context.Context, req trip.CreateTripRequest) (trip.CreateTripResponse, error)
}

type TripService struct {
	repo repository.BaseTripRepository
	vlr  *validator.Validate
}

func NewTripService(repo repository.BaseTripRepository, validator *validator.Validate) *TripService {
	return &TripService{
		repo: repo,
		vlr:  validator,
	}
}

func (s *TripService) CreateTrip(ctx context.Context, tp trip.CreateTripRequest) (trip.CreateTripResponse, error) {
	if err := s.vlr.Struct(&tp); err != nil {
		log.Error(invalidValidateError, err)
		return trip.CreateTripResponse{}, errs.RequestValidationError{Message: err.Error()}
	}

	entity := tp.ToEntity()
	entity.Status = trip.StatusDraft
	entity.Seats = 1

	tr, err := s.repo.CreateTrip(ctx, entity)
	if err != nil {
		log.Debug("error create entity: ", err)
		return trip.CreateTripResponse{}, err
	}

	return tr.ToResponse(), nil
}
