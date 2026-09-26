package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

const (
	httpPortEnv             = "HTTP_PORT"
	httpPortDefault         = "8080"
	databaseDSNEnv          = "DATABASE_DSN"
	contractServiceURLEnv   = "CONTRACT_SERVICE_URL"
	contractServiceURLDef   = "http://localhost:8082"
	
	kafkaBrokersEnv         = "KAFKA_BROKERS"
	kafkaBrokersDefault     = "localhost:9092"
	tripEventsTopicEnv      = "KAFKA_TOPIC_TRIP_EVENTS"
	tripEventsTopicDefault  = "trip.events"
	
	requestTimeoutMSEnv     = "REQUEST_TIMEOUT_MS"
	requestTimeoutMSDefault = "1500"
	retryAttemptsEnv        = "RETRY_ATTEMPTS"
	retryAttemptsDefault    = "2"
	
	outboxPollIntervalEnv   = "OUTBOX_POLL_INTERVAL_MS"
	outboxBatchSizeEnv      = "OUTBOX_BATCH_SIZE"
	outboxPollIntervalDef   = 1000
	outboxBatchSizeDef      = 50

	keycloakIssuerEnv       = "KEYCLOAK_ISSUER"
	keycloakClientIDEnv     = "KEYCLOAK_CLIENT_ID"
	keycloakClientSecretEnv = "KEYCLOAK_CLIENT_SECRET"
)

// Config is the runtime configuration loaded from process environment.
// Same keys work for local (.env via godotenv), Docker, and Kubernetes envFrom.
type Config struct {
	HTTPPort                        string
	DatabaseDSN                     string
	ContractServiceURL              string
	KafkaBrokers                    string
	KafkaTripEventsTopic            string
	RequestTimeout                  time.Duration
	RetryAttempts                   int
	KafkaOutboxPullIntervalDuration time.Duration
	KafkaOutboxBatchSize            int

	KeycloakIssuer       string
	KeycloakClientID     string
	KeycloakClientSecret string
}

func LoadAppConfig() (Config, error) {
	timeoutMS, err := strconv.Atoi(Getenv(requestTimeoutMSEnv, requestTimeoutMSDefault))
	if err != nil {
		return Config{}, fmt.Errorf("parse %s: %w", requestTimeoutMSEnv, err)
	}

	retryAttempts, err := strconv.Atoi(Getenv(retryAttemptsEnv, retryAttemptsDefault))
	if err != nil {
		return Config{}, fmt.Errorf("parse %s: %w", retryAttemptsEnv, err)
	}

	cfg := Config{
		HTTPPort:                        Getenv(httpPortEnv, httpPortDefault),
		DatabaseDSN:                     os.Getenv(databaseDSNEnv),
		ContractServiceURL:              Getenv(contractServiceURLEnv, contractServiceURLDef),
		KafkaBrokers:                    Getenv(kafkaBrokersEnv, kafkaBrokersDefault),
		KafkaTripEventsTopic:            Getenv(tripEventsTopicEnv, tripEventsTopicDefault),
		RequestTimeout:                  time.Duration(timeoutMS) * time.Millisecond,
		RetryAttempts:                   retryAttempts,
		KafkaOutboxPullIntervalDuration: envDurationMS(outboxPollIntervalEnv, outboxPollIntervalDef),
		KafkaOutboxBatchSize:            EnvInt(outboxBatchSizeEnv, outboxBatchSizeDef),
		KeycloakIssuer:                  os.Getenv(keycloakIssuerEnv),
		KeycloakClientID:                os.Getenv(keycloakClientIDEnv),
		KeycloakClientSecret:            os.Getenv(keycloakClientSecretEnv),
	}

	if cfg.DatabaseDSN == "" {
		return Config{}, fmt.Errorf("%s is required", databaseDSNEnv)
	}
	if cfg.ContractServiceURL == "" {
		return Config{}, fmt.Errorf("%s is required", contractServiceURLEnv)
	}
	if cfg.KafkaBrokers == "" {
		return Config{}, fmt.Errorf("%s is required", kafkaBrokersEnv)
	}
	if cfg.KeycloakIssuer == "" {
		return Config{}, fmt.Errorf("%s is required", keycloakIssuerEnv)
	}
	if cfg.KeycloakClientID == "" {
		return Config{}, fmt.Errorf("%s is required", keycloakClientIDEnv)
	}
	if cfg.KeycloakClientSecret == "" || cfg.KeycloakClientSecret == "secret" {
		return Config{}, fmt.Errorf("%s is required (not empty / not placeholder)", keycloakClientSecretEnv)
	}

	return cfg, nil
}

func Getenv(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

// envDurationMS reads an env var as milliseconds (e.g. OUTBOX_POLL_INTERVAL_MS=1000 → 1s).
func envDurationMS(key string, defMS int) time.Duration {
	return time.Duration(EnvInt(key, defMS)) * time.Millisecond
}

func EnvInt(key string, def int) int {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}
