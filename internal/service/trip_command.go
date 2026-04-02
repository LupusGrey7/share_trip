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
	"job4j.ru/share_trip/internal/domain/outbox"
	"job4j.ru/share_trip/internal/domain/trip"
	"job4j.ru/share_trip/internal/repository"
	"job4j.ru/share_trip/internal/service/use_case"
)

// Оркестрация + tx[] wrapper

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
	pool       *pgxpool.Pool
	repo       repository.BaseTxTripRepository
	outboxRepo repository.OutboxRepository
	vlr        *validator.Validate
}

func NewCommandTripService(pool *pgxpool.Pool, repo repository.BaseTxTripRepository, outboxRepo repository.OutboxRepository, validator *validator.Validate) *CommandTripService {
	return &CommandTripService{
		pool:       pool,
		repo:       repo,
		outboxRepo: outboxRepo,
		vlr:        validator,
	}
}

func (s *CommandTripService) CreateTripWithTx(ctx context.Context, req trip.CreateTripRequest) (*trip.CreateTripResponse, error) {
	if err := s.vlr.Struct(&req); err != nil {
		log.Error(invalidValidateError, err)
		return &trip.CreateTripResponse{}, errs.RequestValidationError{Message: err.Error()}
	}

	res, err := tx(ctx, s.pool, func(tx pgx.Tx) (*trip.CreateTripResponse, error) {

		resp, err := use_case.CreateTrip(ctx, tx, s.repo, req)
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

func (s *CommandTripService) MoveTripDraftToPublish(
	ctx context.Context,
	req trip.MoveTripDraftToPublishModelRequest,
) (*trip.MoveTripDraftToPublishModelResponse, error) {
	log.Info("update trip tx")
	var req1 trip.MoveTripDraftToPublishModel

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
	req1.ID = uuID
	req1.ClientID = req.ClientID

	res, err := tx(ctx, s.pool, func(tx pgx.Tx) (*trip.MoveTripDraftToPublishModelResponse, error) {
		resp, err := use_case.MoveTripDraftToPublishTx(ctx, tx, s.repo, req1)

		if err != nil {
			return nil, fmt.Errorf("err trip UseCase MoveTripDraftToPublishTx: %w", err)
		}
		//outbox
		payload := outbox.PayloadEvent{TripID: resp.ID}
		event := outbox.Entity{
			EventName:   string(outbox.EventPublished),
			AggregateId: resp.ID,
			Payload:     payload,
		}

		err = s.outboxRepo.CreateEventTx(ctx, tx, &event)
		if err != nil {
			return nil, fmt.Errorf("error outboxRepository.Create: %w", err)
		}

		return resp, nil
	})

	if err != nil {
		log.Error("error moving trip Draft to Publish: ", err)
		return nil, err
	}

	return res, nil
}
