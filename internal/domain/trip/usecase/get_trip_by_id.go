package usecase

import (
	"context"
	"go.opentelemetry.io/otel"
	"log/slog"

	"errors"
	"github.com/jackc/pgx/v5"
	"job4j.ru/share_trip/internal/domain/trip/model"
	"job4j.ru/share_trip/internal/observability/logctx"
	"job4j.ru/share_trip/internal/repository"
)

func (t *TripUseCase) GetTripByID(
	ctx context.Context,
	tx pgx.Tx,
	repo repository.BaseTxTripRepository,
	req model.GetByIDModelRequest,
) (*model.GetTripByIDModelResponse, error) {
	//tracing
	ctx, span := otel.Tracer("TripService").Start(ctx, "TripService.GetTripByID")
	defer span.End()

	//getting custom logger context
	logger := logctx.Logger(ctx).With(
		slog.String("layer", "usecase"),
		slog.String("usecase", "TripUsecase.GetTripByID"),
		slog.String("client_id", req.ID),
	)
	logger.Info("create trip usecase started")

	resp, err := repo.GetByID(ctx, tx, req.ID)
	if err != nil {
		logger.Error(
			"get trip by ID failed",
			slog.Any("error", err),
		)
		if errors.Is(err, repository.ErrTripNotFound) {
			return nil, ErrTripNotFound
		}
		// Если это не ErrEntityNotFound, значит это системный сбой (500 ошибка)
		return nil, err
	}

	logger.Info(
		"get trip by ID usecase success",
		slog.String("trip_id", resp.ID.String()),
	)
	return resp.ToGetByIdModelResponse(), nil
}
