package usecase

import (
	"context"
	"fmt"
	"job4j.ru/share_trip/internal/observability/logctx"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"job4j.ru/share_trip/internal/domain/trip"
	"job4j.ru/share_trip/internal/repository"
)

func (t *TripUsecase) CreateTrip(
	ctx context.Context,
	tx pgx.Tx,
	repo repository.BaseTxTripRepository,
	req trip.CreateTripRequest,
) (*trip.CreateTripResponse, error) {
	logger := logctx.Logger(ctx).With(
		slog.String("layer", "usecase"),
		slog.String("usecase", "TripUsecase.CreateTrip"),
		slog.String("client_id", req.DriverID.String()),
	)

	logger.Info("create trip usecase started")

	entity := req.ToEntity()
	entity.Status = trip.StatusDraft
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
