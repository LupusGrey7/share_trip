package apierr

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"job4j.ru/share_trip/internal/domain/http"
)

// В данном классе мы будем хранить api-ошибки

var (
	ErrNotFound        = errors.New("not found")
	ErrNotSupported    = errors.New("not supported")
	ErrIllegalArgument = errors.New("illegal argument provided")
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
