package app

import (
	"context"
	"log/slog"
	"os"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus"
	"job4j.ru/share_trip/configs"
	"job4j.ru/share_trip/internal/api"
	clientContract "job4j.ru/share_trip/internal/client/contracts"
	clientContractUsecase "job4j.ru/share_trip/internal/client/contracts/usecase"
	"job4j.ru/share_trip/internal/middleware"
	"job4j.ru/share_trip/internal/observability/logctx"
	"job4j.ru/share_trip/internal/observability/metrics"
	"job4j.ru/share_trip/internal/observability/tracing"
	"job4j.ru/share_trip/internal/storage"
	"job4j.ru/share_trip/internal/trip/service"
	"job4j.ru/share_trip/internal/trip/usecase"
)

// build - build server
func BuildServer(
	app *fiber.App,
	pool *pgxpool.Pool,
	registry *prometheus.Registry,
	m *metrics.Metrics,
	keycloakCfg middleware.KeycloakConfig,
) {
	// Initialize the validator instance
	validate := validator.New(validator.WithRequiredStructEnabled())

	// rest client for contract service
	contractClient := clientContract.NewContractClient(configs.ContractServiceURL())

	repo := storage.NewRepoPg(pool)
	repoTrip := storage.NewTripRepository(m, pool)
	outboxRepo := storage.NewOutboxEventRepository()

	infoUseCase := usecase.NewInfoUseCase()
	contractUsecase := clientContractUsecase.NewContractUsecase(contractClient)
	tripUseCase := usecase.NewTripUseCase(contractUsecase)

	infoService := service.NewInfoService(infoUseCase, repo)
	tripService := service.NewTripService(m, pool, repoTrip, outboxRepo, tripUseCase)

	server := api.NewServer(registry, validate, infoService, tripService)

	keycloakAuth := middleware.KeycloakRefreshTokenMiddleware(keycloakCfg)
	server.SetupRoutes(app, keycloakAuth)
}

func ReadDBConfig() storage.Config {
	return storage.Config{
		Host:     configs.Env("DB_HOST", "localhost"),
		Port:     configs.EnvInt("DB_PORT", 6543),
		User:     configs.Env("DB_USER", "postgres"),
		Password: configs.Env("DB_PASSWORD", "password"),
		DBName:   configs.Env("DB_NAME", "share_trip"),
		SSLMode:  configs.Env("DB_SSLMODE", "disable"),
	}
}

func InitTracing(ctx context.Context) (*tracing.TracerProvider, error) {
	return tracing.NewProvider(ctx, tracing.Config{
		ServiceName:    configs.Env("OTEL_SERVICE_NAME", "share-trip"),
		ServiceVersion: configs.Env("OTEL_SERVICE_VERSION", "1.0.0"),
		Environment:    configs.Env("OTEL_ENVIRONMENT", "local"),
		Endpoint:       configs.Env("OTEL_EXPORTER_ENDPOINT", "localhost:4319"),
	})
}

// getKeycloakConfig - get the keycloak config
func GetKeycloakConfig() middleware.KeycloakConfig {
	return middleware.KeycloakConfig{
		Issuer:       configs.Env("KEYCLOAK_ISSUER", "http://localhost:8087/realms/sharetrip"),
		ClientID:     configs.Env("KEYCLOAK_CLIENT_ID", "sharetrip-api"),
		ClientSecret: configs.Env("KEYCLOAK_CLIENT_SECRET", ""),
	}
}

// LogContractConfig logs resolved Contract Service base URL (from .env or default).
func LogContractConfig() {
	logctx.Logger(context.Background()).Info("contract service config",
		slog.String("url", configs.ContractServiceURL()),
		slog.String("env_var", configs.ContractServiceEnv),
	)
}

// logKeycloakConfig - log the keycloak config
func LogKeycloakConfig(keycloakCfg middleware.KeycloakConfig) {
	cwd, _ := os.Getwd()
	logctx.Logger(context.Background()).Info("keycloak config",
		slog.String("issuer", keycloakCfg.Issuer),
		slog.String("client_id", keycloakCfg.ClientID),
		slog.Int("secret_len", len(keycloakCfg.ClientSecret)),
		slog.String("cwd", cwd),
	)
	if keycloakCfg.ClientSecret == "" || keycloakCfg.ClientSecret == "secret" {
		logctx.Logger(context.Background()).Error("KEYCLOAK_CLIENT_SECRET empty or placeholder 'secret' — save real secret in .env, then: " +
			"Remove-Item Env:KEYCLOAK_CLIENT_SECRET; make run")
	}
}
