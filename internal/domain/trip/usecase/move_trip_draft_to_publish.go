package usecase

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/gofiber/fiber/v2/log"
	"go.opentelemetry.io/otel"
	"job4j.ru/share_trip/internal/domain/trip/model"
	"job4j.ru/share_trip/internal/observability/logctx"

	"github.com/jackc/pgx/v5"
	"job4j.ru/share_trip/internal/repository"
)

func (t *TripUseCase) MoveTripDraftToPublishTx(
	ctx context.Context,
	tx pgx.Tx,
	repo repository.BaseTxTripRepository,
	req model.MoveTripDraftToPublishModel,
) (*model.MoveTripDraftToPublishModelResponse, error) {
	//tracing Jaeger
	ctxSpc, span := otel.Tracer("TripUseCase").Start(ctx, "TripUseCase.MoveTripDraftToPublish")
	defer span.End()

	//getting custom logger context
	logger := logctx.Logger(ctxSpc).With(
		slog.String("layer", "useCase"),
		slog.String("useCase", "TripUseCase.MoveTripDraftToPublish"),
		slog.String("client_id", req.ID),
	)
	logger.Debug("move trip draft to publish useCase started")

	resp, err := repo.GetForUpdateByIDTx(ctxSpc, tx, req.ID)
	if err != nil {
		if errors.Is(err, repository.ErrTripNotFound) {
			logger.Error(
				"get trip repository failed",
				slog.Any("error", err),
			)
			return nil, ErrTripNotFound
		}
		// If this is not a ErrEntityNotFound, This means it's a system failure (500 error)
		return nil, err
	}

	if resp.DriverID != req.ClientID {
		logger.Error(
			"move trip draft to publish failed",
			slog.Any("error", err),
		)
		return nil, fmt.Errorf("%w: client %s is not driver of trip %s", ErrForbidden, req.ClientID, req.ID)
	}

	if resp.Status == model.StatusPublished {
		return &model.MoveTripDraftToPublishModelResponse{
			ID: resp.ID,
		}, nil
	}

	if resp.Status != model.StatusDraft {
		logger.Error(
			"move draft to publish useCase failed",
			slog.Any("error", err),
		)
		return nil, fmt.Errorf("%w: invalid entity status: expected %s", ErrConflict, model.StatusDraft)
	}

	resp.Status = model.StatusPublished

	updatedTrip, err := repo.UpdateTripTx(ctxSpc, tx, resp)
	if err != nil {
		logger.Error("update trip repository failed", slog.Any("error", err))
		return nil, err
	}

	log.Debug("move draft to publish completed", slog.String("trip_id", resp.ID.String()))
	return updatedTrip.ToUpdatedPublishModelResponse(), nil
}
