package apitest

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"job4j.ru/share_trip/internal/observability/metrics"

	"job4j.ru/share_trip/internal/domain/trip/usecase"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"job4j.ru/share_trip/internal/api"
	"job4j.ru/share_trip/internal/repository"

	"job4j.ru/share_trip/internal/api/apitest/fixtures"
	"job4j.ru/share_trip/internal/service"
	client "job4j.ru/share_trip/internal/client/contracts"
	clientUsecase "job4j.ru/share_trip/internal/client/contracts/usecase"
)

const (
	GroupPrefixV1 = "/api/v1"
	GroupPrefixV2 = "/api/v2"
)

// var TestApp *fiber.App
var (
	testCtx       context.Context
	testDB        *sql.DB
	testPool      *pgxpool.Pool
	testApp       *fiber.App
	testContainer *postgres.PostgresContainer

	contractStubMu      sync.Mutex
	contractStubHandler http.HandlerFunc
	contractStubServer  *httptest.Server
)

// UseContractStub sets Contract httptest behavior for this test and restores default (allowed:true).
func UseContractStub(t *testing.T, handler http.HandlerFunc) {
	t.Helper()
	contractStubMu.Lock()
	prev := contractStubHandler
	contractStubHandler = handler
	contractStubMu.Unlock()

	t.Cleanup(func() {
		contractStubMu.Lock()
		contractStubHandler = prev
		contractStubMu.Unlock()
	})
}

func defaultContractStub(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"allowed": true, "reason": "ok"})
}


/*
=== Registered Routes ===
GET /ready
GET /trip/:tripId
HEAD /ready
HEAD /trip/:tripId
POST /trip/createTripDraft
PATCH /trip/moveTripDraft-ToPublish/:tripId
*/
func TestMain(m *testing.M) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("panic in TestMain: %v", r)
		}
	}()
	testCtx = context.Background()

	// === 1. Start PostgreSQL container ===
	var err error

	//init BD
	dsn := initTestDb()

	waitReady(testDB)

	// Migrations
	setUpMigrations(testDB)

	testPool, err = pgxpool.New(testCtx, dsn)
	if err != nil {
		log.Fatalf("failed to create pgxpool: %v", err)
	}
	log.Println("Database and pool ready, migrations applied")

	// Initialization of dependencies (validator, services, etc.)
	validate := validator.New(validator.WithRequiredStructEnabled())

	// 1. Create a clean local registry for test, to not pollute the global
	//registry Prometheus
	registry := prometheus.NewRegistry()

	counter := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "updates_skipped_total",
		Help: "Total skipped updates",
	})
	registry.MustRegister(counter)
	mu := metrics.New(registry)

	repo := repository.NewRepoPg(testPool)
	repoTrip := repository.NewTripRepository(mu, testPool)
	outboxRepo := repository.NewOutboxEventRepository()

	contractStubHandler = defaultContractStub
	contractStubServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		contractStubMu.Lock()
		h := contractStubHandler
		contractStubMu.Unlock()
		if h == nil {
			defaultContractStub(w, r)
			return
		}
		h(w, r)
	}))

	contractClient := client.NewContractClient(contractStubServer.URL)
	contractUsecase := clientUsecase.NewContractUsecase(contractClient)

	infoUseCase := usecase.NewInfoUseCase()
	tripUseCase := usecase.NewTripUseCase(contractUsecase)

	infoService := service.NewInfoService(infoUseCase, repo)
	tripService := service.NewTripService(mu, testPool, repoTrip, outboxRepo, tripUseCase)

	server := api.NewServer(registry, validate, infoService, tripService) // ← add to service

	// === 2. Create Fiber application ===
	//testApp = fiber.New()
	testApp = fiber.New(fiber.Config{
		EnablePrintRoutes: true, // ← Enable automatic route output at startup
	})
	testApp.Use(requestid.New())
	testApp.Use(func(c *fiber.Ctx) error {
		log.Printf("Generated TEST_TRACE_ID: %v", c.Locals("requestid"))
		return c.Next()
	})

	// build test Server with Routes including middleware
	server.SetupRoutes(testApp, fixtures.KeycloakRefreshTokenMiddleware()) // ← add middleware Keycloak Refresh Token

	// Output all registered routes to console (explicitly)
	printRegisteredRoutes(testApp)
	log.Println("=== Test application ready ===")

	//  ===Start tests ===
	code := m.Run()

	// === 3. Correct shutdown of resources ===
	log.Println("=== Starting forced shutdown sequence ===")

	if contractStubServer != nil {
		contractStubServer.Close()
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer cancel()

	// 3.1 Fiber (give time for graceful shutdown)
	if testApp != nil {
		log.Println("Shutting down Fiber server...")
		if err := testApp.ShutdownWithContext(shutdownCtx); err != nil {
			log.Printf("Fiber shutdown error (continuing anyway): %v", err)
		} else {
			log.Println("Fiber shutdown completed")
		}
	}

	// 3.2 Close connection pools with timeout
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

	// 3.3 Terminate Docker container with timeout
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

	// 3.4 Force exit through 2 seconds (guarantees test completion)
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
	routes := app.GetRoutes(true) // true = exclude middleware-only routes
	for _, route := range routes {
		fmt.Printf("%-6s %s\n", route.Method, route.Path)
	}
	fmt.Println("=========================")
}

func initTestDb() string {
	testContainer, err := postgres.Run(
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
	return dsn
}

func setUpMigrations(testDB *sql.DB) {
	if err := goose.SetDialect("postgres"); err != nil {
		log.Fatalf("set goose dialect: %v", err)
	}
	if err := goose.Up(testDB, "../../../migrations"); err != nil {
		log.Fatalf("run migrations: %v", err)
	}
}
