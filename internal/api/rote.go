package api

import (
	"log"

	"github.com/gofiber/fiber/v2"
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
	shipGroup.Get("/:tripId", s.FindByID)
	shipGroup.Post("/create", s.CreateTrip)
}

// RouteV2 - group v2
func (s *Server) RouteV2(route fiber.Router) {
	tripGroupV2 := route.Group(TripPath)

	tripGroupV2.Post("/createTrip", s.CreateTx)
	tripGroupV2.Patch("/moveTripDraft-ToPublish/:tripId", s.MoveTripDraftToPublishTx)
}
