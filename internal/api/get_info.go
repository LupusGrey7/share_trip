package api

import (
	"github.com/gofiber/fiber/v2"
)

type GetInfoResponse struct {
	Status string `json:"status"`
}

func (s *Server) GetConnectInfo(ctx *fiber.Ctx) error {
	res, err := s.InfoService.GetDBInfo(ctx.UserContext())
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "internal server error")
	}

	return ctx.Status(fiber.StatusOK).JSON(GetInfoResponse{Status: res})
}
