package contracts

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/go-resty/resty/v2"
)

// MapHTTPClientError maps Contract HTTP error response → sentinel (%w for errors.Is).
// 404 is business deny (company/offering not found) — handled in CheckAvailableService, not here.
func MapHTTPClientError(resp *resty.Response) error {
	if resp == nil {
		return ErrUnavailable
	}

	code := resp.StatusCode()
	switch code {
	case http.StatusBadRequest:
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

// IsBusinessNotFound reports Contract responses that mean "no such company/offering"
// (бизнес-отказ), not "Contract is down".
func IsBusinessNotFound(resp *resty.Response) bool {
	if resp == nil {
		return false
	}
	code := resp.StatusCode()
	if code == http.StatusNotFound {
		return true
	}
	// Some Contract handlers return 400/422 with "company not found" in body.
	if code == http.StatusBadRequest || code == http.StatusUnprocessableEntity {
		return strings.Contains(strings.ToLower(resp.String()), "company not found")
	}
	return false
}

// ReasonFromResponse extracts a short business reason from Contract error/success body.
func ReasonFromResponse(resp *resty.Response, fallback string) string {
	if resp == nil {
		return fallback
	}
	body := strings.TrimSpace(resp.String())
	if body == "" {
		return fallback
	}

	var payload struct {
		Reason  string `json:"reason"`
		Message string `json:"message"`
		Error   string `json:"error"`
	}
	if err := json.Unmarshal([]byte(body), &payload); err == nil {
		switch {
		case payload.Reason != "":
			return payload.Reason
		case payload.Message != "":
			return payload.Message
		case payload.Error != "":
			return payload.Error
		}
	}
	if len(body) > 200 {
		return body[:200]
	}
	return body
}
