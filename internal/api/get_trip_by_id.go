package api

import (
	"errors"
	"job4j.ru/share_trip/internal/domain/trip/model"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/google/uuid"
	"job4j.ru/share_trip/internal/api/apierr"
	"job4j.ru/share_trip/internal/domain/errs"
)

//api сценарий - поиска поездки

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
		log.Errorf(apierr.InvalidValidateError, err)
		return errs.JsonParseValidationError{Message: err.Error()}
	}

	request := model.GetByIdModelRequest{ID: uuID}
	//--validation
	if err := s.validator.Struct(request); err != nil {
		log.Error(apierr.InvalidValidateError, err)
		return errs.RequestValidationError{Message: err.Error()}
	}

	log.Infof("find Bytrip ID: %s with traceID: %s ", id, traceID)
	resp, err := s.TripService.GetTripByID(ctx, request)
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
