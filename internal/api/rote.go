package api

import (
	"github.com/gofiber/adaptor/v2"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"

	"github.com/gofiber/fiber/v2"
)

const (
	InfoPath    = "/ready"
	TripPath    = "/trip"
	MetricsInfo = "/metrics"
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

// RouteV3 - group v3
func (s *Server) RouteV3(route fiber.Router) {
	route.Get(MetricsInfo, adaptor.HTTPHandler(promhttp.HandlerFor(s.registry, promhttp.HandlerOpts{})))
}
