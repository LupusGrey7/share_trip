//api сценарий - поездки из состояния draft (в транзакции БД)

package api

import (
	"log/slog"

	"job4j.ru/share_trip/internal/domain/trip/model"
	"job4j.ru/share_trip/internal/observability/logctx"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"job4j.ru/share_trip/internal/api/apierr"
	"job4j.ru/share_trip/internal/domain/errs"
)

func (s *Server) CreateTripDraft(c *fiber.Ctx) error {
	ctx := c.UserContext()
	//getting custom logger
	logger := logctx.Logger(ctx).With(
		slog.String("server", "TripServer"),
		slog.String("handler", "CreateTrip"),
	)

	// Достаем ID, который сгенерировал requestid.New()
	traceID := c.GetRespHeader(requestid.ConfigDefault.Header)
	var request model.CreateTripRequest

	// Парсим тело запроса
	if err := c.BodyParser(&request); err != nil {
		logger.Warn(
			"create trip failed: invalid JSON body",
			slog.Any("error", err),
		)

		return c.Status(fiber.StatusBadRequest).JSON(
			fiber.Map{
				"error": invalidParseJson,
			})
	}

	if err := s.validator.Struct(&request); err != nil {
		log.Error(apierr.InvalidValidateError, err)
		logger.Warn("create trip failed: client_id is required")

		return errs.RequestValidationError{Message: err.Error()}
	}
	log.Infof("create trip with traceID: %s", traceID)
	logger = logger.With(
		slog.String("client_id", request.DriverID.String()),
	)

	ctx = logctx.WithLogger(ctx, logger) //update logger in Context app after add new fie0lds
	logger.Info("create trip request accepted",
		slog.String("trip_id", request.DriverID.String()),
	)

	resp, err := s.TripService.CreateTripWithTx(ctx, request)
	if err != nil {
		logger.Error(
			"create trip failed",
			slog.Any("error", err),
		)

		return HandleError(c, err)
	}
	logger.Info(
		"create trip completed",
		slog.String("trip_id", resp.ID.String()),
	)

	return c.Status(fiber.StatusCreated).JSON(resp)
}
