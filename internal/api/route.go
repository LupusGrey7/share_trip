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
// keycloakAuth — middleware для /api/v2/trip/*; nil в apitest (без Keycloak).
func (s *Server) SetupRoutes(app *fiber.App, keycloakAuth fiber.Handler) {
	// === Group API infrastructure ===
	app.Get(ReadyInfoPath, s.GetConnectInfo) // health check
	app.Get("/metrics", adaptor.HTTPHandler(promhttp.HandlerFor(s.registry, promhttp.HandlerOpts{})))

	// === Group API v2 ===
	v2 := app.Group(GroupPrefixV2) //root group - /api/v2

	tripMiddlewares := []fiber.Handler{}
	if keycloakAuth != nil {
		tripMiddlewares = append(tripMiddlewares, keycloakAuth)
	}

	tripGroupV2 := v2.Group(TripPath, tripMiddlewares...) //sub group TripPath = "/ship"
	tripGroupV2.Get(
		"/:tripId",
		middleware.RequireClientRole(middleware.KeycloakClientID, middleware.KeycloakClientRole),
		s.GetTripById,
	)
	tripGroupV2.Post(
		"/createTripDraft",
		middleware.RequireClientRole(middleware.KeycloakClientID, middleware.KeycloakClientRole),
		s.CreateTripDraft,
	)
	tripGroupV2.Patch(
		"/moveTripDraft-ToPublish/:tripId",
		middleware.RequireClientRole(middleware.KeycloakClientID, middleware.KeycloakClientRole),
		s.MoveTripDraftToPublishTx,
	)
}
