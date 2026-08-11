package contracts

import (
	"context"
	"log/slog"

	"job4j.ru/share_trip/internal/client/contracts/model"
	"job4j.ru/share_trip/internal/observability/logctx"
)

const (
	CheckServiceAvailabilityEndpoint = "/api/v2/companies/{companyId}/services/{serviceCode}/availability " // TODO: change to the actual endpoint `check-service` availability endpoint
)

func (c *ContractClient) CheckService(
	ctx context.Context,
	companyID string,
	serviceCode string,
) (model.CheckResult, error) {
	//log the request
	logger := logctx.Logger(ctx).With(
		slog.String("service", "ContractClient"),
		slog.String("operation", "CheckService"),
		slog.String("companyID", companyID),
		slog.String("serviceCode", serviceCode),
	)
	logger.Info("checking service availability")

	var response model.CheckServiceResponse

	resp, err := c.httpClient.R().
		SetContext(ctx).
		SetBody(model.CheckServiceRequest{
			CompanyID:   companyID,
			ServiceCode: serviceCode,
		}).
		SetResult(&response).
		Post(CheckServiceAvailabilityEndpoint)

	if err != nil {
		return model.CheckResult{}, err
	}

	//handle http client error
	if resp.IsError() {
		logger.Error("error checking service availability", "error", MapHTTPClientError(resp))
		return model.CheckResult{}, MapHTTPClientError(resp)
	}

	logger.Info("service availability checked successfully")
	return model.CheckResult{
		Allowed: response.Allowed,
		Reason:  response.Reason,
	}, nil
}
