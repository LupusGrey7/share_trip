package contracts

import (
	"context"
	"net/http"
	"time"

	"github.com/go-resty/resty/v2"
)

type BaseContractClient interface {
	CheckService(ctx context.Context, companyID string, serviceCode string) (CheckResult, error)
}

type ContractClient struct {
	httpClient *resty.Client // httpClient — http resty client for the contract client
}

func NewContractClient(baseURL string) *ContractClient {
	return &ContractClient{ //set up the http resty client
		httpClient: resty.New(). //create a new http resty client
						SetBaseURL(baseURL).
						SetTimeout(2 * time.Second). //set the timeout for the http request
						SetRetryCount(2).            //it means 1+2 retry
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
