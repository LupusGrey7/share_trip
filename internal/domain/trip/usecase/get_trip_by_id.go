package usecase

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"job4j.ru/share_trip/internal/domain/trip/model"
	"job4j.ru/share_trip/internal/observability/logctx"
	"job4j.ru/share_trip/internal/repository"
)

func (t *TripUsecase) GetTripByID(
	ctx context.Context,
	tx pgx.Tx,
	repo repository.BaseTxTripRepository,
	req model.GetByIdModelRequest,
) (*model.GetTripByIdModelResponse, error) {
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
			"repository get trip failed",
			slog.Any("error", err),
		)
		return nil, err
	}

	logger.Info(
		"get trip usecase completed",
		slog.String("trip_id", resp.ID.String()),
	)
	return resp.ToGetByIdModelResponse(), nil
}
