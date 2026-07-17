package api

import (
	"github.com/gofiber/adaptor/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"job4j.ru/share_trip/internal/middleware"
)

const (
	GroupPrefixV2 = "/api/v2"
	ReadyInfoPath = "/ready"
	TripPath      = "/trip"
	MetricsInfo   = "/metrics"
)

// SetupRoutes registers all HTTP routes on the server.
// Called once at application startup.
func (s *Server) SetupRoutes(app *fiber.App) {
	// === Group API infrastructure ===
	app.Get(ReadyInfoPath, s.GetConnectInfo) // health check
	// === Prometheus metrics (without prefix, at the root) ===
	// Prometheus (deploy/prometheus/prometheus.yml) scrapes :8080/metrics — must be on the app root.
	app.Get("/metrics", adaptor.HTTPHandler(promhttp.HandlerFor(s.registry, promhttp.HandlerOpts{})))

	// === Group API v2 ===
	v2 := app.Group(GroupPrefixV2)    //root group
	tripGroupV2 := v2.Group(TripPath) //sub group TripPath = "/ship"
	tripGroupV2.Get(
		"/:tripId",
		middleware.RequireClientRole("sharetrip-api", "get-trip-by-id"), // require the client role to get the trip by id
		s.GetTripById,
	)
	tripGroupV2.Post(
		"/createTripDraft",
		middleware.RequireClientRole("sharetrip-api", "create-trip-draft"), // require the client role to create the trip draft
		s.CreateTripDraft,
	)
	tripGroupV2.Patch(
		"/moveTripDraft-ToPublish/:tripId",
		middleware.RequireClientRole("sharetrip-api", "move-trip-draft-to-publish"), // require the client role to move the trip draft to publish
		s.MoveTripDraftToPublishTx,
	)
}
