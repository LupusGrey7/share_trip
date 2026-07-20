//API scenario - trips from the draft state (in a DB transaction)

package api

import (
	"log/slog"

	"fmt"

	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel"
	"job4j.ru/share_trip/internal/api/apierr"
	"job4j.ru/share_trip/internal/domain/errs"
	"job4j.ru/share_trip/internal/domain/trip/model"
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
		logger.Warn(
			"create trip draft failed: invalid request",
			slog.Any("error", err),
		)
		return errs.RequestValidationError{Message: err.Error()}
	}

	logger = logger.With(slog.String("client_id", request.DriverID.String()))
	ctx = logctx.WithLogger(ctx, logger)
	logger.Debug("create trip draft request accepted")

	// client identity from Keycloak JWT (sub) — must match body driverId (IDOR guard)
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

	logger.Info("clientID", slog.String("client_id", clientID.String())) //FIXME for debug
	logger.Info("tripID", slog.String("trip_id", claims.Subject))        //FIXME for debug
	fmt.Println("clientID", clientID.String())                           //FIXME for debug
	fmt.Println("tripID", claims.Subject)                                //FIXME for debug

	if request.DriverID != clientID {
		logger.Error("client ID mismatch",
			slog.String("request_driver_id", request.DriverID.String()),
			slog.String("token_sub", clientID.String()),
		)
		return HandleError(c, apierr.ErrForbiddenIDMismatch)
	}

	resp, err := s.TripService.CreateTripDraft(
		ctx,
		model.CreateTripRequestModel{
			DriverID:       clientID, // source of truth = Keycloak sub (same as body after check)
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

	logger.Debug("create trip draft completed")
	return c.Status(fiber.StatusCreated).JSON(resp)
}
