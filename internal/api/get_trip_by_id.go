//api сценарий - поиска поездки

package api

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/google/uuid"
	"job4j.ru/share_trip/internal/api/apierr"
	"job4j.ru/share_trip/internal/domain/trip"
)

func (s *Server) GetTripById(c *fiber.Ctx) error {
	ctx := c.UserContext()
	// Достаем ID, который сгенерировал requestid.New()
	traceID := c.GetRespHeader(requestid.ConfigDefault.Header)
	var t trip.GetByIDModelRequest

	tripID := c.Params("tripId")
	if tripID == "" {
		return fiber.NewError(fiber.StatusBadRequest, invalidIdParamFormat)
	}

	request := GetByIDModelRequest{ID: tripID}
	//--validation
	if err := s.validator.Struct(request); err != nil {
		log.Error(apierr.InvalidValidateError, err)
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	uuID, err := uuid.Parse(tripID)
	if err != nil {
		log.Errorf(apierr.InternalServerErrorWith, err)
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	t.ID = uuID
	log.Infof("find By Trip ID: %s with traceID: %s ", tripID, traceID)

	resp, err := s.TripService.GetTripByID(ctx, t)
	if err != nil {
		log.Errorw(
			"get trip failed",
			"error", err.Error(),
			"trip_id", tripID,
			"trace_id", traceID,
		)
		return HandleError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(resp)
}
