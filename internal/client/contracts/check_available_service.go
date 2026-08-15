// check_available_service.go - check if a service is available for a company
package contracts

import (
	"context"
	"log/slog"

	"job4j.ru/share_trip/internal/observability/logctx"
)

const (
	// TODO: confirm exact path with Contract Service OpenAPI
	CheckAvailableServiceEndpoint = "/api/v2/companies/{companyId}/services/{serviceCode}/availability"
)

func (c *ContractClient) CheckAvailableService(
	ctx context.Context,
	companyID string,
	serviceCode string,
) (CheckResult, error) {
	logger := logctx.Logger(ctx).With(
		slog.String("service", "ContractClient"),
		slog.String("operation", "CheckAvailableService"),
		slog.String("companyID", companyID),
		slog.String("serviceCode", serviceCode),
	)
	logger.Info("checking service availability")

	var response CheckServiceResponse

	resp, err := c.httpClient.R().
		SetContext(ctx).
		SetHeader("Content-Type", "application/json").
		SetPathParam("companyId", companyID).
		SetPathParam("serviceCode", serviceCode).
		SetResult(&response).
		Get(CheckAvailableServiceEndpoint)

	if err != nil {
		logger.Error("error checking service availability", "error", err)
		return CheckResult{}, err
	}

	if resp.IsError() {
		logger.Error("error checking service availability", "error", resp.Error())
		return CheckResult{}, MapHTTPClientError(resp)
	}

	logger.Info("service availability checked successfully")
	return CheckResult{
		Allowed: response.Allowed,
		Reason:  response.Reason,
	}, nil
}
