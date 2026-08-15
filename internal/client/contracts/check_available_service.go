package contracts

import (
	"context"
	"log/slog"

	"job4j.ru/share_trip/internal/observability/logctx"
)

const (
	CheckAvailableServiceEndpoint = "/api/v2/companies/{companyId}/services/{serviceCode}/availability " // TODO: change to the actual endpoint `check-service` availability endpoint
)

func (c *ContractClient) CheckService(
	ctx context.Context,
	companyID string,
	serviceCode string,
) (CheckResult, error) {
	//log the request
	logger := logctx.Logger(ctx).With(
		slog.String("service", "ContractClient"),
		slog.String("operation", "CheckService"),
		slog.String("companyID", companyID),
		slog.String("serviceCode", serviceCode),
	)
	logger.Info("checking service availability")

	var response CheckServiceResponse

	resp, err := c.httpClient.R(). //R() create a new request
					SetContext(ctx).
					SetHeader("Content-Type", "application/json"). //SetHeader() set the header for the request
					SetPathParam("companyId", companyID).
					SetPathParam("serviceCode", serviceCode).
		// SetBody(CheckServiceRequest{
		// 	CompanyID:   companyID,
		// 	ServiceCode: serviceCode,
		// }). //fix me - Body is not used in the request? 13/08/2026
		SetResult(&response).
		Post(CheckAvailableServiceEndpoint) //Post() send a POST request

	if err != nil {
		return CheckResult{}, err
	}

	//handle http client error
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
