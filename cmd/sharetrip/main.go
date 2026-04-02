package main

import (
	"context"
	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"job4j.ru/share_trip/configs"
	"job4j.ru/share_trip/internal/api"
	"job4j.ru/share_trip/internal/repository"
	"job4j.ru/share_trip/internal/service"
	_ "time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2/log"
	"job4j.ru/share_trip/internal/storage"
)

const (
	APIPrefix   = "/api/v1"
	APIPrefixV2 = "/api/v2"
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

	cfg := storage.Config{
		Host:     configs.Env("DB_HOST", "localhost"),
		Port:     configs.EnvInt("DB_PORT", 6543),
		User:     configs.Env("DB_USER", "postgres"),
		Password: configs.Env("DB_PASSWORD", "password"),
		DBName:   configs.Env("DB_NAME", "share_trip"),
		SSLMode:  configs.Env("DB_SSLMODE", "disable"),
	}

	pool, err := storage.NewPool(ctx, cfg.DSN())
	if err != nil {
		log.Fatal(err)
	}

	defer pool.Close()
	// логирование подключения
	if pingErr := pool.Ping(ctx); pingErr != nil {
		log.Fatalf("failed to ping database: %v", pingErr)
	}
	log.Info("Connected to database successfully")

	// Initialize the validator instance
	validate := validator.New(validator.WithRequiredStructEnabled())
	repo := repository.NewRepoPg(pool)
	repoTrip := repository.NewTripRepository(pool)
	outboxRepo := repository.NewOutboxEventRepository()
	infoService := service.NewInfoService(repo)
	tripService := service.NewTripService(repoTrip, validate)
	commandService := service.NewCommandTripService(pool, repoTrip, outboxRepo, validate)
	queryService := service.NewQueryTripService(repoTrip)

	server := api.NewServer(infoService, tripService, commandService, queryService) // ← add to service

	app := fiber.New(fiber.Config{
		EnablePrintRoutes: true,
	})
	server.Route(app.Group(APIPrefix))
	server.RouteV2(app.Group(APIPrefixV2))

	err = app.Listen(":8080")
	if err != nil {
		log.Fatal(err)
	}
}
