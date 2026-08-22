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

	tripID := c.Params("tripId")
	if tripID == "" {
		return fiber.NewError(fiber.StatusBadRequest, invalidIdParamFormat)
	}
	request.ID = tripID

	request.CompanyID = c.Params("companyId")
	if request.CompanyID == "" {
		return fiber.NewError(fiber.StatusBadRequest, invalidIdParamFormat)
	}

	request.ServiceCode = ServiceCodeEnum(c.Params("serviceCode"))
	if request.ServiceCode == "" {
		return fiber.NewError(fiber.StatusBadRequest, invalidIdParamFormat)
	}

	if err := s.validator.Struct(&request); err != nil {
		logger.Warn("move trip published to started invalid request",
			slog.String("tripId", tripID),
			slog.Any("error", err),
		)
		return HandleError(c, ErrInvalidValidate)
	}

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

	logger = logger.With(
		slog.String("trip_id", request.ID),
		slog.String("company_id", request.CompanyID),
		slog.String("service_code", string(request.ServiceCode)),
		slog.String("client_id", clientID.String()),
	)
	ctx = logctx.WithLogger(ctx, logger)
	logger.Debug("move trip published to started request accepted")

	domainReq := toMoveTripPublishedToStartedInput(request, clientID)

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
