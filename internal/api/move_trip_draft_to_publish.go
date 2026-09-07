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

	var request MoveTripDraftToPublishRequest

	// parse / auth / validate — different error classes → different HTTP (см. handler-error-mapping-cheatsheet)
	if err := c.ParamsParser(&request); err != nil {
		logger.Warn("failed to parse path params", slog.Any("error", err))
		return HandleError(c, ErrInvalidValidate) // 400
	}

	// Identity from Keycloak JWT sub — not from body (lead requirement / IDOR-safe)
	driverID, err := getDriverIDFromContext(c)
	if err != nil {
		logger.Error("failed to get driver ID from context", slog.Any("error", err))
		return HandleError(c, err) // 401 / 403 / 502
	}
	request.DriverID = driverID

	if err := s.validator.Struct(&request); err != nil {
		logger.Warn("move trip draft to publish invalid request",
			slog.String("tripId", request.ID),
			slog.String("companyId", request.CompanyID),
			slog.Any("error", err),
		)
		return HandleError(c, ErrInvalidValidate) // → 400
	}

	logger = logger.With(
		slog.String("tripId", request.ID),
		slog.String("client_id", request.DriverID.String()),
	)
	ctx = logctx.WithLogger(ctx, logger)
	logger.Debug("move trip to publish")

	resp, err := s.TripService.MoveTripDraftToPublish(ctx, toMoveTripDraftToPublishInput(&request))
	if err != nil {
		logger.Error("move trip to publish failed", slog.Any("error", err))
		return HandleError(c, err)
	}

	if resp.DriverID == uuid.Nil {
		logger.Debug("move trip to publish skipped: no changes detected")
		return c.SendStatus(fiber.StatusNoContent) //http code -204, MUST NOT return a body (JSON)
	}

	logger.Debug("move trip to publish completed")
	return c.Status(fiber.StatusOK).JSON(toMoveTripDraftToPublishResponse(resp)) //200
}
