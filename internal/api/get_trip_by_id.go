//api сценарий - поиска поездки

package api

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"job4j.ru/share_trip/internal/domain/errs"
)

func (s *Server) GetTripById(c *fiber.Ctx) error {
	ctx := c.UserContext()
	// Достаем ID, который сгенерировал requestid.New()
	traceID := c.GetRespHeader(requestid.ConfigDefault.Header)

	tripID := c.Params("tripId")
	if tripID == "" {
		return fiber.NewError(fiber.StatusBadRequest, invalidIdParamFormat)
	}

	request := GetByIDModelRequest{ID: tripID}
	//--validation
	if err := s.validator.Struct(request); err != nil {
		return errs.RequestValidationError{Message: err.Error()}
	}
	// логирование на границе компонента.
	log.Infof("findByTrip ID: %s with traceID: %s ", tripID, traceID)

	resp, err := s.TripService.GetTripByID(ctx, request)
	if err != nil {
		if err != nil {
			log.Errorw(
				"get trip failed",
				"error", err.Error(),
				"trip_id", tripID,
				"trace_id", traceID,
			)
			return HandleError(c, err)
		}
	}
	return c.Status(fiber.StatusOK).JSON(resp)
}
