// scenario: move trip published to started (published → started)

package api

import (
	"log/slog"

	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel"
	"job4j.ru/share_trip/internal/observability/logctx"
)

func (s *Server) MoveTripPublishedToStarted(c *fiber.Ctx) error {
	tracer := otel.Tracer("trip-api")
	ctx, span := tracer.Start(c.UserContext(), "MoveTripPublishedToStartedHandler")
	traceID := span.SpanContext().TraceID().String()
	c.Set("X-Request-ID", traceID)
	defer span.End()

	logger := logctx.Logger(ctx).With(
		slog.String("server", "TripServer"),
		slog.String("handler", "MoveTripPublishedToStarted"),
		slog.String("trace_id", traceID),
	)

	var request MoveTripPublishedToStartedRequest

	// parse / auth / validate — different error classes → different HTTP (см. handler-error-mapping-cheatsheet)
	if err := c.ParamsParser(&request); err != nil {
		logger.Warn("failed to parse path params", slog.Any("error", err))
		return HandleError(c, ErrInvalidValidate) // 400
	}

	if err := s.validator.Struct(&request); err != nil {
		logger.Warn("move trip published to started invalid request",
			slog.String("tripId", request.ID),
			slog.Any("error", err), // details in the log, short class for the client
		)
		return HandleError(c, ErrInvalidValidate) // 400
	}

	logger = logger.With(
		slog.String("trip_id", request.ID),
		slog.String("company_id", request.CompanyID),
		slog.String("service_code", string(request.ServiceCode)),
		slog.String("client_id", request.DriverID.String()),
	)
	ctx = logctx.WithLogger(ctx, logger)
	logger.Debug("move trip published to started request accepted")

	domainReq := toMoveTripPublishedToStartedInput(request)

	resp, err := s.TripService.MoveTripPublishedToStarted(ctx, domainReq)
	if err != nil {
		logger.Error("move trip published to started failed", slog.Any("error", err))
		return HandleError(c, err)
	}

	logger.Debug("move trip published to started completed")
	return c.Status(fiber.StatusOK).JSON(toMoveTripPublishedToStartedResponse(resp))
}
