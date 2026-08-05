//API scenario - trips from the draft state (in a DB transaction)

package api

import (
	"log/slog"

	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel"
	"job4j.ru/share_trip/internal/api/apierr"
	"job4j.ru/share_trip/internal/observability/logctx"
)

func (s *Server) CreateTripDraft(c *fiber.Ctx) error {
	// OpenTelemetry: child span inside root HTTP span from otelfiber middleware
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
		logger.Warn("create trip draft failed: invalid request", slog.Any("error", err))
		return HandleError(c, apierr.ErrInvalidValidate) // → 400, not unmapped RequestValidationError → 500
	}

	// Identity: client IS the driver. No body.driverId — source of truth = Keycloak sub.
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
	logger.Debug("create trip draft request accepted")

	domainReq := request.ToCreateTripRequestModel()
	domainReq.DriverID = clientID

	resp, err := s.TripService.CreateTripDraft(ctx, domainReq)
	if err != nil {
		logger.Error("create trip draft failed", slog.Any("error", err))
		return HandleError(c, err)
	}

	logger.Debug("create trip draft completed")
	return c.Status(fiber.StatusCreated).JSON(resp)
}
