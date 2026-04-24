//api сценарий - перевода поездки из состояния draft в published.

package api

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/google/uuid"
	"job4j.ru/share_trip/internal/api/apierr"
	"job4j.ru/share_trip/internal/domain/errs"
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
	// Parse request body
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			fiber.Map{
				"error": invalidParseJson,
			})
	}
	request.ID = id
	//--validation
	if err := s.validator.Struct(&request); err != nil {
		log.Error(apierr.InvalidValidateError, err)
		return errs.RequestValidationError{Message: err.Error()}
	}

	uuID, err := uuid.Parse(id)
	if err != nil {
		log.Error(apierr.InvalidValidateError, err)
		return apierr.ErrResponse(c, fiber.StatusInternalServerError, apierr.InternalServerError)
	}

	// логирование на границе компонента.
	log.Infof("move trip to publish ID: %v with traceID: %s ", uuID, traceID)

	resp, err := s.TripService.MoveTripDraftToPublish(ctx, request.ToRequest(uuID))
	if err != nil {
		return HandleError(c, err)
	}

	if resp.DriverID == uuid.Nil {
		return c.Status(fiber.StatusNoContent).JSON(resp) //204
	}

	return c.Status(fiber.StatusOK).JSON(resp) //200
}
