package app

import (
	"context"
	"log/slog"
	"os"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus"

	"job4j.ru/share_trip/internal/api"
	clientContract "job4j.ru/share_trip/internal/clients/http/contract"
	clientContractUsecase "job4j.ru/share_trip/internal/clients/http/contract/usecase"
	"job4j.ru/share_trip/internal/clients/kafka"
	"job4j.ru/share_trip/internal/config"
	"job4j.ru/share_trip/internal/middleware"
	"job4j.ru/share_trip/internal/observability/logctx"
	"job4j.ru/share_trip/internal/observability/metrics"
	"job4j.ru/share_trip/internal/observability/tracing"

	"job4j.ru/share_trip/internal/storage"
	"job4j.ru/share_trip/internal/trip/service"
	"job4j.ru/share_trip/internal/trip/usecase"
)

// BuildServer wires HTTP API and returns the outbox publisher (start with go p.Run(ctx)).
func BuildServer(
	app *fiber.App,
	pool *pgxpool.Pool,
	registry *prometheus.Registry,
	m *metrics.Metrics,
	keycloakCfg middleware.KeycloakConfig,
) {
	validate := validator.New(validator.WithRequiredStructEnabled())

	contractClient := clientContract.NewContractClient(config.ContractServiceURL())

	repo := storage.NewRepoPg(pool)
	repoTrip := storage.NewTripRepository(m, pool)
	outboxRepo := storage.NewOutboxEventRepository(m)

	infoUseCase := usecase.NewInfoUseCase()
	contractUseCase := clientContractUsecase.NewContractUsecase(contractClient)
	tripUseCase := usecase.NewTripUseCase(contractUseCase)

	infoService := service.NewInfoService(infoUseCase, repo)
	tripService := service.NewTripService(m, pool, repoTrip, outboxRepo, tripUseCase)

	server := api.NewServer(registry, validate, infoService, tripService)

	keycloakAuth := middleware.KeycloakRefreshTokenMiddleware(keycloakCfg)
	server.SetupRoutes(app, keycloakAuth)
}

func NewKafkaProducer() *kafka.Producer {
	brokersCSV := config.Env("KAFKA_BROKERS", "localhost:9092")
	topic := config.Env("KAFKA_TOPIC_TRIP_EVENTS", "trip.events")
	brokers := strings.Split(brokersCSV, ",")
	for i := range brokers {
		brokers[i] = strings.TrimSpace(brokers[i])
	}
	return kafka.NewProducer(brokers, topic)
}

func ReadDBConfig() storage.Config {
	return storage.Config{
		Host:     config.Env("DB_HOST", "localhost"),
		Port:     config.EnvInt("DB_PORT", 6543),
		User:     config.Env("DB_USER", "postgres"),
		Password: config.Env("DB_PASSWORD", "password"),
		DBName:   config.Env("DB_NAME", "share_trip"),
		SSLMode:  config.Env("DB_SSLMODE", "disable"),
	}
}

func InitTracing(ctx context.Context) (*tracing.TracerProvider, error) {
	return tracing.NewProvider(ctx, tracing.Config{
		ServiceName:    config.Env("OTEL_SERVICE_NAME", "share-trip"),
		ServiceVersion: config.Env("OTEL_SERVICE_VERSION", "1.0.0"),
		Environment:    config.Env("OTEL_ENVIRONMENT", "local"),
		Endpoint:       config.Env("OTEL_EXPORTER_ENDPOINT", "localhost:4319"),
	})
}

// getKeycloakConfig - get the keycloak config
func GetKeycloakConfig() middleware.KeycloakConfig {
	return middleware.KeycloakConfig{
		Issuer:       config.Env("KEYCLOAK_ISSUER", "http://localhost:8087/realms/sharetrip"),
		ClientID:     config.Env("KEYCLOAK_CLIENT_ID", "sharetrip-api"),
		ClientSecret: config.Env("KEYCLOAK_CLIENT_SECRET", ""),
	}
}

// LogContractConfig logs resolved Contract Service base URL (from .env or default).
func LogContractConfig() {
	logctx.Logger(context.Background()).Info("contract service config",
		slog.String("url", config.ContractServiceURL()),
		slog.String("env_var", config.ContractServiceEnv),
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
