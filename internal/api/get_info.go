package api

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
)

type GetInfoResponse struct {
	Status string `json:"status"`
}

func (s *Server) GetConnectInfo(ctx *fiber.Ctx) error {
	res, err := s.Service.GetDBInfo(ctx.UserContext())
	if err != nil {
		log.Errorw("s.Repository.List", err)
		return fiber.NewError(fiber.StatusInternalServerError, "internal server error")
	}
	return ctx.Status(fiber.StatusOK).JSON(GetInfoResponse{Status: res})
}
