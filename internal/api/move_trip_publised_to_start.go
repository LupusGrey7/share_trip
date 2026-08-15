// scenario: move trip published to started
package api

import (
	"log/slog"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"job4j.ru/share_trip/internal/api/apierr"
	"job4j.ru/share_trip/internal/observability/logctx"
)

func (s *Server) MoveTripPublishedToStarted(c *fiber.Ctx) error {
	// OpenTelemetry: child span inside root HTTP span from otelfiber middleware
	tracer := otel.Tracer("trip-api")
	ctx, span := tracer.Start(c.UserContext(), "MoveTripPublishedToStartedHandler")
	traceID := span.SpanContext().TraceID().String()
	c.Set("X-Request-ID", traceID)
	defer span.End()

	// getting custom logger for logging
	logger := logctx.Logger(ctx).With(
		slog.String("server", "TripServer"),
		slog.String("handler", "MoveTripPublishedToStarted"),
		slog.String("trace_id", traceID), // Key field for Grafana
	)

	var request MoveTripPublishedToStartedRequestModel

	// request param
	tripID := c.Params("tripId")
	if tripID == "" {
		return fiber.NewError(fiber.StatusBadRequest, invalidIdParamFormat)
	}
	request.CompanyID = c.Params("companyId")
	if request.CompanyID == "" {
		return fiber.NewError(fiber.StatusBadRequest, invalidIdParamFormat)
	}
	request.ServiceCode = ServiceCodeEnum(c.Params("serviceCode"))
	if request.ServiceCode == "" {
		return fiber.NewError(fiber.StatusBadRequest, invalidIdParamFormat)
	}

	// validation
	if err := s.validator.Struct(&request); err != nil {
		logger.Warn("move trip published to started invalid request",
			slog.String("tripId", tripID),
			slog.Any("error", err),
		)
		return HandleError(c, apierr.ErrInvalidValidate) // → 400, not unmapped RequestValidationError → 500
	}

	logger = logger.With(
		slog.String("tripId", request.ID),
		slog.String("company_id", request.CompanyID.String()),
		slog.String("service_code", string(request.ServiceCode)),
	)
	ctx = logctx.WithLogger(ctx, logger)
	logger.Debug("move trip to started")

	resp, err := s.TripService.MoveTripPublishedToStartedTx(ctx, request.ToMoveTripPublishedToStartedModel())
	if err != nil {
		logger.Error("move trip to started failed", slog.Any("error", err))
		return HandleError(c, err)
	}

	if resp.DriverID == uuid.Nil {
		logger.Debug("move trip to started skipped: no changes detected")
		return c.SendStatus(fiber.StatusNoContent) //http code -204, MUST NOT return a body (JSON)
	}

	logger.Debug("move trip to started completed")
	return c.Status(fiber.StatusOK).JSON(resp) //200
}
