package api

import (
	"errors"
	"job4j.ru/share_trip/internal/service/usecase"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/google/uuid"
	"job4j.ru/share_trip/internal/api/apierr"
	"job4j.ru/share_trip/internal/domain/errs"
)

//api сценарий - перевода поездки из состояния draft в published.

const (
	invalidValidateError = "Validation errors: %v\n"
	statusNotFound       = "trip not found"
	errorForbidden       = "forbidden"
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

	resp, err := s.TripService.MoveTripDraftToPublish(ctx, request.ToRequest())
	if err != nil {
		switch { // Проверяем конкретное значение в цепочке
		case errors.Is(err, usecase.ErrForbidden): //403
			return apierr.ErrResponse(c, fiber.StatusForbidden, errors.Unwrap(err).Error()) //  разыменовать цепочку ошибок(вынуть основное описание)
		case errors.Is(err, usecase.ErrTripNotFound): //404
			return apierr.ErrResponse(c, fiber.StatusNotFound, statusNotFound) //404
		case errors.Is(err, usecase.ErrConflict):
			return apierr.ErrResponse(c, fiber.StatusConflict, errors.Unwrap(err).Error()) //409
		default:
			return apierr.ErrResponse(c, fiber.StatusInternalServerError, err.Error()) //500
		}
	}
	if resp.DriverID == uuid.Nil {
		return c.Status(fiber.StatusNoContent).JSON(resp) //204
	}

	return c.Status(fiber.StatusOK).JSON(resp) //200
}
