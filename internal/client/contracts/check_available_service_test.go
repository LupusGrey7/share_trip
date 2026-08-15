package contracts_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"job4j.ru/share_trip/internal/client/contracts"
)

func TestCheckAvailableService_Allowed(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		require.Contains(t, r.URL.Path, "/companies/acme/services/trip_start/availability")
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"allowed": true, "reason": "ok"})
	}))
	t.Cleanup(srv.Close)

	client := contracts.NewContractClient(srv.URL)
	got, err := client.CheckAvailableService(context.Background(), "acme", "trip_start")
	require.NoError(t, err)
	require.True(t, got.Allowed)
	require.Equal(t, "ok", got.Reason)
}

func TestCheckAvailableService_DeniedNoRetry(t *testing.T) {
	t.Parallel()

	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"allowed": false, "reason": "quota"})
	}))
	t.Cleanup(srv.Close)

	client := contracts.NewContractClient(srv.URL)
	got, err := client.CheckAvailableService(context.Background(), "acme", "trip_start")
	require.NoError(t, err)
	require.False(t, got.Allowed)
	require.Equal(t, "quota", got.Reason)
	require.Equal(t, int32(1), calls.Load(), "allowed:false must not retry")
}

func TestCheckAvailableService_BadRequestNoRetry(t *testing.T) {
	t.Parallel()

	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"bad"}`))
	}))
	t.Cleanup(srv.Close)

	client := contracts.NewContractClient(srv.URL)
	_, err := client.CheckAvailableService(context.Background(), "acme", "trip_start")
	require.Error(t, err)
	require.True(t, errors.Is(err, contracts.ErrBadRequest))
	require.Equal(t, int32(1), calls.Load(), "400 must not retry")
}

func TestCheckAvailableService_Forbidden(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	t.Cleanup(srv.Close)

	client := contracts.NewContractClient(srv.URL)
	_, err := client.CheckAvailableService(context.Background(), "acme", "trip_start")
	require.Error(t, err)
	require.True(t, errors.Is(err, contracts.ErrForbidden))
}

func TestCheckAvailableService_Retry503ThenOK(t *testing.T) {
	t.Parallel()

	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := calls.Add(1)
		if n == 1 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"allowed": true, "reason": "ok"})
	}))
	t.Cleanup(srv.Close)

	client := contracts.NewContractClient(srv.URL)
	got, err := client.CheckAvailableService(context.Background(), "acme", "trip_start")
	require.NoError(t, err)
	require.True(t, got.Allowed)
	require.GreaterOrEqual(t, calls.Load(), int32(2))
}

func TestCheckAvailableService_UnavailableAfterRetries(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	t.Cleanup(srv.Close)

	client := contracts.NewContractClient(srv.URL)
	_, err := client.CheckAvailableService(context.Background(), "acme", "trip_start")
	require.Error(t, err)
	require.True(t, errors.Is(err, contracts.ErrUnavailable))
}

func TestCheckAvailableService_Timeout(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(3 * time.Second) // configs.Timeout = 2s
		_ = json.NewEncoder(w).Encode(map[string]any{"allowed": true})
	}))
	t.Cleanup(srv.Close)

	client := contracts.NewContractClient(srv.URL)
	_, err := client.CheckAvailableService(context.Background(), "acme", "trip_start")
	require.Error(t, err)
	require.True(t, errors.Is(err, contracts.ErrTimeout) || errors.Is(err, contracts.ErrUnavailable),
		"got: %v", err)
}
