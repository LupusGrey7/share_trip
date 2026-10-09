// check_available_service.go - check if a service is available for a company

package contracts

import (
	"context"
	"log/slog"
	"time"

	"job4j.ru/share_trip/internal/observability/logctx"
)

const (
	CheckAvailableServiceEndpoint = "/api/v2/companies/{companyId}/services/{serviceCode}/availability"
	ContentType                   = "application/json"
)

func (c *Client) CheckAvailableService(ctx context.Context, req CheckServiceRequest) (CheckResult, error) {
	started := time.Now()
	logger := logctx.Logger(ctx).With(
		slog.String("service", "ContractClient"),
		slog.String("operation", "CheckAvailableService"),
		slog.String("company_id", req.CompanyID),
		slog.String("service_code", req.ServiceCode),
	)
	logger.Info("checking service availability")

	var response CheckServiceResponse

	resp, err := c.httpClient.R().
		SetContext(ctx).
		SetHeader("Content-Type", ContentType).
		SetHeader("Accept", ContentType).
		ForceContentType(ContentType).
		SetPathParam("companyId", req.CompanyID).
		SetPathParam("serviceCode", req.ServiceCode).
		SetResult(&response).
		Get(CheckAvailableServiceEndpoint)

	durationMs := time.Since(started).Milliseconds()

	if err != nil {
		mappedTransportErr := MapTransportError(err)
		logger.Error("contract check transport failed",
			slog.Int64("duration_ms", durationMs),
			slog.String("result", "error"),
			slog.Any("error", mappedTransportErr),
		)
		return CheckResult{}, mappedTransportErr
	}

	//function to check if the response is a business deny: company/offering missing — not fail closed.
	if IsBusinessNotFound(resp) {
		businessDenyReason := ReasonFromResponse(resp, "company or service not found")
		logger.Info("contract check denied (not found)",
			slog.Int64("duration_ms", durationMs),
			slog.Int("http_status", resp.StatusCode()),
			slog.String("result", "denied"),
			slog.String("reason", businessDenyReason),
		)
		return CheckResult{Allowed: false, Reason: businessDenyReason}, nil
	}

	//function to map the response to a sentinel error
	if resp.IsError() {
		mappedClientErr := MapHTTPClientError(resp)
		logger.Error("contract check http error",
			slog.Int64("duration_ms", durationMs),
			slog.Int("http_status", resp.StatusCode()),
			slog.String("result", "error"),
			slog.Any("error", mappedClientErr),
		)
		return CheckResult{}, mappedClientErr
	}

	//function to convert the response to a check result
	result := response.ToCheckResult()
	logger.Info("service availability checked",
		slog.Int64("duration_ms", durationMs),
		slog.String("result", "ok"),
		slog.Bool("allowed", result.Allowed),
		slog.String("reason", result.Reason),
	)
	return result, nil
}
