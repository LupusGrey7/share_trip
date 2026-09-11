package contracts

import (
	"context"

	"github.com/go-resty/resty/v2"
	"job4j.ru/share_trip/configs"
)

type BaseContractClient interface {
	CheckAvailableService(ctx context.Context, companyID string, serviceCode string) (CheckResult, error)
}

type ContractClient struct {
	httpClient *resty.Client //library for making http requests
}

func NewContractClient(baseURL string) *ContractClient {
	return &ContractClient{
		httpClient: resty.New().
			SetBaseURL(baseURL).
			SetTimeout(configs.Timeout).
			SetRetryCount(configs.RetryCount).
			SetRetryWaitTime(configs.RetryWaitTime).
			SetRetryMaxWaitTime(configs.RetryMaxWaitTime).
			AddRetryCondition(configs.RetryConditionFunc),
	}
}
