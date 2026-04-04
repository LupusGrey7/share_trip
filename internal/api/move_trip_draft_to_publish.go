package api

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"job4j.ru/share_trip/internal/api/apierr"
	"job4j.ru/share_trip/internal/domain/errs"
	"job4j.ru/share_trip/internal/domain/trip"
)

//сценарий перевода поездки из состояния draft в published.

func (s *Server) UpdateTripDraftToPublishTx(c *fiber.Ctx) error {
	ctx := c.UserContext()
	var request trip.MoveTripDraftToPublishModelRequest

	id := c.Params("tripId")
	if id == "" {
		return fiber.NewError(fiber.StatusBadRequest, invalidIdParamFormat)
	}
	// Парсим тело запроса
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			fiber.Map{
				"error": invalidParseJson,
			})
	}

	request.ID = id

	resp, err := s.CommandTripService.MoveTripDraftToPublish(ctx, request)
	if err != nil {
		log.Error("error update is: ", err)
		switch {
		case errors.As(err, &errs.RequestValidationError{}):
			return apierr.ErrResponse(c, fiber.StatusBadRequest, err.Error())

		default:
			return apierr.ErrResponse(c, fiber.StatusInternalServerError, internalServerError)
		}
	}
	return c.Status(fiber.StatusCreated).JSON(resp)
}
