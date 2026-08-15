package contracts

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"

	"github.com/go-resty/resty/v2"
)

// MapHTTPClientError maps Contract HTTP error response → sentinel (%w for errors.Is).
func MapHTTPClientError(resp *resty.Response) error {
	if resp == nil {
		return ErrUnavailable
	}

	code := resp.StatusCode()
	switch code {
	case http.StatusBadRequest, http.StatusNotFound:
		return fmt.Errorf("%w: status %d", ErrBadRequest, code)
	case http.StatusUnauthorized, http.StatusForbidden:
		return fmt.Errorf("%w: status %d", ErrForbidden, code)
	case http.StatusRequestTimeout, http.StatusGatewayTimeout:
		return fmt.Errorf("%w: status %d", ErrTimeout, code)
	case http.StatusTooManyRequests,
		http.StatusBadGateway,
		http.StatusServiceUnavailable,
		http.StatusInternalServerError:
		return fmt.Errorf("%w: status %d", ErrUnavailable, code)
	default:
		return fmt.Errorf("%w: status %d", ErrUnavailable, code)
	}
}

// MapTransportError maps network / deadline errors from resty → sentinel.
func MapTransportError(err error) error {
	if err == nil {
		return nil
	}
	if isTimeoutErr(err) {
		return fmt.Errorf("%w: %v", ErrTimeout, err)
	}
	return fmt.Errorf("%w: %v", ErrUnavailable, err)
}

func isTimeoutErr(err error) bool {
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}
	if errors.Is(err, os.ErrDeadlineExceeded) {
		return true
	}
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return true
	}
	return false
}
