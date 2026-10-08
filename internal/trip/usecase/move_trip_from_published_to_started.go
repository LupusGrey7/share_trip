package usecase

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"go.opentelemetry.io/otel"
	"job4j.ru/share_trip/internal/observability/logctx"
	"job4j.ru/share_trip/internal/storage"
	"job4j.ru/share_trip/internal/trip/domain"
)

// MoveTripFromPublishedToStarted applies published→started under an already-open short transaction.
func (t *TripUseCase) MoveTripFromPublishedToStarted(
	ctx context.Context,
	tx pgx.Tx,
	repo storage.BaseTxTripRepository,
	req domain.MoveTripFromPublishedToStartedInput,
) (*domain.MoveTripFromPublishedToStartedOutput, error) {
	ctxSpc, span := otel.Tracer("TripUseCase").Start(ctx, "TripUseCase.MoveTripFromPublishedToStarted")
	defer span.End()

	logger := logctx.Logger(ctxSpc).With(
		slog.String("layer", "useCase"),
		slog.String("useCase", "TripUseCase.MoveTripFromPublishedToStarted"),
		slog.String("trip_id", req.ID),
		slog.String("company_id", req.CompanyID),
		slog.String("service_code", string(req.ServiceCode)),
		slog.String("client_id", req.ClientID.String()),
	)
	logger.Debug("move trip from published to started useCase started")

	// Contract must already be checked by the service (outside tx).
	if !req.ContractCheck.IsAllowed() {
		return nil, fmt.Errorf("%w: service is not allowed: %s", ErrConflict, req.ContractCheck.Reason)
	}

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

	// Idempotent: already started → 204 via ErrAlreadyDone (not empty DriverID)
	if resp.Status == domain.StatusStarted {
		return nil, fmt.Errorf("%w: trip already started", ErrAlreadyDone)
	}

	if resp.Status != domain.StatusPublished {
		return nil, fmt.Errorf("%w: invalid entity status: expected %s", ErrConflict, domain.StatusPublished)
	}

	resp.Status = domain.StatusStarted

	updatedTrip, err := repo.UpdateTripTx(ctxSpc, tx, resp)
	if err != nil {
		return nil, fmt.Errorf("update trip: %w", err)
	}

	logger.Debug("move trip from published to started useCase completed", slog.String("trip_id", resp.ID.String()))
	return updatedTrip.ToMoveTripFromPublishedToStartedOutput(req.ContractCheck.Allowed, req.ContractCheck.Reason), nil
}
