package api

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"job4j.ru/share_trip/internal/api/apierr"
	"job4j.ru/share_trip/internal/domain/errs"
	"job4j.ru/share_trip/internal/domain/trip"
)

//api сценарий - поездки из состояния draft (в транзакции БД)

func (s *Server) CreateTripDraft(c *fiber.Ctx) error {
	ctx := c.UserContext()
	// Достаем ID, который сгенерировал requestid.New()
	traceID := c.GetRespHeader(requestid.ConfigDefault.Header)
	var request trip.CreateTripRequest

	// Парсим тело запроса
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			fiber.Map{
				"error": invalidParseJson,
			})
	}

	if err := s.validator.Struct(&request); err != nil {
		log.Error(invalidValidateError, err)
		return errs.RequestValidationError{Message: err.Error()}
	}
	log.Infof("create trip traceID: %s", traceID)

	resp, err := s.TripService.CreateTripWithTx(ctx, request)
	if err != nil {
		log.Error("error create is: ", err)
		switch {
		case errors.As(err, &errs.RequestValidationError{}):
			return apierr.ErrResponse(c, fiber.StatusBadRequest, err.Error())

		default:
			return apierr.ErrResponse(c, fiber.StatusInternalServerError, internalServerError)
		}
	}
	return c.Status(fiber.StatusCreated).JSON(resp)
}
