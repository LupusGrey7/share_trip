//api сценарий - поиска поездки

package api

import (
	"log/slog"

	"job4j.ru/share_trip/internal/domain/trip/model"
	"job4j.ru/share_trip/internal/observability/logctx"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"job4j.ru/share_trip/internal/domain/errs"
)

func (s *Server) GetTripById(c *fiber.Ctx) error {
	ctx := c.UserContext()
	// We get the ID that was generated requestid.New()
	traceID := c.GetRespHeader(requestid.ConfigDefault.Header)
	//getting custom logger context
	logger := logctx.Logger(ctx).With(
		slog.String("server", "TripServer"),
		slog.String("handler", "CreateTrip"),
		slog.String("traceID", traceID),
	)

	id := c.Params("tripId")
	if id == "" {
		return fiber.NewError(fiber.StatusBadRequest, invalidIdParamFormat)
	}

	request := GetTripByIdRequestModel{ID: id}

	//--validation
	if err := s.validator.Struct(request); err != nil {
		logger.Warn("trip get validation failed",
			slog.String("tripId", id),
			slog.String("error", err.Error()),
		)
		return errs.RequestValidationError{Message: err.Error()}
	}
	// logging at the component boundary
	logger.Info("get Trip By ID",
		"trip_id", id,
	)

	resp, err := s.TripService.GetTripByID(ctx, model.GetByIdModelRequest{ID: request.ID})
	if err != nil {
		logger.Error(
			"get Trip By ID failed",
			slog.Any("error", err),
		)
		return HandleError(c, err)
	}

	logger.Info(
		"get Trip By ID completed",
		slog.String("trip_id", resp.ID.String()),
	)
	return c.Status(fiber.StatusOK).JSON(resp)
}
