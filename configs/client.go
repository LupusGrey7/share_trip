// config for resty client
package configs

import (
	"net/http"
	"time"

	"github.com/go-resty/resty/v2"
)

const (
	ContractServiceEnv = "CONTRACT_SERVICE_URL"  // environment variable for the contract service url
	BaseURL            = "http://localhost:8082" // base url for the contract service

	Timeout          = 2 * time.Second        // Every trying to get response from server, it waits for 2 seconds
	RetryCount       = 2                      // retry count for the http request (1 try + 2 retries)
	RetryWaitTime    = 200 * time.Millisecond // retry wait time for the http request
	RetryMaxWaitTime = 1 * time.Second        // retry max wait time for the http request
	MaxAttempts      = 3                      // max attempts (1 try + RetryCount)
)

// RetryConditionFunc returns true if the request should be retried.
var RetryConditionFunc = func(resp *resty.Response, err error) bool {
	if err != nil {
		return true // network error or timeout
	}
	return resp.StatusCode() == http.StatusTooManyRequests || // 429
		resp.StatusCode() == http.StatusBadGateway || // 502
		resp.StatusCode() == http.StatusServiceUnavailable || // 503
		resp.StatusCode() == http.StatusGatewayTimeout // 504
}
