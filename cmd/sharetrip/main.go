package main

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel/trace"
	"job4j.ru/share_trip/internal/observability/metrics"
	"job4j.ru/share_trip/internal/observability/tracing"

	"job4j.ru/share_trip/internal/middleware"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/joho/godotenv"
	"job4j.ru/share_trip/internal/api"
	appConfigs "job4j.ru/share_trip/internal/app"
	"job4j.ru/share_trip/internal/config"
	"job4j.ru/share_trip/internal/outbox"
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
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg := appConfigs.ReadDBConfig()

	// database
	pool, err := storage.NewPool(ctx, cfg.DSN())
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	if pingErr := pool.Ping(ctx); pingErr != nil {
		log.Fatalf("failed to ping database: %v", pingErr)
	}
	log.Info("Connected to database successfully")

	logger, logFile, err := appConfigs.NewLogger()
	if err != nil {
		panic(err)
	}
	defer func(logFile *os.File) {
		err := logFile.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(logFile)

	// metrics
	registry := prometheus.NewRegistry()
	m := metrics.New(registry)

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

	app := fiber.New()
	app.Use(tracing.NewFiberMiddleware())
	app.Use(func(c *fiber.Ctx) error {
		reqCtx := c.UserContext()
		traceID := trace.SpanFromContext(reqCtx).SpanContext().TraceID().String()
		c.Set("X-Request-ID", traceID)
		c.Locals("requestid", traceID)
		return c.Next()
	})
	app.Use(middleware.Correlation(logger))
	app.Use(api.NewHTTPMetricsMiddleware(m))

	keycloakCfg := appConfigs.GetKeycloakConfig()
	appConfigs.LogKeycloakConfig(keycloakCfg)
	appConfigs.LogContractConfig()

	// producer
	kafkaProducer := appConfigs.NewKafkaProducer()
	defer func() {
		if err := kafkaProducer.Close(); err != nil {
			log.Errorf("failed to close kafka producer: %v", err)
		}
	}()

	// outbox publisher
	outboxPublisher := outbox.NewOutboxPublisher(
		m,
		kafkaProducer,
		pool,
		config.EnvDurationMS("OUTBOX_POLL_INTERVAL_MS", 1000),
		config.EnvInt("OUTBOX_BATCH_SIZE", 50),
	)
	outboxDone := make(chan struct{})
	go func() {
		defer close(outboxDone)
		if runErr := outboxPublisher.Run(ctx); runErr != nil && !errors.Is(runErr, context.Canceled) {
			log.Errorf("outbox publisher stopped: %v", runErr)
		}
	}()

	// server
	appConfigs.BuildServer(app, pool, registry, m, keycloakCfg)

	api.LogRegisteredRoutes(":8080")

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if shutErr := app.ShutdownWithContext(shutdownCtx); shutErr != nil {
			log.Errorf("fiber shutdown: %v", shutErr)
		}
	}()

	err = app.Listen(":8080")
	stop()
	<-outboxDone
	if err != nil {
		log.Fatal("failed to listen: %v", err)
	}
}
