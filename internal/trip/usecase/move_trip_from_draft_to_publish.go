package usecase

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"go.opentelemetry.io/otel"

	"job4j.ru/share_trip/internal/observability/logctx"
	"job4j.ru/share_trip/internal/storage"
	"job4j.ru/share_trip/internal/trip/domain"

	"github.com/jackc/pgx/v5"
)

func (t *TripUseCase) MoveTripFromDraftToPublish(
	ctx context.Context,
	tx pgx.Tx,
	repo storage.BaseTxTripRepository,
	req domain.MoveTripFromDraftToPublishInput,
) (*domain.MoveTripFromDraftToPublishOutput, error) {
	ctxSpc, span := otel.Tracer("TripUseCase").Start(ctx, "TripUseCase.MoveTripFromDraftToPublish")
	defer span.End()

	logger := logctx.Logger(ctxSpc).With(
		slog.String("layer", "useCase"),
		slog.String("useCase", "TripUseCase.MoveTripFromDraftToPublish"),
		slog.String("client_id", req.ID),
	)
	logger.Debug("move trip from draft to publish useCase started")

	resp, err := repo.GetForUpdateByIDTx(ctxSpc, tx, req.ID)
	if err != nil {
		if errors.Is(err, storage.ErrTripNotFound) {
			return nil, ErrTripNotFound
		}
		return nil, fmt.Errorf("get trip for update: %w", err)
	}

	if resp.DriverID != req.ClientID {
		return nil, fmt.Errorf("%w: client %s is not driver of trip %s", ErrForbidden, req.ClientID, req.ID)
	}

	if resp.Status == domain.StatusPublished {
		return nil, ErrAlreadyDone
	}

	if resp.Status != domain.StatusDraft {
		return nil, fmt.Errorf("%w: invalid entity status: expected %s", ErrConflict, domain.StatusDraft)
	}

	resp.Status = domain.StatusPublished

	updatedTrip, err := repo.UpdateTripTx(ctxSpc, tx, resp)
	if err != nil {
		return nil, fmt.Errorf("update trip: %w", err)
	}

	logger.Debug("move trip from draft to publish useCase completed", slog.String("trip_id", resp.ID.String()))
	return updatedTrip.ToMoveTripFromDraftToPublishOutput(), nil
}
