package main

import (
	"context"
	"os"
	"path/filepath"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel/trace"
	"job4j.ru/share_trip/internal/observability/metrics"
	"job4j.ru/share_trip/internal/observability/tracing"

	applog "job4j.ru/share_trip/internal/app"
	"job4j.ru/share_trip/internal/middleware"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/joho/godotenv"
	"job4j.ru/share_trip/internal/api"
	appConfigs "job4j.ru/share_trip/internal/app"
	"job4j.ru/share_trip/internal/storage"
)

// init is invoked before main()
// Explicit path to .env. Makefile does `include .env` + `export` and may
// Pass outdated KEYCLOAK_CLIENT_SECRET from OS env; Overload overrides the file.
func init() {
	cwd, err := os.Getwd()
	envFile := ".env"
	if err == nil {
		envFile = filepath.Join(cwd, ".env")
	}
	if loadErr := godotenv.Overload(envFile); loadErr != nil {
		log.Infof("No .env file at %s: %v", envFile, loadErr)
	}
}

func main() {
	ctx := context.Background()
	cfg := appConfigs.ReadDBConfig()

	pool, err := storage.NewPool(ctx, cfg.DSN())
	if err != nil {
		log.Fatal(err)
	}

	defer pool.Close()

	// logging connection to DB
	if pingErr := pool.Ping(ctx); pingErr != nil {
		log.Fatalf("failed to ping database: %v", pingErr)
	}
	log.Info("Connected to database successfully")

	//logger
	logger, logFile, err := applog.NewLogger()
	if err != nil {
		panic(err)
	}
	defer func(logFile *os.File) {
		err := logFile.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(logFile)

	//registry Prometheus
	registry := prometheus.NewRegistry()
	m := metrics.New(registry)

	// init Tracing (OpenTelemetry → otel-collector → Jaeger)
	tp, err := appConfigs.InitTracing(ctx)
	if err != nil {
		log.Error("init tracing failed", "error", err)
		os.Exit(1)
	}
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if shutdownErr := tp.Shutdown(shutdownCtx); shutdownErr != nil {
			log.Error("shutdown tracing failed", "error", shutdownErr)
		}
	}()

	//app fiber
	app := fiber.New(fiber.Config{
		EnablePrintRoutes: true,
	})
	app.Use(tracing.NewFiberMiddleware()) // init Tracing OpenTelemetry
	app.Use(func(c *fiber.Ctx) error {
		ctx := c.UserContext()
		// correlation id = OTel TraceID (X-Request-ID + Locals; в handlers — trace_id в slog)
		traceID := trace.SpanFromContext(ctx).SpanContext().TraceID().String()
		c.Set("X-Request-ID", traceID)
		c.Locals("requestid", traceID)
		return c.Next()
	})
	//add custom logger, before add api and metrics
	app.Use(middleware.Correlation(logger))
	//add metrics middleware
	app.Use(api.NewHTTPMetricsMiddleware(m))

	keycloakCfg := appConfigs.GetKeycloakConfig()
	appConfigs.LogKeycloakConfig(keycloakCfg)

	//build the Server
	appConfigs.BuildServer(app, pool, registry, m, keycloakCfg)

	//listen the app
	err = app.Listen(":8080")
	if err != nil {
		log.Fatal("failed to listen: %v", err)
	}
}
