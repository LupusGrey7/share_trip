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
	"job4j.ru/share_trip/config"
	"job4j.ru/share_trip/internal/api"
	cc "job4j.ru/share_trip/internal/clients/http/contract"
	ccus "job4j.ru/share_trip/internal/clients/http/contract/usecase"
	"job4j.ru/share_trip/internal/clients/kafka"
	"job4j.ru/share_trip/internal/middleware"
	"job4j.ru/share_trip/internal/observability/logctx"
	"job4j.ru/share_trip/internal/observability/metrics"
	"job4j.ru/share_trip/internal/observability/tracing"
	"job4j.ru/share_trip/internal/outbox"
	outboxservice "job4j.ru/share_trip/internal/outbox/service"
	outboxusecase "job4j.ru/share_trip/internal/outbox/usecase"
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
	config config.Config,
) *outbox.Publisher {
	validate := validator.New(validator.WithRequiredStructEnabled())

	contractClient := cc.NewContractClient(config.ContractServiceURL)

	repo := storage.NewRepoPg(pool)
	repoTrip := storage.NewTripRepository(m, pool)
	outboxRepo := storage.NewOutboxEventRepository(m)

	infoUseCase := usecase.NewInfoUseCase()
	contractUseCase := ccus.NewContractUsecase(contractClient)
	tripUseCase := usecase.NewTripUseCase(contractUseCase)
	outboxUC := outboxusecase.NewOutboxUseCase(outboxRepo)

	infoService := service.NewInfoService(infoUseCase, repo)
	tripService := service.NewTripService(m, pool, repoTrip, outboxRepo, tripUseCase)
	outboxSvc := outboxservice.NewOutboxService(m, pool, outboxUC)

	kafkaProducer := newKafkaProducer(config)
	outboxPublisher := outbox.NewPublisher(
		m,
		kafkaProducer,
		outboxSvc,
		config.KafkaOutboxPullIntervalDuration,
		config.KafkaOutboxBatchSize,
	)

	server := api.NewServer(registry, validate, infoService, tripService)

	keycloakAuth := middleware.KeycloakRefreshTokenMiddleware(keycloakCfg)
	server.SetupRoutes(app, keycloakAuth)

	return outboxPublisher
}

func newKafkaProducer(config config.Config) kafka.TripEventProducer {
	brokersCSV := config.KafkaBrokers
	topic := config.KafkaTripEventsTopic
	brokers := strings.Split(brokersCSV, ",")
	for i := range brokers {
		brokers[i] = strings.TrimSpace(brokers[i])
	}
	return kafka.NewProducer(brokers, topic)
}

func InitTracing(ctx context.Context) (*tracing.TracerProvider, error) {
	return tracing.NewProvider(ctx, tracing.Config{
		ServiceName:    config.Getenv("OTEL_SERVICE_NAME", "share-trip"),
		ServiceVersion: config.Getenv("OTEL_SERVICE_VERSION", "1.0.0"),
		Environment:    config.Getenv("OTEL_ENVIRONMENT", "local"),
		Endpoint:       config.Getenv("OTEL_EXPORTER_ENDPOINT", "localhost:4319"),
	})
}

// KeycloakFromConfig maps validated app config into middleware settings (no secrets logged here).
func KeycloakFromConfig(cfg config.Config) middleware.KeycloakConfig {
	return middleware.KeycloakConfig{
		Issuer:       cfg.KeycloakIssuer,
		ClientID:     cfg.KeycloakClientID,
		ClientSecret: cfg.KeycloakClientSecret,
	}
}

// LogContractConfig logs resolved Contract Service base URL (never secrets).
func LogContractConfig(cfg config.Config) {
	logctx.Logger(context.Background()).Info("contract service config",
		slog.String("url", cfg.ContractServiceURL),
	)
}

// LogKeycloakConfig logs issuer/client_id and secret length only (not the secret value).
func LogKeycloakConfig(keycloakCfg middleware.KeycloakConfig) {
	cwd, _ := os.Getwd()
	logctx.Logger(context.Background()).Info("keycloak config",
		slog.String("issuer", keycloakCfg.Issuer),
		slog.String("client_id", keycloakCfg.ClientID),
		slog.Int("secret_len", len(keycloakCfg.ClientSecret)),
		slog.String("cwd", cwd),
	)
}
