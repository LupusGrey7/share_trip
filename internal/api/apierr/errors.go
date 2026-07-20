// In this class we will store api errors

package apierr

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"job4j.ru/share_trip/internal/domain/http"
)

const (
	RequestValidationError   = "request validation error"
	StatusNotFound           = "trip not found"
	ErrorForbidden           = "forbidden"
	InternalServerError      = "internal server error"
	InternalServerErrorWith  = "internal server error, %v"
	ErrorClaimsNotFound      = "claims not found in context"
	ErrorForbiddenRole       = "forbidden: missing required client role"
	ErrorForbiddenIDMismatch = "forbidden: driverId does not match authenticated user"
	ErrorBadGateway          = "bad gateway: invalid subject in token"
)

var (
	ErrConflict            = errors.New("conflict")
	ErrForbidden           = errors.New("forbidden")
	ErrInternalServerError = errors.New("internal server error")
	ErrNotFound            = errors.New("not found")
	ErrNotSupported        = errors.New("not supported")
	ErrIllegalArgument     = errors.New("illegal argument provided")
	ErrClaimsNotFound      = errors.New("claims not found in context")
	ErrForbiddenIDMismatch = errors.New("forbidden ID mismatch")
	ErrForbiddenRole       = errors.New("forbidden role for the client")
	ErrBadGateway          = errors.New("bad gateway error")
	ErrInvalidValidate     = errors.New("request validation error")
)

func ErrResponse(
	c *fiber.Ctx,
	code int,
	message string,
) error {
	return c.Status(code).JSON(&http.Response{
		Success: false,
		Message: message,
		Data:    nil,
	})
}
