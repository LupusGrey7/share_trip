//api scenario - http handler for getting trip by ID.

package api

import (
	"log/slog"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"

	"job4j.ru/share_trip/internal/domain/trip/model"
	"job4j.ru/share_trip/internal/observability/logctx"

	"github.com/gofiber/fiber/v2"
	"job4j.ru/share_trip/internal/domain/errs"
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

	tripID := c.Params("tripId")
	if tripID == "" {
		return fiber.NewError(fiber.StatusBadRequest, invalidIdParamFormat)
	}
	//check if the token is valid and the role is client
	_, err := GetClaimsFromContext(c)
	if err != nil {
		logger.Error("failed to get claims from context", slog.Any("error", err))
		return HandleError(c, err)
	}

	request := GetByIDRequestModel{ID: tripID}

	if err := s.validator.Struct(request); err != nil {
		return errs.RequestValidationError{Message: err.Error()}
	}
	//tracing
	span.SetAttributes(
		attribute.String("trip_id", tripID),
	)
	// logging at the component boundary
	logger.Debug("get trip by id", slog.String("tripId", tripID))

	resp, err := s.TripService.GetTripByID(ctx, model.GetByIDModelRequest{ID: request.ID})
	if err != nil {
		logger.Error("get trip by id failed", slog.Any("error", err))
		return HandleError(c, err)
	}

	logger.Debug("get trip by id completed", slog.String("trip_id", request.ID))
	return c.Status(fiber.StatusOK).JSON(resp)
}
