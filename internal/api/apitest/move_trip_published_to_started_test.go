package apitest

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"job4j.ru/share_trip/internal/api"
	"job4j.ru/share_trip/internal/api/apitest/fixtures"
)

const (
	testCompanyID                 = "acme01"
	testServiceCode               = "trip_start"
	moveTripPublishedToStartedURL = GroupPrefixV2 +
		"/trip/moveTripPublished-ToStarted/%s/company/%s/service/%s"
)

func TestServer_MoveTripPublishedToStarted(t *testing.T) {
	t.Parallel()

	t.Run("success_when_contract_allows", func(t *testing.T) {
		t.Parallel()
		lockIT(t)
		fixtures.UseStubClientID(t, fixtures.NormalClientID)
		UseContractStub(t, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{"allowed": true, "reason": "ok"})
		})

		tripID := mustCreatePublishedTrip(t)

		resp := mustStartTrip(t, tripID)
		defer closeResponseBody(t, resp)

		require.Equal(t, http.StatusOK, resp.StatusCode)

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		var got api.MoveTripPublishedToStartedResponse
		require.NoError(t, json.Unmarshal(body, &got))
		require.Equal(t, api.StatusEnum("started"), got.Status)
		require.True(t, got.Allowed)

		require.Equal(t, "started", tripStatusName(t, tripID))
	})

	t.Run("conflict_when_contract_denies_status_unchanged", func(t *testing.T) {
		t.Parallel()
		lockIT(t)
		fixtures.UseStubClientID(t, fixtures.NormalClientID)
		UseContractStub(t, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{"allowed": false, "reason": "quota"})
		})

		tripID := mustCreatePublishedTrip(t)

		resp := mustStartTrip(t, tripID)
		defer closeResponseBody(t, resp)

		require.Equal(t, http.StatusConflict, resp.StatusCode)
		require.Equal(t, "published", tripStatusName(t, tripID))
	})

	t.Run("conflict_when_contract_company_not_found", func(t *testing.T) {
		t.Parallel()
		lockIT(t)
		fixtures.UseStubClientID(t, fixtures.NormalClientID)
		UseContractStub(t, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]any{"error": "company not found"})
		})

		tripID := mustCreatePublishedTrip(t)

		resp := mustStartTrip(t, tripID)
		defer closeResponseBody(t, resp)

		require.Equal(t, http.StatusConflict, resp.StatusCode)
		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		var apiResp api.Response
		require.NoError(t, json.Unmarshal(body, &apiResp))
		require.Contains(t, strings.ToLower(apiResp.Message), "company not found")
		require.Equal(t, "published", tripStatusName(t, tripID))
	})

	t.Run("unavailable_when_contract_503_status_unchanged", func(t *testing.T) {
		t.Parallel()
		lockIT(t)
		fixtures.UseStubClientID(t, fixtures.NormalClientID)
		UseContractStub(t, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusServiceUnavailable)
		})

		tripID := mustCreatePublishedTrip(t)

		resp := mustStartTrip(t, tripID)
		defer closeResponseBody(t, resp)

		require.Equal(t, http.StatusServiceUnavailable, resp.StatusCode)

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		var apiResp api.Response
		require.NoError(t, json.Unmarshal(body, &apiResp))
		require.Equal(t, api.ErrorContractUnavailable, apiResp.Message)
		require.Equal(t, "published", tripStatusName(t, tripID))
	})

	t.Run("forbidden_when_caller_is_not_driver", func(t *testing.T) {
		t.Parallel()
		lockIT(t)
		fixtures.UseStubClientID(t, fixtures.NormalClientID)
		UseContractStub(t, defaultContractStub)

		tripID := mustCreatePublishedTrip(t)

		fixtures.UseStubClientID(t, fixtures.InvalidClientID)
		resp := mustStartTrip(t, tripID)
		defer closeResponseBody(t, resp)

		require.Equal(t, http.StatusForbidden, resp.StatusCode)
		require.Equal(t, "published", tripStatusName(t, tripID))
	})

	t.Run("unauthorized_without_claims", func(t *testing.T) {
		t.Parallel()
		lockIT(t)
		fixtures.UseStubClientID(t, fixtures.NormalClientID)
		tripID := mustCreatePublishedTrip(t)

		fixtures.UseStubNoClaims(t)
		resp := mustStartTrip(t, tripID)
		defer closeResponseBody(t, resp)

		require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})
}

func mustCreatePublishedTrip(t *testing.T) string {
	t.Helper()
	created := mustCreateTripDraft(t, createTripDraftRequestModel())
	publishBody := api.MoveTripDraftToPublishRequest{
		ClientID: fixtures.CurrentStubClientID(),
	}
	resp := mustPublishTripDraft(t, created.ID.String(), publishBody)
	defer closeResponseBody(t, resp)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	return created.ID.String()
}

func closeResponseBody(t *testing.T, resp *http.Response) {
	t.Helper()
	if err := resp.Body.Close(); err != nil {
		t.Logf("failed to close body: %v", err)
	}
}

func mustStartTrip(t *testing.T, tripID string) *http.Response {
	t.Helper()
	url := fmt.Sprintf(moveTripPublishedToStartedURL, tripID, testCompanyID, testServiceCode)
	req, err := http.NewRequest(http.MethodPatch, url, nil)
	require.NoError(t, err)
	req.Header.Set(fixtures.RefreshTokenHeader, fixtures.RefreshTokenValue)

	resp, err := testApp.Test(req, -1)
	require.NoError(t, err)
	return resp
}

func tripStatusName(t *testing.T, tripID string) string {
	t.Helper()
	var name string
	err := testDB.QueryRowContext(testCtx, `
		SELECT ts.name
		FROM trips t
		JOIN trip_status ts ON ts.id = t.trip_status_id
		WHERE t.id = $1
	`, tripID).Scan(&name)
	require.NoError(t, err)
	return name
}
