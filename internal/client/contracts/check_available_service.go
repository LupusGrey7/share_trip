// check_available_service.go - check if a service is available for a company
package contracts

import (
	"context"
	"log/slog"
	"time"

	"job4j.ru/share_trip/internal/observability/logctx"
)

const (
	// TODO: confirm exact path/method with Contract Service OpenAPI / lead
	CheckAvailableServiceEndpoint = "/api/v2/companies/{companyId}/services/{serviceCode}/availability"
)

func (c *ContractClient) CheckAvailableService(
	ctx context.Context,
	companyID string,
	serviceCode string,
) (CheckResult, error) {
	started := time.Now()
	logger := logctx.Logger(ctx).With(
		slog.String("service", "ContractClient"),
		slog.String("operation", "CheckAvailableService"),
		slog.String("company_id", companyID),
		slog.String("service_code", serviceCode),
	)
	logger.Info("checking service availability")

	var response CheckServiceResponse

	resp, err := c.httpClient.R().
		SetContext(ctx).
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept", "application/json").
		ForceContentType("application/json").
		SetPathParam("companyId", companyID).
		SetPathParam("serviceCode", serviceCode).
		SetResult(&response).
		Get(CheckAvailableServiceEndpoint)

	durationMs := time.Since(started).Milliseconds()

	if err != nil {
		mapped := MapTransportError(err)
		logger.Error("contract check transport failed",
			slog.Int64("duration_ms", durationMs),
			slog.String("result", "error"),
			slog.Any("error", mapped),
		)
		return CheckResult{}, mapped
	}

	// Business deny: company/offering missing — not fail closed.
	if IsBusinessNotFound(resp) {
		reason := ReasonFromResponse(resp, "company or service not found")
		logger.Info("contract check denied (not found)",
			slog.Int64("duration_ms", durationMs),
			slog.Int("http_status", resp.StatusCode()),
			slog.String("result", "denied"),
			slog.String("reason", reason),
		)
		return CheckResult{Allowed: false, Reason: reason}, nil
	}

	if resp.IsError() {
		mapped := MapHTTPClientError(resp)
		logger.Error("contract check http error",
			slog.Int64("duration_ms", durationMs),
			slog.Int("http_status", resp.StatusCode()),
			slog.String("result", "error"),
			slog.Any("error", mapped),
		)
		return CheckResult{}, mapped
	}

	result := CheckResult{
		Allowed: response.Allowed,
		Reason:  response.Reason,
	}
	logger.Info("service availability checked",
		slog.Int64("duration_ms", durationMs),
		slog.String("result", "ok"),
		slog.Bool("allowed", result.Allowed),
		slog.String("reason", result.Reason),
	)
	return result, nil
}
