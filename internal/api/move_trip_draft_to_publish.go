//api scenario - transferring a trip from the draft to published state.

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
	tripID := c.Params("tripId")
	if tripID == "" {
		return fiber.NewError(fiber.StatusBadRequest, invalidIdParamFormat)
	}
	// Parse request body
	if err := c.BodyParser(&request); err != nil {
		logger.Warn("move trip to publish: invalid json body",
			slog.String("tripId", tripID),
			slog.Any("error", err),
		)
		return c.Status(fiber.StatusBadRequest).JSON(
			fiber.Map{
				"error":  invalidParseJson,
				"reason": err,
			})
	}
	request.ID = tripID

	//--validation
	if err := s.validator.Struct(&request); err != nil {
		logger.Warn("move trip to publish invalid request",
			slog.String("tripId", tripID),
			slog.Any("error", err),
		)
		return errs.RequestValidationError{Message: err.Error()}
	}

	logger = logger.With(
		slog.String("tripId", request.ID),
		slog.String("client_id", request.ClientID.String()),
	)
	ctx = logctx.WithLogger(ctx, logger) //update logger in Context app after add new fields
	logger.Info("move trip to publish")  // logging at the component boundary

	resp, err := s.TripService.MoveTripDraftToPublish(ctx, request.ToRequest())
	if err != nil {
		logger.Error(
			"move trip to publish failed",
			slog.Any("error", err),
		)
		return HandleError(c, err)
	}

	if resp.DriverID == uuid.Nil {
		logger.Info(
			"move trip to publish skipped: no changes detected",
		)
		return c.SendStatus(fiber.StatusNoContent) //http code -204, MUST NOT return a body (JSON)
	}

	logger.Info("move trip to publish completed")
	return c.Status(fiber.StatusOK).JSON(resp) //200
}
