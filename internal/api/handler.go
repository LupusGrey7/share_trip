package api

import (
	"context"
	"errors"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"job4j.ru/share_trip/internal/api/apierr"
	"job4j.ru/share_trip/internal/domain/errs"

	"job4j.ru/share_trip/internal/domain/trip"
	"job4j.ru/share_trip/internal/service"
)

const (
	invalidRequestFormat = "Invalid request format"
	validationFailed     = "validate name error"
	internalServerError  = "Internal server error"
	invalidIdParamFormat = "id param is required"
)

type Service interface {
	CreateTrip(ctx context.Context, tr trip.CreateTripCommand) (trip.Response, error)
	GetById(context.Context, string) (trip.Response, error)
}

type Handler struct {
	createSvc service.CommandTripService
	getSvc    service.QueryTripService
}

func NewHandler(
	cSvc *service.CommandTripService,
	gSvs *service.QueryTripService,
) *Handler {
	return &Handler{
		createSvc: *cSvc,
		getSvc:    *gSvs}
}

func (h *Handler) FindByID(c *fiber.Ctx) error {
	ctx := c.UserContext()

	id := c.Params("tripId")
	if id == "" {
		return fiber.NewError(fiber.StatusBadRequest, invalidIdParamFormat)
	}

	resp, err := h.getSvc.GetById(ctx, id)
	if err != nil {
		log.Error("error when FindById trip is: ", err)

		switch {
		case errors.As(err, &errs.RequestValidationError{}):
			return apierr.ErrResponse(c, fiber.StatusBadRequest, err.Error())

		default:
			return apierr.ErrResponse(c, fiber.StatusInternalServerError, internalServerError)
		}
	}
	return c.Status(fiber.StatusOK).JSON(resp)
}

func (h *Handler) Create(c *fiber.Ctx) error {
	ctx := c.UserContext()
	var request trip.CreateTripCommand

	// Парсим тело запроса
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			fiber.Map{
				"error": "Cannot parse JSON",
			})
	}

	resp, err := h.createSvc.CreateTrip(ctx, request)
	if err != nil {
		log.Error("error create is: ", err)
		switch {
		case errors.As(err, &errs.RequestValidationError{}):
			return apierr.ErrResponse(c, fiber.StatusBadRequest, err.Error())

		default:
			return apierr.ErrResponse(c, fiber.StatusInternalServerError, internalServerError)
		}
	}
	return c.Status(fiber.StatusCreated).JSON(resp)
}
