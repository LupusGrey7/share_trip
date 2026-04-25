package api

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"job4j.ru/share_trip/internal/api/apierr"
	"job4j.ru/share_trip/internal/domain/errs"
)

//api сценарий - поездки из состояния draft (в транзакции БД)

func (s *Server) CreateTripDraft(c *fiber.Ctx) error {
	ctx := c.UserContext()
	// Достаем ID, который сгенерировал requestid.New()
	traceID := c.GetRespHeader(requestid.ConfigDefault.Header)
	var request CreateTripRequest

	// Парсим тело запроса
	if err := c.BodyParser(&request); err != nil {
		e := invalidParseJson + err.Error()
		return c.Status(fiber.StatusBadRequest).JSON(
			fiber.Map{
				"error": e,
			})
	}

	if err := s.validator.Struct(&request); err != nil {
		log.Error(apierr.InvalidValidateError, err)
		return errs.RequestValidationError{Message: err.Error()}
	}
	log.Infof("create trip with traceID: %s, DriveID : %v", traceID, request.DriverID)

	resp, err := s.TripService.CreateTripWithTx(ctx, request.ToCreateTripDomainRequest())
	if err != nil {
		log.Error("error create draft is: ", err)
		return HandleError(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(resp)
}
