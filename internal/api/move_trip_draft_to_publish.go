//api сценарий - перевода поездки из состояния draft в published.

package api

import (
	"log/slog"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/google/uuid"
	"job4j.ru/share_trip/internal/api/apierr"
	"job4j.ru/share_trip/internal/domain/errs"
	"job4j.ru/share_trip/internal/observability/logctx"
)

func (s *Server) MoveTripDraftToPublishTx(c *fiber.Ctx) error {
	ctx := c.UserContext()
	logger := logctx.Logger(ctx).With(
		slog.String("server", "TripServer"),
		slog.String("handler", "CreateTrip"),
	)
	// Достаем ID, который сгенерировал requestid.New()
	traceID := c.GetRespHeader(requestid.ConfigDefault.Header)
	var request MoveTripDraftToPublishModelRequest

	id := c.Params("tripId")
	if id == "" {
		return fiber.NewError(fiber.StatusBadRequest, invalidIdParamFormat)
	}
	// Parse request body
	if err := c.BodyParser(&request); err != nil {
		logger.Warn(
			"create trip failed: invalid JSON body",
			slog.Any("error", err),
		)

		return c.Status(fiber.StatusBadRequest).JSON(
			fiber.Map{
				"error": invalidParseJson,
			})
	}
	request.ID = id
	//--validation
	if err := s.validator.Struct(&request); err != nil {
		return errs.RequestValidationError{Message: err.Error()}
	}

	uuID, err := uuid.Parse(id)
	if err != nil {
		return apierr.ErrResponse(c, fiber.StatusInternalServerError, apierr.InternalServerError)
	}
	// логирование на границе компонента.
	log.Infof("move trip to publish ID: %v with traceID: %s ", uuID, traceID)

	resp, err := s.TripService.MoveTripDraftToPublish(ctx, request.ToRequest(uuID))
	if err != nil {
		return HandleError(c, err)
	}

	if resp.DriverID == uuid.Nil {
		return c.Status(fiber.StatusNoContent).JSON(resp) //204
	}
	logger.Info(
		"create trip completed",
		slog.String("trip_id", resp.ID.String()),
	)

	return c.Status(fiber.StatusOK).JSON(resp) //200
}
