package service

import (
	"context"
	"github.com/gofiber/fiber/v2/log"
	"github.com/google/uuid"
	"job4j.ru/share_trip/internal/domain/errs"
	"job4j.ru/share_trip/internal/domain/trip"
)

type RepositoryReader interface {
	GetById(context.Context, string) (trip.Entity, error)
}

type QueryTripService struct {
	repo RepositoryReader
}

func NewQueryTripService(repo RepositoryReader) *QueryTripService {
	return &QueryTripService{
		repo: repo,
	}
}

func (s *QueryTripService) GetById(ctx context.Context, id string) (trip.Response, error) {
	_, err := uuid.Parse(id)
	if err != nil {
		log.Error("uuid parse errors: %v\n", err)
		return trip.Response{}, errs.RequestValidationError{Message: err.Error()}
	}

	tr, err := s.repo.GetById(ctx, id)
	if err != nil {
		log.Debug("error when get by ID: ", err)
		return trip.Response{}, err
	}

	return tr.ToResponse(), nil
}
