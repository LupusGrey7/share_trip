package contracts

import (
	"context"
	"net/http"
	"time"

	"github.com/go-resty/resty/v2"
	"job4j.ru/share_trip/internal/client/contracts/model"
)

type BaseContractClient interface {
	CheckService(ctx context.Context, companyID string, serviceCode string) (model.CheckResult, error)
}

type ContractClient struct {
	httpClient *resty.Client
}

func NewContractClient(baseURL string) *ContractClient {
	return &ContractClient{
		httpClient: resty.New().
			SetBaseURL(baseURL).
			SetTimeout(2 * time.Second).
			SetRetryCount(2). //it means 1+2 retry
			SetRetryWaitTime(200 * time.Millisecond).
			SetRetryMaxWaitTime(1 * time.Second).
			AddRetryCondition(func(r *resty.Response, err error) bool {
				if err != nil {
					return true
				}
				return r.StatusCode() == http.StatusTooManyRequests ||
					r.StatusCode() == http.StatusBadGateway ||
					r.StatusCode() == http.StatusServiceUnavailable ||
					r.StatusCode() == http.StatusGatewayTimeout
			}),
	}
}
