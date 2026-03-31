package service

import (
	"context"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2/log"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"job4j.ru/share_trip/internal/domain/errs"
	"job4j.ru/share_trip/internal/domain/trip"
	"job4j.ru/share_trip/internal/repository"
)

const (
	invalidValidateError = "Validation errors: %v\n"
)

type CommandTripWriterService interface {
	CreateTripWithTx(context.Context, trip.CreateTripRequest) (*trip.CreateTripResponse, error)
	MoveTripDraftToPublish(ctx context.Context, req trip.MoveTripDraftToPublishModelRequest) (*trip.MoveTripDraftToPublishModelResponse, error)
}

type Validator interface {
	Validate(request any) error
}

type CommandTripService struct {
	pool *pgxpool.Pool
	repo repository.BaseTxTripRepository
	vlr  *validator.Validate
}

func NewCommandTripService(pool *pgxpool.Pool, repo repository.BaseTxTripRepository, validator *validator.Validate) *CommandTripService {
	return &CommandTripService{
		pool: pool,
		repo: repo,
		vlr:  validator,
	}
}

func (s *CommandTripService) CreateTripWithTx(ctx context.Context, req trip.CreateTripRequest) (*trip.CreateTripResponse, error) {
	if err := s.vlr.Struct(&req); err != nil {
		log.Error(invalidValidateError, err)
		return &trip.CreateTripResponse{}, errs.RequestValidationError{Message: err.Error()}
	}

	res, err := tx(ctx, s.pool, func(tx pgx.Tx) (*trip.CreateTripResponse, error) {
		entity := req.ToEntity()
		entity.Status = trip.StatusDraft
		entity.Seats = 1

		resp, err := s.repo.CreateTripTx(ctx, tx, entity)
		if err != nil {
			return nil, fmt.Errorf("usecase.CreateTrip: %w", err)
		}

		return resp.ToCreateResponse(), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed in transaction: %w", err)
	}

	return res, nil
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
	log.Info("trip ID is: ", uuID)

	res, err := tx(ctx, s.pool, func(tx pgx.Tx) (*trip.MoveTripDraftToPublishModelResponse, error) {
		resp, err := s.repo.GetForUpdateByIDTx(ctx, tx, uuID)

		if err != nil {
			return nil, fmt.Errorf("tripRepository.GetForUpdateByID: %w", err)
		}

		if resp.DriverID != req.ClientID {
			return nil, fmt.Errorf("forbidden: client %s is not driver of trip %s", req.ClientID, uuID)
		}

		if resp.Status == trip.StatusPublished {
			return &trip.MoveTripDraftToPublishModelResponse{
				ID: resp.ID,
			}, nil
		}

		if resp.Status != trip.StatusDraft {
			return nil, fmt.Errorf("invalid entity status: expected %s, got %s", trip.StatusDraft, resp.Status)
		}

		resp.Status = trip.StatusPublished

		updatedTrip, err := s.repo.UpdateTripTx(ctx, tx, resp)
		if err != nil {
			return nil, fmt.Errorf("error tripRepository.Update: %w", err)
		}

		return updatedTrip.UpdateToPublishModelResponse(), nil
	})

	if err != nil {
		log.Error("error moving trip Draft to Publish: ", err)
		return nil, err
	}

	return res, nil
}
