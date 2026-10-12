// Package app contains application-level initialization and configuration logic.

package app

import (
	"context"
	"log/slog"
	"os"
	"strings"

	"job4j.ru/share_trip/internal/clients/kafka"
	"job4j.ru/share_trip/internal/config"
	"job4j.ru/share_trip/internal/middleware"
	"job4j.ru/share_trip/internal/observability/logctx"
	"job4j.ru/share_trip/internal/storage"
)

const (
	appName                 = "share-trip"
	kafkaBrokersPortDef     = "localhost:9092"
	kafkaTopicTripEventsDef = "trip.events"

	keycloakIssuerDef   = "http://localhost:8087/realms/sharetrip"
	keycloakClientIDDef = "sharetrip-api"
)

func NewKafkaProducer() *kafka.Producer {
	brokersCSV := config.Env("KAFKA_BROKERS", kafkaBrokersPortDef)
	topic := config.Env("KAFKA_TOPIC_TRIP_EVENTS", kafkaTopicTripEventsDef)
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

// GetKeycloakConfig - get the keycloak config
func GetKeycloakConfig() middleware.KeycloakConfig {
	return middleware.KeycloakConfig{
		Issuer:       config.Env("KEYCLOAK_ISSUER", keycloakIssuerDef),
		ClientID:     config.Env("KEYCLOAK_CLIENT_ID", keycloakClientIDDef),
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

// LogKeycloakConfig - log the keycloak config
func LogKeycloakConfig(keycloakCfg middleware.KeycloakConfig) {
	cwd, _ := os.Getwd()
	logctx.Logger(context.Background()).Info("keycloak config",
		slog.String("issuer", keycloakCfg.Issuer),
		slog.String("client_id", keycloakCfg.ClientID),
		slog.Int("secret_len", len(keycloakCfg.ClientSecret)),
		slog.String("cwd", cwd),
	)
	if keycloakCfg.ClientSecret == "" || keycloakCfg.ClientSecret == "secret" {
		logctx.Logger(context.Background()).
			Error("KEYCLOAK_CLIENT_SECRET empty or placeholder 'secret' — save real secret in .env, then: " +
				"Remove-Item Env:KEYCLOAK_CLIENT_SECRET; make run")
	}
}
