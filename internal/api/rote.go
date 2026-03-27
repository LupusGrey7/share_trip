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
	tripGroup := route.Group(TripPath)
	tripGroup.Get("/:tripId", s.TripHandler.FindByID)
	tripGroup.Post("/create", s.TripHandler.Create)
}

// RouteV2 - group v2
func (s *Server) RouteV2(route fiber.Router) {
	tripGroupV2 := route.Group(TripPath)
	tripGroupV2.Post("/create-tx", s.TripHandler.CreateTx)
	tripGroupV2.Patch("/:tripId", s.TripHandler.UpdateTripDraftToPublishTx)

}
