//api сценарий - перевода поездки из состояния draft в published.

package api

import (
	"log/slog"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/google/uuid"
	"job4j.ru/share_trip/internal/domain/errs"
	"job4j.ru/share_trip/internal/observability/logctx"
)

func (s *Server) MoveTripDraftToPublishTx(c *fiber.Ctx) error {
	ctx := c.UserContext()
	// We get the ID that was generated requestid.New()
	traceID := c.GetRespHeader(requestid.ConfigDefault.Header)
	//getting custom logger
	logger := logctx.Logger(ctx).With(
		slog.String("server", "TripServer"),
		slog.String("handler", "CreateTrip"),
		slog.String("traceID", traceID),
	)

	var request MoveTripDraftToPublishRequestModel

	// request param
	id := c.Params("tripId")
	if id == "" {
		return fiber.NewError(fiber.StatusBadRequest, invalidIdParamFormat)
	}
	// Parse request body
	if err := c.BodyParser(&request); err != nil {
		logger.Warn("trip get validation failed",
			slog.String("tripId", id),
			slog.String("error", err.Error()),
		)
		return c.Status(fiber.StatusBadRequest).JSON(
			fiber.Map{
				"error": invalidParseJson,
			})
	}
	request.ID = id

	//--validation
	if err := s.validator.Struct(&request); err != nil {
		logger.Warn("trip get validation failed",
			slog.String("tripId", id),
			slog.String("error", err.Error()),
		)
		return errs.RequestValidationError{Message: err.Error()}
	}
	// logging at the component boundary
	logger.Info(
		"move trip to publish",
		"trip_id", request.ID,
	)

	resp, err := s.TripService.MoveTripDraftToPublish(ctx, request.ToRequest())
	if err != nil {
		logger.Error(
			"move trip to publish failed",
			slog.Any("error", err),
		)
		return HandleError(c, err)
	}

	if resp.DriverID == uuid.Nil {
		return c.Status(fiber.StatusNoContent).JSON(resp) //204
	}
	logger.Info(
		"move trip to publish completed",
		slog.String("trip_id", resp.ID.String()),
	)
	return c.Status(fiber.StatusOK).JSON(resp) //200
}
