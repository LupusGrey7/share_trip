//api http handler for getting trip by ID

package api

import (
	"log/slog"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"

	"job4j.ru/share_trip/internal/domain/trip/model"
	"job4j.ru/share_trip/internal/observability/logctx"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"job4j.ru/share_trip/internal/domain/errs"
)

func (s *Server) GetTripById(c *fiber.Ctx) error {
	// OpenTelemetry with Jaeger
	tracer := otel.Tracer("trip-api") //
	ctx, span := tracer.Start(c.UserContext(), "GetTripByIDHandler")
	defer span.End()
	c.Set("trace-id", span.SpanContext().TraceID().String())

	// We get the ID that was generated requestid.New()
	traceID := c.GetRespHeader(requestid.ConfigDefault.Header)

	//getting custom logger context
	logger := logctx.Logger(ctx).With(
		slog.String("server", "TripServer"),
		slog.String("handler", "CreateTrip"),
		slog.String("traceID", traceID),
	)

	tripID := c.Params("tripId")
	if tripID == "" {
		return fiber.NewError(fiber.StatusBadRequest, invalidIdParamFormat)
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
	logger.Debug(
		"get Trip By ID",
		slog.String("tripId", tripID),
	)

	resp, err := s.TripService.GetTripByID(ctx, model.GetByIDModelRequest{ID: request.ID})
	if err != nil {
		logger.Error(
			"get Trip By ID failed",
			slog.Any("error", err),
		)
		return HandleError(c, err)
	}

	logger.Debug(
		"get Trip By ID completed",
		slog.String("trip_id", request.ID),
	)
	return c.Status(fiber.StatusOK).JSON(resp)
}
