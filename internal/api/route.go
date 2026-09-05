package api

import (
	"log/slog"
	"os"

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

// RouteInfo — method + path for logging at startup.
type RouteInfo struct {
	Method string
	Path   string
	Note   string
}

// RegisteredRoutes — list of HTTP API that SetupRoutes registers.
func RegisteredRoutes() []RouteInfo {
	return []RouteInfo{
		{Method: "GET", Path: ReadyInfoPath, Note: "liveness / DB ping"},
		{Method: "GET", Path: MetricsInfo, Note: "Prometheus scrape"},
		{Method: "GET", Path: GroupPrefixV2 + TripPath + "/:tripId", Note: "get trip by id (Keycloak)"},
		{Method: "POST", Path: GroupPrefixV2 + TripPath + "/createTripDraft", Note: "create trip draft (Keycloak)"},
		{
			Method: "PATCH",
			Path:   GroupPrefixV2 + TripPath + "/moveTripDraft-ToPublish/:tripId/company/:companyId",
			Note:   "draft → published (Keycloak) -> Kafka -> Notification app",
		},
		{
			Method: "PATCH",
			Path:   GroupPrefixV2 + TripPath + "/moveTripPublished-ToStarted/:tripId/company/:companyId/service/:serviceCode",
			Note:   "published → started + Contract check (Keycloak)",
		},
	}
}

// LogRegisteredRoutes logs available endpoints (same style as Contract Service).
func LogRegisteredRoutes(listenAddr string) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil)).With(slog.String("layer", "http"))
	logger.Info("available endpoints", slog.String("listen", listenAddr))
	for _, r := range RegisteredRoutes() {
		logger.Info("route",
			slog.String("method", r.Method),
			slog.String("url", r.Path),
			slog.String("note", r.Note),
		)
	}
}

// SetupRoutes registers all HTTP routes on the server.
// keycloakAuth — middleware to check the client role for /api/v2/trip/*; nil in apitest (without Keycloak).
func (s *Server) SetupRoutes(app *fiber.App, keycloakAuth fiber.Handler) {
	// === Group API infrastructure ===
	app.Get(ReadyInfoPath, s.GetConnectInfo) // health check
	app.Get(MetricsInfo, adaptor.HTTPHandler(promhttp.HandlerFor(s.registry, promhttp.HandlerOpts{})))

	// === Group API v2 ===
	v2 := app.Group(GroupPrefixV2) //root group - /api/v2

	tripMiddlewares := []fiber.Handler{}
	if keycloakAuth != nil {
		tripMiddlewares = append(tripMiddlewares, keycloakAuth)
	}

	tripGroupV2 := v2.Group(TripPath, tripMiddlewares...) //sub group TripPath = "/trip"
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
		"/moveTripDraft-ToPublish/:tripId/company/:companyId",
		middleware.RequireClientRole(middleware.KeycloakClientID, middleware.KeycloakClientRole),
		s.MoveTripDraftToPublishTx,
	)
	tripGroupV2.Patch(
		"/moveTripPublished-ToStarted/:tripId/company/:companyId/service/:serviceCode",
		middleware.RequireClientRole(middleware.KeycloakClientID, middleware.KeycloakClientRole),
		s.MoveTripPublishedToStarted,
	)
}
