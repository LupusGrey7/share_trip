package usecase

import (
	"context"
	"errors"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"go.opentelemetry.io/otel"
	"job4j.ru/share_trip/internal/observability/logctx"
	"job4j.ru/share_trip/internal/storage"
	"job4j.ru/share_trip/internal/trip/domain"
)

func (t *TripUseCase) GetTripByIDTx(
	ctx context.Context,
	tx pgx.Tx,
	repo storage.BaseTxTripRepository,
	input *domain.GetByIDInput,
) (*domain.GetTripByIDOutput, error) {
	ctxSpc, span := otel.Tracer("TripUseCase").Start(ctx, "TripUseCase.GetTripByIDTx")
	defer span.End()

	logger := logctx.Logger(ctxSpc).With(
		slog.String("layer", "useCase"),
		slog.String("useCase", "TripUseCase.GetTripByIDTx"),
		slog.String("trip_id", input.ID),
	)
	logger.Debug("get trip by ID useCase started")

	entity, err := repo.GetTripByIDTx(ctxSpc, tx, input.ID)
	if err != nil {
		logger.Error("get trip by ID useCase failed", slog.Any("error", err))
		if errors.Is(err, storage.ErrTripNotFound) {
			return nil, ErrTripNotFound
		}
		return nil, err
	}

	logger.Debug("get trip by ID useCase completed")
	return domain.TripEntityToOutput(entity), nil
}
