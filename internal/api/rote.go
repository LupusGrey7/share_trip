package api

import (
	"log"

	"github.com/gofiber/fiber/v2"
)

const (
	InfoPath = "/ready"
	TripPath = "/trip"
)

// Route - group v1
func (s *Server) Route(route fiber.Router) {
	log.Println("Server listening on :8080")
	route.Get(InfoPath, s.GetConnectInfo)

}

// RouteV2 - group v2
func (s *Server) RouteV2(route fiber.Router) {
	tripGroupV2 := route.Group(TripPath)

	tripGroupV2.Get("/:tripId", s.GetTripById)
	tripGroupV2.Post("/createTripDraft", s.CreateTripDraft)
	tripGroupV2.Patch("/moveTripDraft-ToPublish/:tripId", s.MoveTripDraftToPublishTx)
}
