package api

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/google/uuid"
	"job4j.ru/share_trip/internal/api/apierr"
	"job4j.ru/share_trip/internal/domain/errs"
)

//api сценарий - перевода поездки из состояния draft в published.

const (
	invalidValidateError string = "Validation errors: %v\n"
)

func (s *Server) MoveTripDraftToPublishTx(c *fiber.Ctx) error {
	ctx := c.UserContext()
	// Достаем ID, который сгенерировал requestid.New()
	traceID := c.GetRespHeader(requestid.ConfigDefault.Header)
	var request MoveTripDraftToPublishModelRequest

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
	//--validation
	if err := s.validator.Struct(&request); err != nil {
		log.Error(invalidValidateError, err)
		return errs.RequestValidationError{Message: err.Error()}
	}

	uuID, err := uuid.Parse(id)
	if err != nil {
		log.Error(invalidValidateError, err)
		return errs.JsonParseValidationError{Message: err.Error()}
	}
	request.ID = uuID
	// логирование на границе компонента.
	log.Infof("move trip to publish ID: %v with traceID: %s ", uuID, traceID)

	resp, err := s.CommandTripService.MoveTripDraftToPublish(ctx, request.ToRequest())
	if err != nil {
		log.Error("error move trip to publish is: ", err)
		switch {
		case errors.As(err, &errs.RequestValidationError{}):
			return apierr.ErrResponse(c, fiber.StatusBadRequest, err.Error())

		default:
			return apierr.ErrResponse(c, fiber.StatusInternalServerError, internalServerError)
		}
	}
	return c.Status(fiber.StatusOK).JSON(resp)
}
