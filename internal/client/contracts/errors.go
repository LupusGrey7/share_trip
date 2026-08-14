package contracts

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/go-resty/resty/v2"
)

const (
	ErrServiceUnavailable = "service is not available"
)

var (
	ErrCheckServiceUnavailable = errors.New("service is not available in the contract")
)

func MapHTTPClientError(resp *resty.Response) error {
	switch resp.StatusCode() {
	case http.StatusBadRequest:
		return fmt.Errorf("HTTP error: %d %s", resp.StatusCode(), resp.Status())
	case http.StatusUnauthorized:
		return fmt.Errorf("HTTP error: %d %s", resp.StatusCode(), resp.Status())
	case http.StatusForbidden:
		return fmt.Errorf("HTTP error: %d %s", resp.StatusCode(), resp.Status())
	case http.StatusNotFound:
		return fmt.Errorf("HTTP error: %d %s", resp.StatusCode(), resp.Status())
	case http.StatusTooManyRequests:
		return fmt.Errorf("HTTP error: %d %s", resp.StatusCode(), resp.Status())
	case http.StatusInternalServerError:
		return fmt.Errorf("HTTP error: %d %s", resp.StatusCode(), resp.Status())
	case http.StatusBadGateway:
		return fmt.Errorf("HTTP error: %d %s", resp.StatusCode(), resp.Status())
	case http.StatusServiceUnavailable:
		return fmt.Errorf("HTTP error: %d %s", resp.StatusCode(), resp.Status())
	default:
		return fmt.Errorf("HTTP error: %d %s", resp.StatusCode(), resp.Status())
	}
}
