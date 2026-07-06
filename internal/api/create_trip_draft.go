//API scenario - trips from the draft state (in a DB transaction)

package api

import (
	"log/slog"

	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel"
	"job4j.ru/share_trip/internal/domain/errs"
	"job4j.ru/share_trip/internal/domain/trip/model"
	"job4j.ru/share_trip/internal/observability/logctx"
)

func (s *Server) CreateTripDraft(c *fiber.Ctx) error {
	// OpenTelemetry: child span внутри root HTTP span от otelfiber middleware
	tracer := otel.Tracer("trip-api")
	ctx, span := tracer.Start(c.UserContext(), "CreateTripDraftHandler")
	traceID := span.SpanContext().TraceID().String()
	c.Set("X-Request-ID", traceID)
	defer span.End()

	//getting custom logger context
	logger := logctx.Logger(ctx).With(
		slog.String("server", "TripServer"),
		slog.String("handler", "CreateTripDraft"),
		slog.String("trace_id", traceID), // Key field for Grafana
	)

	var request CreateTripRequestModel

	// Parsing the request body
	if err := c.BodyParser(&request); err != nil {
		logger.Warn("create trip draft failed: invalid json body", slog.Any("error", err))
		return c.Status(fiber.StatusBadRequest).JSON(
			fiber.Map{
				"error":  invalidParseJson,
				"reason": err,
			})
	}

	if err := s.validator.Struct(&request); err != nil {
		logger.Warn(
			"create trip draft failed: invalid request",
			slog.Any("error", err),
		)
		return errs.RequestValidationError{Message: err.Error()}
	}

	logger = logger.With(slog.String("client_id", request.DriverID.String()))
	ctx = logctx.WithLogger(ctx, logger)               //update logger in Context app after add new fields
	logger.Debug("create trip draft request accepted") // logging at the component boundary

	resp, err := s.TripService.CreateTripDraft(
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
		logger.Error("create trip draft failed", slog.Any("error", err))
		return HandleError(c, err)
	}

	logger.Debug("create trip draft completed") // logging at the component boundary
	return c.Status(fiber.StatusCreated).JSON(resp)
}
