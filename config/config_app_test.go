package config

import (
	"testing"
	"time"
)

// Tests use t.Setenv to isolate from make/shell .env.
// Since Go 1.25+, t.Setenv is incompatible with t.Parallel — paralleltest waived.

//nolint:paralleltest // t.Setenv cannot run with t.Parallel
func TestLoadAppConfig_RequiresDatabaseDSN(t *testing.T) {
	// Empty DSN must fail even if the parent shell exported DATABASE_DSN (e.g. make include .env).
	t.Setenv("DATABASE_DSN", "")
	t.Setenv("CONTRACT_SERVICE_URL", "http://localhost:8082")
	t.Setenv("KAFKA_BROKERS", "localhost:9092")
	t.Setenv("KEYCLOAK_ISSUER", "http://localhost:8087/realms/sharetrip")
	t.Setenv("KEYCLOAK_CLIENT_ID", "sharetrip-api")
	t.Setenv("KEYCLOAK_CLIENT_SECRET", "test-secret")

	_, err := LoadAppConfig()
	if err == nil {
		t.Fatal("expected error when DATABASE_DSN is empty")
	}
}

//nolint:paralleltest // t.Setenv cannot run with t.Parallel
func TestLoadAppConfig_OK(t *testing.T) {
	t.Setenv("DATABASE_DSN", "postgres://u:p@localhost:6543/db?sslmode=disable")
	t.Setenv("CONTRACT_SERVICE_URL", "http://localhost:8082")
	t.Setenv("KAFKA_BROKERS", "localhost:9092")
	t.Setenv("KEYCLOAK_ISSUER", "http://localhost:8087/realms/sharetrip")
	t.Setenv("KEYCLOAK_CLIENT_ID", "sharetrip-api")
	t.Setenv("KEYCLOAK_CLIENT_SECRET", "test-secret")
	t.Setenv("REQUEST_TIMEOUT_MS", "2000")
	t.Setenv("RETRY_ATTEMPTS", "3")
	t.Setenv("OUTBOX_POLL_INTERVAL_MS", "500")
	t.Setenv("OUTBOX_BATCH_SIZE", "10")

	cfg, err := LoadAppConfig()
	if err != nil {
		t.Fatalf("LoadAppConfig: %v", err)
	}
	if cfg.DatabaseDSN == "" {
		t.Fatal("DatabaseDSN empty")
	}
	if cfg.RequestTimeout != 2*time.Second {
		t.Fatalf("RequestTimeout=%v", cfg.RequestTimeout)
	}
	if cfg.RetryAttempts != 3 {
		t.Fatalf("RetryAttempts=%d", cfg.RetryAttempts)
	}
	if cfg.KafkaOutboxPullIntervalDuration != 500*time.Millisecond {
		t.Fatalf("outbox interval=%v", cfg.KafkaOutboxPullIntervalDuration)
	}
	if cfg.KafkaOutboxBatchSize != 10 {
		t.Fatalf("batch=%d", cfg.KafkaOutboxBatchSize)
	}
	if cfg.KeycloakClientSecret != "test-secret" {
		t.Fatal("keycloak secret not loaded")
	}
}
