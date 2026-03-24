package api

import (
	"github.com/gofiber/fiber/v2"
	"log"
)

const (
	InfoPath = "/ready"
	TripPath = "/trip"
)

func (s *Server) Route(route fiber.Router) {
	log.Println("Server listening on :8080")
	route.Get(InfoPath, s.GetConnectInfo)

	//group
	shipGroup := route.Group(TripPath)
	shipGroup.Get("/:tripId", s.TripHandler.FindByID)
	shipGroup.Post("/create", s.TripHandler.Create)
}
