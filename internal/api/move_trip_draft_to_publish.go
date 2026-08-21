//api scenario - transferring a trip from the draft to published state.

package api

import (
	"log/slog"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"job4j.ru/share_trip/internal/observability/logctx"
)

func (s *Server) MoveTripDraftToPublishTx(c *fiber.Ctx) error {
	// OpenTelemetry: child span inside root HTTP span from otelfiber middleware
	tracer := otel.Tracer("trip-api")
	ctx, span := tracer.Start(c.UserContext(), "MoveTripDraftToPublishTxHandler")
	traceID := span.SpanContext().TraceID().String()
	c.Set("X-Request-ID", traceID)
	defer span.End()

	// getting custom logger for logging
	logger := logctx.Logger(ctx).With(
		slog.String("server", "TripServer"),
		slog.String("handler", "MoveTripDraftToPublish"),
		slog.String("trace_id", traceID), // Key field for Grafana
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

	// validation
	if err := s.validator.Struct(&request); err != nil {
		logger.Warn("move trip draft to publish invalid request",
			slog.String("tripId", tripID),
			slog.Any("error", err),
		)
		return HandleError(c, ErrInvalidValidate) // → 400, not unmapped RequestValidationError → 500
	}

	logger = logger.With(
		slog.String("tripId", request.ID),
		slog.String("client_id", request.ClientID.String()),
	)
	ctx = logctx.WithLogger(ctx, logger)
	logger.Debug("move trip to publish")

	resp, err := s.TripService.MoveTripDraftToPublish(ctx, request.ToMoveTripDraftToPublishModel())
	if err != nil {
		logger.Error("move trip to publish failed", slog.Any("error", err))
		return HandleError(c, err)
	}

	if resp.DriverID == uuid.Nil {
		logger.Debug("move trip to publish skipped: no changes detected")
		return c.SendStatus(fiber.StatusNoContent) //http code -204, MUST NOT return a body (JSON)
	}

	logger.Debug("move trip to publish completed")
	return c.Status(fiber.StatusOK).JSON(resp) //200
}
