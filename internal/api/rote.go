package api

import (
	"github.com/gofiber/adaptor/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	GroupPrefixV1 = "/api/v1"
	GroupPrefixV2 = "/api/v2"
	ReadyInfoPath = "/ready"
	TripPath      = "/trip"
	MetricsInfo   = "/metrics"
)

// SetupRoutes registers all HTTP routes on the server.
// Called once at application startup.
func (s *Server) SetupRoutes(app *fiber.App) {
	// === Group API v1 ===
	v1 := app.Group(GroupPrefixV1)
	v1.Get(ReadyInfoPath, s.GetConnectInfo) // health check

	// === Group API v2 ===
	v2 := app.Group(GroupPrefixV2)    //root group
	tripGroupV2 := v2.Group(TripPath) //sub group TripPath = "/ship"
	tripGroupV2.Get("/:tripId", s.GetTripById)
	tripGroupV2.Post("/createTripDraft", s.CreateTripDraft)
	tripGroupV2.Patch("/moveTripDraft-ToPublish/:tripId", s.MoveTripDraftToPublishTx)

	// === Group API v3 ===
	// === Prometheus metrics (without prefix, at the root) ===
	// Prometheus (deploy/prometheus/prometheus.yml) scrapes :8080/metrics — must be on the app root.
	app.Get("/metrics", adaptor.HTTPHandler(promhttp.HandlerFor(s.registry, promhttp.HandlerOpts{})))
}
