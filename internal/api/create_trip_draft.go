//API scenario - trips from the draft state (in a DB transaction)

package api

import (
	"log/slog"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"job4j.ru/share_trip/internal/domain/errs"
	"job4j.ru/share_trip/internal/domain/trip/model"
	"job4j.ru/share_trip/internal/observability/logctx"
)

func (s *Server) CreateTripDraft(c *fiber.Ctx) error {
	ctx := c.UserContext()
	// We get the ID that was generated requestid.New()
	traceID := c.GetRespHeader(requestid.ConfigDefault.Header)
	//getting custom logger context
	logger := logctx.Logger(ctx).With(
		slog.String("server", "TripServer"),
		slog.String("handler", "CreateTrip"),
		slog.String("traceID", traceID),
	)

	var request CreateTripRequestModel

	// Parsing the request body
	if err := c.BodyParser(&request); err != nil {
		logger.Warn(
			"create trip failed: invalid json body",
			slog.Any("error", err),
		)
		return c.Status(fiber.StatusBadRequest).JSON(
			fiber.Map{
				"error":  invalidParseJson,
				"reason": err,
			})
	}

	if err := s.validator.Struct(&request); err != nil {
		logger.Warn(
			"create trip failed: invalid request",
			slog.Any("error", err),
		)
		return errs.RequestValidationError{Message: err.Error()}
	}

	logger = logger.With(
		slog.String("client_id", request.DriverID.String()),
	)
	ctx = logctx.WithLogger(ctx, logger) //update logger in Context app after add new fields
	logger.Info("create trip request accepted")

	logger = logger.With(
		slog.String("client_id", request.DriverID.String()),
	)
	ctx = logctx.WithLogger(ctx, logger)        //update logger in Context app after add new fields
	logger.Info("create trip request accepted") // logging at the component boundary

	resp, err := s.TripService.CreateTripWithTx(
		ctx,
		model.CreateTripRequestModel{
			DriverID:       request.DriverID,
			FromPoint:      request.FromPoint,
			ToPoint:        request.ToPoint,
			DepartureTime:  request.DepartureTime,
			AvailableSeats: request.AvailableSeats,
		},
	)
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
