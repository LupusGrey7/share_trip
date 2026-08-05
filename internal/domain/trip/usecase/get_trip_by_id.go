package usecase

import (
	"context"
	"log/slog"

	"go.opentelemetry.io/otel"

	"errors"

	"github.com/jackc/pgx/v5"
	"job4j.ru/share_trip/internal/domain/trip/model"
	"job4j.ru/share_trip/internal/observability/logctx"
	"job4j.ru/share_trip/internal/repository"
)

func (t *TripUseCase) GetTripByIDTx(
	ctx context.Context,
	tx pgx.Tx,
	repo repository.BaseTxTripRepository,
	req model.GetByIDModelRequest,
) (*model.GetTripByIDModelResponse, error) {
	//tracing
	ctxSpc, span := otel.Tracer("TripUseCase").Start(ctx, "TripUseCase.GetTripByIDTx")
	defer span.End()

	//getting custom logger context
	logger := logctx.Logger(ctxSpc).With(
		slog.String("layer", "useCase"),
		slog.String("useCase", "TripUseCase.GetTripByIDTx"),
		slog.String("client_id", req.ID),
	)
	logger.Debug("get trip by ID useCase started")

	resp, err := repo.GetTripByIDTx(ctxSpc, tx, req.ID)
	if err != nil {
		logger.Error("get trip by ID useCase failed", slog.Any("error", err))
		if errors.Is(err, repository.ErrTripNotFound) {
			return nil, ErrTripNotFound
		}
		// If this is not a ErrEntityNotFound, This means it's a system error (500 error)
		return nil, err
	}

	logger.Debug("get trip by ID useCase completed")
	return resp.ToGetByIdModelResponse(), nil
}
