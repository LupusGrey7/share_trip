package service

import (
	"context"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2/log"
	"job4j.ru/share_trip/internal/domain/errs"
	"job4j.ru/share_trip/internal/domain/trip"
)

type RepositoryWriter interface {
	CreateTrip(context.Context, *trip.Entity) (trip.Entity, error)
}

type Validator interface {
	Validate(request any) error
}

type CommandTripService struct {
	repo RepositoryWriter
	vlr  *validator.Validate
}

func NewCommandTripService(repo RepositoryWriter, validator *validator.Validate) *CommandTripService {
	return &CommandTripService{
		repo: repo,
		vlr:  validator,
	}
}

func (s *CommandTripService) CreateTrip(ctx context.Context, tp trip.CreateTripCommand) (trip.Response, error) {
	if err := s.vlr.Struct(&tp); err != nil {
		log.Error("Validation errors: %v\n", err)
		return trip.Response{}, errs.RequestValidationError{Message: err.Error()}
	}

	entity := tp.ToEntity()
	entity.Status = trip.Draft

	tr, err := s.repo.CreateTrip(ctx, entity)
	if err != nil {
		log.Debug("error create entity: ", err)
		return trip.Response{}, err
	}

	return tr.ToResponse(), nil
}
