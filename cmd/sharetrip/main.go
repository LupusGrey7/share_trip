package main

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/prometheus/client_golang/prometheus"
	clientContract "job4j.ru/share_trip/internal/clients/http/contract"
	"job4j.ru/share_trip/internal/observability/metrics"
	"job4j.ru/share_trip/internal/observability/tracing"
	"job4j.ru/share_trip/internal/trip/service"
	"job4j.ru/share_trip/internal/trip/usecase"

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

const (
	envFileName = ".env"

	// outbox config
	outboxPollIntervalMS = 1000
	outboxBatchSize      = 50

	// tracing
	otelServiceNameDef      = "share-trip"
	otelServiceVersionDef   = "1.0.0"
	otelEnvironmentDef      = "local"
	otelExporterEndpointDef = "localhost:4319"
)

// init is invoked before main()
// Explicit path to .env. Makefile does `include .env` + `export` and may
// Pass outdated KEYCLOAK_CLIENT_SECRET from OS env; Overload overrides the file.
func init() {
	cwd, err := os.Getwd()
	envFile := envFileName
	if err == nil {
		envFile = filepath.Join(cwd, envFileName)
	}
	if loadErr := godotenv.Overload(envFile); loadErr != nil {
		log.Infof("No .env file at %s: %v", envFile, loadErr)
	}
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// read config
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
	metric := metrics.New(registry)

	// tracing
	tp, err := tracing.NewProvider(ctx, tracing.Config{
		ServiceName:    config.Env("OTEL_SERVICE_NAME", otelServiceNameDef),
		ServiceVersion: config.Env("OTEL_SERVICE_VERSION", otelServiceVersionDef),
		Environment:    config.Env("OTEL_ENVIRONMENT", otelEnvironmentDef),
		Endpoint:       config.Env("OTEL_EXPORTER_ENDPOINT", otelExporterEndpointDef),
	})
	if err != nil {
		log.Fatalf("init tracing failed: %v", err)
	}
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if shutdownErr := tp.Shutdown(shutdownCtx); shutdownErr != nil {
			log.Errorf("shutdown tracing failed: %v", shutdownErr)
		}
	}()

	// fiber app
	app := fiber.New()
	app.Use(tracing.NewFiberMiddleware())
	app.Use(middleware.TraceIDHeader())
	app.Use(middleware.Correlation(logger))
	app.Use(api.NewHTTPMetricsMiddleware(metric))

	// keycloak config
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
		metric,
		kafkaProducer,
		pool,
		config.EnvDurationMS("OUTBOX_POLL_INTERVAL_MS", outboxPollIntervalMS),
		config.EnvInt("OUTBOX_BATCH_SIZE", outboxBatchSize),
	)
	outboxDone := make(chan struct{})
	go func() {
		defer close(outboxDone)
		if runErr := outboxPublisher.Run(ctx); runErr != nil && !errors.Is(runErr, context.Canceled) {
			log.Errorf("outbox publisher stopped: %v", runErr)
		}
	}()

	// server (build server and setup routes)
	validate := validator.New(validator.WithRequiredStructEnabled())

	contractClient := clientContract.NewClient(config.ContractServiceURL())

	repo := storage.NewRepoPg(pool)
	repoTrip := storage.NewTripRepository(metric, pool)
	outboxRepo := storage.NewOutboxEventRepository(metric)

	infoUseCase := usecase.NewInfoUseCase()
	tripUseCase := usecase.NewTripUseCase()

	infoService := service.NewInfoService(infoUseCase, repo)
	tripService := service.NewTripService(metric, pool, repoTrip, outboxRepo, tripUseCase, contractClient)

	server := api.NewServer(registry, validate, infoService, tripService)

	keycloakAuth := middleware.KeycloakRefreshTokenMiddleware(keycloakCfg)
	server.SetupRoutes(app, keycloakAuth)

	// app print registered routes
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
		log.Fatalf("failed to listen: %v", err)
	}
}
