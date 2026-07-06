package usecase

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"go.opentelemetry.io/otel"
	"job4j.ru/share_trip/internal/domain/trip/model"
	"job4j.ru/share_trip/internal/observability/logctx"
	"job4j.ru/share_trip/internal/repository"
)

func (t *TripUseCase) CreateTripDraft(
	ctx context.Context,
	tx pgx.Tx,
	repo repository.BaseTxTripRepository,
	req model.CreateTripRequestModel,
) (*model.CreateTripDraftResponse, error) {
	//tracing Jaeger
	ctxSpc, span := otel.Tracer("TripUseCase").Start(ctx, "TripUseCase.CreateTripDraft")
	defer span.End()

	//getting custom logger context
	logger := logctx.Logger(ctxSpc).With(
		slog.String("layer", "usecase"),
		slog.String("usecase", "TripUseCase.CreateTripDraft"),
		slog.String("client_id", req.DriverID.String()),
	)

	logger.Debug("create trip draft usecase started")

	entity := req.ToEntity()
	entity.Status = model.StatusDraft

	resp, err := repo.CreateTripDraftTx(ctxSpc, tx, entity)

	if err != nil {
		logger.Error("repository create trip draft failed", slog.Any("error", err))
		return nil, fmt.Errorf("repoTrip.Create: %w", err)
	}

	logger.Debug(
		"create trip draft usecase completed",
		slog.String("trip_id", resp.ID.String()),
	)

	return resp.ToCreateResponse(), nil
}
