package usecase

import (
	"context"
	"log/slog"

	"go.opentelemetry.io/otel"
	contracts "job4j.ru/share_trip/internal/clients/http/contract"
	"job4j.ru/share_trip/internal/observability/logctx"
)

// CheckStartAllowed calls Contract Service outside DB transaction (lead sequence).
func (t *TripUseCase) CheckServiceIsAllowed(
	ctx context.Context,
	companyID string,
	serviceCode string,
) (contracts.CheckResult, error) {
	ctxSpc, span := otel.Tracer("TripUseCase").Start(ctx, "TripUseCase.CheckServiceIsAllowed")
	defer span.End()

	logger := logctx.Logger(ctxSpc).With(
		slog.String("layer", "useCase"),
		slog.String("useCase", "TripUseCase.CheckServiceIsAllowed"),
		slog.String("company_id", companyID),
		slog.String("service_code", serviceCode),
	)
	logger.Debug("contract check for service is allowed started")

	result, err := t.contractUseCase.CheckAvailableService(ctxSpc, companyID, serviceCode)
	if err != nil {
		logger.Error("contract check for service is allowed failed", slog.Any("error", err))
		return contracts.CheckResult{}, err
	}

	logger.Debug("contract check for service is allowed completed",
		slog.Bool("allowed", result.Allowed),
		slog.String("reason", result.Reason),
	)
	return result, nil
}
