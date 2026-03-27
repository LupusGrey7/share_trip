package service

import (
	"context"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2/log"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"job4j.ru/share_trip/internal/domain/errs"
	"job4j.ru/share_trip/internal/domain/trip"
)

const (
	invalidValidateError = "Validation errors: %v\n"
)

type RepositoryWriter interface {
	CreateTrip(context.Context, *trip.Entity) (trip.Entity, error)
	GetForUpdateByID(ctx context.Context, tx pgx.Tx, id uuid.UUID) (trip.Entity, error)
	UpdateTripTx(ctx context.Context, tx pgx.Tx, tp trip.Entity) (trip.Entity, error)
	CreateTripTx(ctx context.Context, tx pgx.Tx, t *trip.Entity) (trip.Entity, error)
	Tx(ctx context.Context, block func(tx pgx.Tx) error) error
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

func (s *CommandTripService) CreateTrip(ctx context.Context, tp trip.CreateTripRequest) (trip.CreateTripResponse, error) {
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

func (s *CommandTripService) CreateTripTx(
	ctx context.Context,
	tp trip.CreateTripRequest,
) (*trip.CreateTripResponse, error) {
	log.Info("create trip tx")

	if err := s.vlr.Struct(&tp); err != nil {
		log.Error(invalidValidateError, err)
		return nil, errs.RequestValidationError{Message: err.Error()}
	}
	var result *trip.CreateTripResponse

	entityToSave := tp.ToEntity()
	entityToSave.Status = trip.StatusDraft
	entityToSave.Seats = 1

	err := s.repo.Tx(ctx, func(tx pgx.Tx) error {
		entity, err := s.repo.CreateTripTx(ctx, tx, entityToSave)
		if err != nil {
			log.Error("error save trip: ", err)
			return err
		}

		result = result.ToResponse(entity)
		return nil
	})
	if err != nil {
		log.Error("error create entity: ", err)
		return nil, err
	}
	return result, nil
}

func (s *CommandTripService) MoveTripDraftToPublish(
	ctx context.Context,
	req trip.MoveTripDraftToPublishModelRequest,
) (*trip.MoveTripDraftToPublishModelResponse, error) {
	log.Info("update trip tx")

	if err := s.vlr.Struct(&req); err != nil {
		log.Error(invalidValidateError, err)
		return nil, errs.RequestValidationError{Message: err.Error()}
	}

	uuID, err := uuid.Parse(req.ID)
	if err != nil {
		log.Error(invalidValidateError, err)
		return nil, errs.JsonParseValidationError{Message: err.Error()}
	}
	log.Info("ID is: ", uuID)

	var result *trip.MoveTripDraftToPublishModelResponse

	err = s.repo.Tx(ctx, func(tx pgx.Tx) error {
		resp, err := s.repo.GetForUpdateByID(ctx, tx, uuID)

		if err != nil {
			return fmt.Errorf("tripRepository.GetForUpdateByID: %w", err)
		}

		if resp.DriverID != req.ClientID {
			return fmt.Errorf("forbidden: client %s is not driver of trip %s", req.ClientID, uuID)
		}

		if resp.Status == trip.StatusPublished {
			result = &trip.MoveTripDraftToPublishModelResponse{
				ID: resp.ID,
			}
		}

		if resp.Status != trip.StatusDraft {
			return fmt.Errorf("invalid entity status: expected %s, got %s", trip.StatusDraft, resp.Status)
		}

		resp.Status = trip.StatusPublished

		updatedTrip, err := s.repo.UpdateTripTx(ctx, tx, resp)
		if err != nil {
			return fmt.Errorf("error tripRepository.Update: %w", err)
		}

		result = result.ToPublishModelResponse(updatedTrip)

		return nil
	})

	if err != nil {
		log.Error("error moving trip Draft to Publish: ", err)
		return nil, err
	}

	return result, nil
}
