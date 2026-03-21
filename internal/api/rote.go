package api

import (
	"github.com/gofiber/fiber/v2"
	"log"
)

func (s *Server) Route(route fiber.Router) {
	log.Println("Server listening on :8080")
	route.Get("/ready", s.GetConnectInfo)
}
