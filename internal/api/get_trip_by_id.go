//api сценарий - поиска поездки

package api

import (
	"job4j.ru/share_trip/internal/domain/trip/model"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/google/uuid"
	"job4j.ru/share_trip/internal/domain/errs"
)

func (s *Server) GetTripById(c *fiber.Ctx) error {
	ctx := c.UserContext()
	// Достаем ID, который сгенерировал requestid.New()
	traceID := c.GetRespHeader(requestid.ConfigDefault.Header)

	id := c.Params("tripId")
	if id == "" {
		return fiber.NewError(fiber.StatusBadRequest, invalidIdParamFormat)
	}
	uuID, err := uuid.Parse(id)
	if err != nil {
		return errs.JsonParseValidationError{Message: err.Error()}
	}

	request := model.GetByIdModelRequest{ID: uuID}

	//--validation
	if err := s.validator.Struct(request); err != nil {
		return errs.RequestValidationError{Message: err.Error()}
	}
	// логирование на границе компонента.
	log.Infof("findByTrip ID: %s with traceID: %s ", id, traceID)

	resp, err := s.TripService.GetTripByID(ctx, request)
	if err != nil {
		return HandleError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(resp)
}
