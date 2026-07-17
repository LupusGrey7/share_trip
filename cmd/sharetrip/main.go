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

	"github.com/jackc/pgx/v5/pgxpool"
	applog "job4j.ru/share_trip/internal/app"
	"job4j.ru/share_trip/internal/domain/trip/usecase"
	"job4j.ru/share_trip/internal/middleware"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/joho/godotenv"
	"job4j.ru/share_trip/configs"
	"job4j.ru/share_trip/internal/api"
	"job4j.ru/share_trip/internal/repository"
	"job4j.ru/share_trip/internal/service"
	"job4j.ru/share_trip/internal/storage"
)

// init is invoked before main()
func init() {
	// Явный путь к .env. Makefile делает `include .env` + `export` и может
	// прокинуть устаревший KEYCLOAK_CLIENT_SECRET из OS env; Overload перекрывает файлом.
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

	cfg := readCfg()

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
	tp, err := initTracing(ctx)
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

	//app
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

	keycloakCfg := middleware.KeycloakConfig{
		Issuer:       configs.Env("KEYCLOAK_ISSUER", "http://localhost:8087/realms/sharetrip"),
		ClientID:     configs.Env("KEYCLOAK_CLIENT_ID", "sharetrip-api"),
		ClientSecret: configs.Env("KEYCLOAK_CLIENT_SECRET", ""),
	}
	cwd, _ := os.Getwd()
	log.Infof("keycloak: issuer=%s client_id=%s secret_len=%d cwd=%s",
		keycloakCfg.Issuer, keycloakCfg.ClientID, len(keycloakCfg.ClientSecret), cwd)
	if keycloakCfg.ClientSecret == "" || keycloakCfg.ClientSecret == "secret" {
		log.Error("KEYCLOAK_CLIENT_SECRET empty or placeholder 'secret' — save real secret in .env, then: " +
			"Remove-Item Env:KEYCLOAK_CLIENT_SECRET; make run")
	}

	//build Server
	build(app, pool, registry, m, keycloakCfg)

	//listen app
	err = app.Listen(":8080")
	if err != nil {
		log.Fatal("failed to listen: %v", err)
	}
}

// build - build server
func build(
	app *fiber.App,
	pool *pgxpool.Pool,
	registry *prometheus.Registry,
	m *metrics.Metrics,
	keycloakCfg middleware.KeycloakConfig,
) {
	// Initialize the validator instance
	validate := validator.New(validator.WithRequiredStructEnabled())

	repo := repository.NewRepoPg(pool)
	repoTrip := repository.NewTripRepository(m, pool)
	outboxRepo := repository.NewOutboxEventRepository()

	infoUseCase := usecase.NewInfoUseCase()
	tripUseCase := usecase.NewTripUseCase()

	infoService := service.NewInfoService(infoUseCase, repo)
	tripService := service.NewTripService(m, pool, repoTrip, outboxRepo, tripUseCase)

	server := api.NewServer(registry, validate, infoService, tripService)

	keycloakAuth := middleware.KeycloakRefreshTokenMiddleware(keycloakCfg)
	server.SetupRoutes(app, keycloakAuth)
}

func readCfg() storage.Config {
	return storage.Config{
		Host:     configs.Env("DB_HOST", "localhost"),
		Port:     configs.EnvInt("DB_PORT", 6543),
		User:     configs.Env("DB_USER", "postgres"),
		Password: configs.Env("DB_PASSWORD", "password"),
		DBName:   configs.Env("DB_NAME", "share_trip"),
		SSLMode:  configs.Env("DB_SSLMODE", "disable"),
	}
}

func initTracing(ctx context.Context) (*tracing.TracerProvider, error) {
	return tracing.NewProvider(ctx, tracing.Config{
		ServiceName:    configs.Env("OTEL_SERVICE_NAME", "share-trip"),
		ServiceVersion: configs.Env("OTEL_SERVICE_VERSION", "1.0.0"),
		Environment:    configs.Env("OTEL_ENVIRONMENT", "local"),
		Endpoint:       configs.Env("OTEL_EXPORTER_ENDPOINT", "localhost:4319"),
	})
}
