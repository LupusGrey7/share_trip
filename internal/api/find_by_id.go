package api

import (
	"errors"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"job4j.ru/share_trip/internal/api/apierr"
	"job4j.ru/share_trip/internal/domain/errs"
)

func (s *Server) FindByID(c *fiber.Ctx) error {
	ctx := c.UserContext()

	id := c.Params("tripId")
	if id == "" {
		return fiber.NewError(fiber.StatusBadRequest, invalidIdParamFormat)
	}

	resp, err := s.QueryTripService.GetById(ctx, id)
	if err != nil {
		log.Error("error when FindById trip is: ", err)

		switch {
		case errors.As(err, &errs.RequestValidationError{}):
			return apierr.ErrResponse(c, fiber.StatusBadRequest, err.Error())

		default:
			return apierr.ErrResponse(c, fiber.StatusInternalServerError, internalServerError)
		}
	}
	return c.Status(fiber.StatusOK).JSON(resp)
}
