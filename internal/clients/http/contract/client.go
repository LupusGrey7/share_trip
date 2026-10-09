package contracts

import (
	"context"

	"github.com/go-resty/resty/v2"
	"job4j.ru/share_trip/internal/config"
)

type ContractClient interface {
	CheckAvailableService(ctx context.Context, req CheckServiceRequest) (CheckResult, error)
}

type Client struct {
	httpClient *resty.Client //library for making http requests
}

func NewClient(baseURL string) *Client {
	return &Client{
		httpClient: resty.New().
			SetBaseURL(baseURL).
			SetTimeout(config.Timeout).
			SetRetryCount(config.RetryCount).
			SetRetryWaitTime(config.RetryWaitTime).
			SetRetryMaxWaitTime(config.RetryMaxWaitTime).
			AddRetryCondition(config.RetryConditionFunc),
	}
}
