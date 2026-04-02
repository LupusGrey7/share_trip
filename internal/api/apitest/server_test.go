package apitest

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/go-playground/validator/v10"
	"job4j.ru/share_trip/internal/api"
	"job4j.ru/share_trip/internal/repository"
	"log"
	"os"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/testcontainers/testcontainers-go/modules/postgres"

	"job4j.ru/share_trip/internal/service"
)

// var TestApp *fiber.App
var (
	testCtx       context.Context
	testDB        *sql.DB
	testPool      *pgxpool.Pool
	testApp       *fiber.App
	testContainer *postgres.PostgresContainer
)

func TestMain(m *testing.M) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("panic in TestMain: %v", r)
		}
	}()
	testCtx = context.Background()

	// === 1. Запуск PostgreSQL контейнера ===
	var err error
	testContainer, err = postgres.Run(
		testCtx,
		"postgres:17",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("password"),
	)
	if err != nil {
		log.Fatalf("failed to start postgres container: %v", err)
	}

	dsn, err := testContainer.ConnectionString(testCtx, "sslmode=disable")
	if err != nil {
		log.Fatalf("failed to get connection string: %v", err)
	}
	log.Println("Postgres container started")
	testDB, err = sql.Open("pgx", dsn)
	if err != nil {
		log.Fatalf("failed to open sql.DB: %v", err)
	}

	waitReady(testDB)
	// Миграции
	if err = goose.SetDialect("postgres"); err != nil {
		log.Fatalf("set goose dialect: %v", err)
	}
	if err = goose.Up(testDB, "../../../migrations"); err != nil {
		log.Fatalf("run migrations: %v", err)
	}

	testPool, err = pgxpool.New(testCtx, dsn)
	if err != nil {
		log.Fatalf("failed to create pgxpool: %v", err)
	}
	log.Println("Database and pool ready, migrations applied")
	// Инициализация зависимостей (validator, сервисы и т.д.)
	validate := validator.New(validator.WithRequiredStructEnabled())
	repo := repository.NewRepoPg(testPool)
	outboxRepo := repository.NewOutboxEventRepository()
	repoTrip := repository.NewTripRepository(testPool)

	infoService := service.NewInfoService(repo)
	tripService := service.NewTripService(repoTrip, validate)
	commandService := service.NewCommandTripService(testPool, repoTrip, outboxRepo, validate)
	queryService := service.NewQueryTripService(repoTrip)

	//server
	server := api.NewServer(infoService, tripService, commandService, queryService)

	// === 2. Создание Fiber приложения ===
	//testApp = fiber.New()
	testApp = fiber.New(fiber.Config{
		EnablePrintRoutes: true, // ← Включаем автоматический вывод маршрутов при старте
	})

	server.Route(testApp.Group(""))
	server.RouteV2(testApp.Group(""))
	// Вывод всех зарегистрированных маршрутов в консоль (явно)
	printRegisteredRoutes(testApp)
	log.Println("=== Test application ready ===")

	//  ===Запускаем тесты ===
	code := m.Run()

	// === 3. Корректное завершение ресурсов ===
	log.Println("=== Starting forced shutdown sequence ===")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer cancel()

	// 3.1 Fiber (даём время на graceful shutdown)
	if testApp != nil {
		log.Println("Shutting down Fiber server...")
		if err := testApp.ShutdownWithContext(shutdownCtx); err != nil {
			log.Printf("Fiber shutdown error (continuing anyway): %v", err)
		} else {
			log.Println("Fiber shutdown completed")
		}
	}

	// 3.2 Закрываем пулы соединений с таймаутом
	if testPool != nil {
		log.Println("Closing pgxpool...")
		done := make(chan struct{})
		go func() {
			testPool.Close()
			close(done)
		}()
		select {
		case <-done:
			log.Println("pgxpool closed successfully")
		case <-time.After(5 * time.Second):
			log.Println("pgxpool.Close() timed out - forcing continue")
		}
	}

	if testDB != nil {
		log.Println("Closing sql.DB...")
		_ = testDB.Close()
		log.Println("sql.DB closed")
	}

	// 3.3 Завершаем Docker-контейнер с большим таймаутом
	if testContainer != nil {
		log.Println("Terminating Postgres container...")
		termCtx, termCancel := context.WithTimeout(context.Background(), 20*time.Second)
		if err := testContainer.Terminate(termCtx); err != nil {
			log.Printf("Container terminate error (continuing): %v", err)
		} else {
			log.Println("Postgres container terminated successfully")
		}
		termCancel()
	}

	// 3.4 Принудительных выход через 2 секунды (гарантирует завершение теста)
	log.Println("=== All cleanup done. Forcing os.Exit ===")
	time.Sleep(2 * time.Second) // даём логам вывестись
	os.Exit(code)
}

func waitReady(db *sql.DB) {
	deadline := time.Now().Add(30 * time.Second)

	for time.Now().Before(deadline) {
		ctx, cancel := context.WithTimeout(
			context.Background(),
			2*time.Second,
		)
		err := db.PingContext(ctx)
		cancel()

		if err == nil {
			return
		}

		time.Sleep(500 * time.Millisecond)
	}

	log.Fatalf("database is not ready after timeout")
}

func printRegisteredRoutes(app *fiber.App) {
	fmt.Println("\n=== Registered Routes ===")
	routes := app.GetRoutes(true) // true = исключить middleware-only роуты
	for _, route := range routes {
		fmt.Printf("%-6s %s\n", route.Method, route.Path)
	}
	fmt.Println("=========================")
}
