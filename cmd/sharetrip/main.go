package main

import (
	"context"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"job4j.ru/share_trip/configs"
	"job4j.ru/share_trip/internal/api"
	"job4j.ru/share_trip/internal/service"
	"log"
	_ "time"

	"job4j.ru/share_trip/internal/storage"
)

const (
	APIPrefix = "/api"
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
	fmt.Println("Connected to database successfully")

	repo := storage.NewRepoPg(pool)
	service := service.NewCommonService(repo)

	server := api.NewServer(service) // ← add to service

	app := fiber.New()
	server.Route(app.Group(APIPrefix))

	err = app.Listen(":8080")
	if err != nil {
		log.Fatal(err)
	}
}
