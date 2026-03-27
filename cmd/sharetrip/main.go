package main

import (
	"context"
	"github.com/gofiber/fiber/v2"
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
	infoService := service.NewInfoService(repo)
	commandService := service.NewCommandTripService(repoTrip, validate)
	queryService := service.NewQueryTripService(repoTrip)
	tripHandler := api.NewHandler(commandService, queryService)

	server := api.NewServer(infoService, tripHandler) // ← add to service

	app := fiber.New()
	server.Route(app.Group(APIPrefix))
	server.RouteV2(app.Group(APIPrefixV2))

	printRoute(app)

	err = app.Listen(":8080")
	if err != nil {
		log.Fatal(err)
	}
}

// ВЫВОД МАРШРУТОВ, precheck.
func printRoute(app *fiber.App) {
	log.Info("Зарегистрированные маршруты:")
	for _, routes := range app.Stack() {
		for _, route := range routes {
			log.Debug("Метод и Путь: ", route.Method, route.Path)
		}
	}
}
