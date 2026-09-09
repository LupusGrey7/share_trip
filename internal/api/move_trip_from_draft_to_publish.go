// Draft → published: publish trip owned by JWT driver.
package api

import (
	"log/slog"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"job4j.ru/share_trip/internal/observability/logctx"
)

func (s *Server) MoveTripFromDraftToPublish(c *fiber.Ctx) error {
	tracer := otel.Tracer("trip-api")
	ctx, span := tracer.Start(c.UserContext(), "MoveTripFromDraftToPublishHandler")
	traceID := span.SpanContext().TraceID().String()
	c.Set("X-Request-ID", traceID)
	defer span.End()

	logger := logctx.Logger(ctx).With(
		slog.String("server", "TripServer"),
		slog.String("handler", "MoveTripFromDraftToPublish"),
		slog.String("trace_id", traceID),
	)

	var request MoveTripFromDraftToPublishRequest

	if err := c.ParamsParser(&request); err != nil {
		logger.Warn("failed to parse path params", slog.Any("error", err))
		return HandleError(c, ErrInvalidValidate)
	}

	driverID, err := getDriverIDFromContext(c)
	if err != nil {
		logger.Error("failed to get driver ID from context", slog.Any("error", err))
		return HandleError(c, err)
	}
	request.DriverID = driverID

	if err := s.validator.Struct(&request); err != nil {
		logger.Warn("move trip from draft to publish invalid request",
			slog.String("tripId", request.ID),
			slog.String("companyId", request.CompanyID),
			slog.Any("error", err),
		)
		return HandleError(c, ErrInvalidValidate)
	}

	logger = logger.With(
		slog.String("tripId", request.ID),
		slog.String("client_id", request.DriverID.String()),
	)
	ctx = logctx.WithLogger(ctx, logger)
	logger.Debug("move trip from draft to publish")

	resp, err := s.TripService.MoveTripFromDraftToPublish(ctx, toMoveTripFromDraftToPublishInput(&request))
	if err != nil {
		logger.Error("move trip from draft to publish failed", slog.Any("error", err))
		return HandleError(c, err)
	}

	if resp.DriverID == uuid.Nil {
		logger.Debug("move trip from draft to publish skipped: no changes detected")
		return c.SendStatus(fiber.StatusNoContent)
	}

	logger.Debug("move trip from draft to publish completed")
	return c.Status(fiber.StatusOK).JSON(toMoveTripFromDraftToPublishResponse(resp))
}
