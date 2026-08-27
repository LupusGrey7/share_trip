// scenario: move trip published to started (published → started)
package api

import (
	"log/slog"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
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

	// parse / auth / validate — разные классы ошибок → разные HTTP (см. handler-error-mapping-cheatsheet)
	if err := c.ParamsParser(&request); err != nil {
		logger.Warn("failed to parse path params", slog.Any("error", err))
		return HandleError(c, ErrInvalidValidate) // 400
	}

	driverID, err := getDriverIDFromContext(c)
	if err != nil {
		logger.Error("failed to get driver ID from context", slog.Any("error", err))
		return HandleError(c, err) // 401 / 403 / 502 — НЕ подменять на ErrInvalidValidate
	}
	request.DriverID = driverID

	if err := s.validator.Struct(&request); err != nil {
		logger.Warn("move trip published to started invalid request",
			slog.String("tripId", request.ID),
			slog.Any("error", err), // детали — в лог, клиенту короткий класс
		)
		return HandleError(c, ErrInvalidValidate) // 400
	}

	logger = logger.With(
		slog.String("trip_id", request.ID),
		slog.String("company_id", request.CompanyID),
		slog.String("service_code", string(request.ServiceCode)),
		slog.String("client_id", driverID.String()),
	)
	ctx = logctx.WithLogger(ctx, logger)
	logger.Debug("move trip published to started request accepted")

	domainReq := toMoveTripPublishedToStartedInput(request, driverID)

	resp, err := s.TripService.MoveTripPublishedToStarted(ctx, domainReq)
	if err != nil {
		logger.Error("move trip published to started failed", slog.Any("error", err))
		return HandleError(c, err)
	}

	if resp.DriverID == uuid.Nil {
		logger.Debug("move trip published to started skipped: already started")
		return c.SendStatus(fiber.StatusNoContent)
	}

	logger.Debug("move trip published to started completed",
		slog.String("status", string(resp.Status)),
	)
	return c.Status(fiber.StatusOK).JSON(toMoveTripPublishedToStartedResponse(resp))
}

func getDriverIDFromContext(c *fiber.Ctx) (uuid.UUID, error) {
	claims, err := GetClaimsFromContext(c)
	if err != nil {
		return uuid.Nil, err
	}
	clientID, err := ClientIDFromClaims(claims)
	if err != nil {
		return uuid.Nil, err
	}
	return clientID, nil
}
