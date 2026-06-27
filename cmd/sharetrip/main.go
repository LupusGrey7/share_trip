package main

import (
	"context"
	"github.com/prometheus/client_golang/prometheus"
	"job4j.ru/share_trip/internal/observability/metrics"
	"job4j.ru/share_trip/internal/observability/tracing"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	applog "job4j.ru/share_trip/internal/app"
	"job4j.ru/share_trip/internal/domain/trip/usecase"
	"job4j.ru/share_trip/internal/middleware"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/joho/godotenv"
	"job4j.ru/share_trip/configs"
	"job4j.ru/share_trip/internal/api"
	"job4j.ru/share_trip/internal/repository"
	"job4j.ru/share_trip/internal/service"
	"job4j.ru/share_trip/internal/storage"
)

// init is invoked before main()
func init() {
	// loads values from .env into the system
	if err := godotenv.Load(); err != nil {
		log.Info("No .env file found")
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

	//init Tracing (OpenTelemetry and Jaeger)
	initTracing(ctx)

	//app
	app := fiber.New(fiber.Config{
		EnablePrintRoutes: true,
	})
	app.Use(requestid.New())
	app.Use(func(c *fiber.Ctx) error {
		log.Infof("Generated ID: %v", c.Locals("requestid"))
		return c.Next()
	})
	app.Use(middleware.Correlation(logger)) //add custom logger, before add api
	app.Use(api.NewHTTPMetricsMiddleware(m))

	//build Server
	build(app, pool, registry, m)

	err = app.Listen(":8080")
	if err != nil {
		log.Fatal(err)
	}
}

// build - build server
func build(
	app *fiber.App,
	pool *pgxpool.Pool,
	registry *prometheus.Registry,
	m *metrics.Metrics,
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

	server := api.NewServer(registry, validate, infoService, tripService) // ← add all service`s

	//route
	server.SetupRoutes(app)
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

func initTracing(ctx context.Context) {
	tp, err := tracing.NewProvider(ctx, tracing.Config{
		ServiceName:    "share-trip",
		ServiceVersion: "1.0.0",
		Environment:    "local",
		Endpoint:       "localhost:4319",
	})
	if err != nil {
		log.Error("init tracing failed", "error", err)
		os.Exit(1)
	}

	defer func() {
		shutdownCtx, cancel := context.WithTimeout(
			context.Background(),
			5*time.Second,
		)
		defer cancel()

		if err := tp.Shutdown(shutdownCtx); err != nil {
			log.Error("shutdown tracing failed", "error", err)
		}
	}()
}
