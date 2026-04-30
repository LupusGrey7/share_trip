package usecase

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"job4j.ru/share_trip/internal/domain/trip/model"
	"job4j.ru/share_trip/internal/observability/logctx"
	"job4j.ru/share_trip/internal/repository"
)

func (t *TripUsecase) CreateTrip(
	ctx context.Context,
	tx pgx.Tx,
	repo repository.BaseTxTripRepository,
	req model.CreateTripRequestModel,
) (*model.CreateTripResponse, error) {
	//getting custom logger context
	logger := logctx.Logger(ctx).With(
		slog.String("layer", "usecase"),
		slog.String("usecase", "TripUsecase.CreateTrip"),
		slog.String("client_id", req.DriverID.String()),
	)

	logger.Info("create trip usecase started")

	entity := req.ToEntity()
	entity.Status = model.StatusDraft
	entity.Seats = 1

	resp, err := repo.CreateTripTx(ctx, tx, entity)

	if err != nil {
		logger.Error(
			"repository create trip failed",
			slog.Any("error", err),
		)
		return nil, fmt.Errorf("repoTrip.Create: %w", err)
	}

	logger.Info(
		"create trip usecase completed",
		slog.String("trip_id", resp.ID.String()),
	)

	return resp.ToCreateResponse(), nil
}
