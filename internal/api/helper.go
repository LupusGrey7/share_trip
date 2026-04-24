package api

import (
	"errors"
	"github.com/gofiber/fiber/v2"
	"job4j.ru/share_trip/internal/api/apierr"
	"job4j.ru/share_trip/internal/service/usecase"
)

func HandleError(c *fiber.Ctx, err error) error {
	switch { // Проверяем конкретное значение в цепочке
	case errors.Is(err, usecase.ErrForbidden): //403
		return apierr.ErrResponse(c, fiber.StatusForbidden, errors.Unwrap(err).Error()) //  разыменовать цепочку ошибок(вынуть основное описание)
	case errors.Is(err, usecase.ErrTripNotFound): //404
		return apierr.ErrResponse(c, fiber.StatusNotFound, apierr.StatusNotFound) //404
	case errors.Is(err, usecase.ErrConflict):
		return apierr.ErrResponse(c, fiber.StatusConflict, errors.Unwrap(err).Error()) //409
	default:
		return apierr.ErrResponse(c, fiber.StatusInternalServerError, err.Error()) //500
	}
}
