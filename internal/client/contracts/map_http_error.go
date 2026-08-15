package contracts

import (
	"fmt"
	"net/http"

	"github.com/go-resty/resty/v2"
)

// MapHTTPClientError — mapping HTTP response to error.
// This is not an error class, but an adapter: status/body → error (later → sentinel from errors.go).
// Separate file, same package contracts — not client_helper and not clienterr package.
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
