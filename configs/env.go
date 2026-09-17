package configs

import (
	"os"
	"strconv"
	"time"
)

func Env(key, def string) string {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	return v
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

// EnvDurationMS reads an env var as milliseconds (e.g. OUTBOX_POLL_INTERVAL_MS=1000 → 1s).
func EnvDurationMS(key string, defMS int) time.Duration {
	return time.Duration(EnvInt(key, defMS)) * time.Millisecond
}

// ContractServiceURL returns CONTRACT_SERVICE_URL or configs.BaseURL default.
func ContractServiceURL() string {
	return Env(ContractServiceEnv, BaseURL)
}
