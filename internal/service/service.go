package service

import (
	"context"
	"fmt"

	"github.com/gofiber/fiber/v2/log"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"job4j.ru/share_trip/internal/domain/outbox"
	"job4j.ru/share_trip/internal/domain/trip"
	"job4j.ru/share_trip/internal/repository"
	"job4j.ru/share_trip/internal/service/use_case"
)

// Оркестрация + tx[] wrapper

type WriterService interface {
	CreateTripWithTx(context.Context, trip.CreateTripRequest) (*trip.CreateTripResponse, error)
	MoveTripDraftToPublish(ctx context.Context, req trip.MoveTripDraftToPublishModel) (*trip.MoveTripDraftToPublishModelResponse, error)
	GetTripByID(ctx context.Context, tx pgx.Tx, req trip.GetByIdModelRequest) (*trip.GetTripByIdModelResponse, error)
}

type Validator interface {
	Validate(request any) error
}

type TripWriterService struct {
	pool       *pgxpool.Pool
	repo       repository.BaseTxTripRepository
	outboxRepo repository.OutboxRepository
	useCase    use_case.BaseUsecase
}

func NewTripWriterService(
	pool *pgxpool.Pool,
	repo repository.BaseTxTripRepository,
	outbox repository.OutboxRepository,
	uc use_case.BaseUsecase,
) *TripWriterService {
	return &TripWriterService{
		pool:       pool,
		repo:       repo,
		outboxRepo: outbox,
		useCase:    uc,
	}
}

func (s *TripWriterService) CreateTripWithTx(ctx context.Context, req trip.CreateTripRequest) (*trip.CreateTripResponse, error) {
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

func (s *TripWriterService) MoveTripDraftToPublish(
	ctx context.Context,
	req trip.MoveTripDraftToPublishModel,
) (*trip.MoveTripDraftToPublishModelResponse, error) {
	res, err := tx(ctx, s.pool, func(tx pgx.Tx) (*trip.MoveTripDraftToPublishModelResponse, error) {
		resp, err := s.useCase.MoveTripDraftToPublishTx(ctx, tx, s.repo, req)

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

		err = s.outboxRepo.CreateNotificationTripPublishTx(ctx, tx, &event)
		if err != nil {
			return nil, fmt.Errorf("error create Event Notification: %w", err)
		}

		return resp, nil

	})

	if err != nil {
		log.Error("error moving trip Draft to Publish: ", err)
		return nil, err
	}

	return res, nil
}

func (s *TripWriterService) GetTripByID(
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
		log.Error("error moving trip Draft to Publish: ", err)
		return nil, err
	}

	return res, nil
}
