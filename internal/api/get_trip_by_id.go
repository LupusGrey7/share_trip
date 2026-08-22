//api scenario - http handler for getting trip by ID.

package api

import (
	"log/slog"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"

	"job4j.ru/share_trip/internal/observability/logctx"

	"github.com/gofiber/fiber/v2"
)

func (s *Server) GetTripById(c *fiber.Ctx) error {
	// OpenTelemetry: child span inside root HTTP span from otelfiber middleware
	tracer := otel.Tracer("trip-api")
	ctx, span := tracer.Start(c.UserContext(), "GetTripByIDHandler")
	traceID := span.SpanContext().TraceID().String()
	c.Set("trace-id", traceID)
	defer span.End()

	logger := logctx.Logger(ctx).With(
		slog.String("server", "TripServer"),
		slog.String("handler", "GetTripByID"),
		slog.String("trace_id", traceID), // Key field for Grafana
	)

	var request GetTripByIDRequest
	if err := c.ParamsParser(&request); err != nil {
		logger.Warn("get trip by id failed: invalid request", slog.String("error", invalidIdParamFormat))
		return ErrResponse(c, fiber.StatusBadRequest, invalidIdParamFormat)
	}

	if err := s.validator.Struct(&request); err != nil {
		logger.Warn("get trip by id failed: invalid request", slog.Any("error", err))
		return HandleError(c, ErrInvalidValidate) // → 400, not unmapped RequestValidationError → 500
	}

	// identity from Keycloak (role already checked on route; helper = defense in depth)
	claims, err := GetClaimsFromContext(c)
	if err != nil {
		logger.Error("failed to get claims from context", slog.Any("error", err))
		return HandleError(c, err)
	}
	clientID, err := ClientIDFromClaims(claims)
	if err != nil {
		logger.Error("failed to parse client ID from token subject", slog.Any("error", err))
		return HandleError(c, err)
	}

	logger = logger.With(slog.String("client_id", clientID.String()))
	ctx = logctx.WithLogger(ctx, logger)

	//tracing
	span.SetAttributes(
		attribute.String("trip_id", request.ID),
	)
	// logging at the component boundary
	logger.Debug("get trip by id", slog.String("tripId", request.ID))

	resp, err := s.TripService.GetTripByID(ctx, toGetByIDInput(&request))
	if err != nil {
		logger.Error("get trip by id failed", slog.Any("error", err))
		return HandleError(c, err)
	}

	// ownership: only the driver from Keycloak sub may read this trip (IDOR guard)
	if resp.DriverID != clientID {
		logger.Error("get trip forbidden: caller is not trip owner",
			slog.String("trip_driver_id", resp.DriverID.String()),
			slog.String("token_sub", clientID.String()),
		)
		return HandleError(c, ErrForbiddenIDMismatch)
	}

	out := toGetTripByIDResponse(resp)
	logger.Debug("get trip by id completed", slog.String("trip_id", request.ID))
	return c.Status(fiber.StatusOK).JSON(out)
}
