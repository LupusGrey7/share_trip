package apitest

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"time"

	"job4j.ru/share_trip/internal/api"
	"job4j.ru/share_trip/internal/api/apitest/fixtures"

	"github.com/stretchr/testify/require"
)

const createTripDraftURL = GroupPrefixV2 + "/trip/createTripDraft"

func TestServer_CreateTrip(t *testing.T) {
	t.Parallel()


	// Given: valid request + JWT stub
	// When: POST createTripDraft
	// Then: 201, DriverID = Keycloak sub (not from body)
	t.Run("success_create_trip_draft", func(t *testing.T) {
		t.Parallel()
		lockIT(t)
		fixtures.UseStubClientID(t, fixtures.NormalClientID)
		payload := createTripRequestModel()

		body, err := json.Marshal(payload)
		require.NoError(t, err)

		req, err := http.NewRequest(http.MethodPost, createTripDraftURL, bytes.NewReader(body))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set(fixtures.RefreshTokenHeader, fixtures.RefreshTokenValue)

		resp, err := testApp.Test(req, -1)
		require.NoError(t, err)
		defer func() {
			if err := resp.Body.Close(); err != nil {
				t.Errorf("close response body: %v", err)
			}
		}()

		require.Equal(t, http.StatusCreated, resp.StatusCode)

		respBody, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		var got api.CreateTripDraftResponse
		require.NoError(t, json.Unmarshal(respBody, &got))
		require.Equal(t, fixtures.NormalClientID, got.DriverID)
		require.Equal(t, payload.FromPoint, got.FromPoint)
		require.Equal(t, payload.ToPoint, got.ToPoint)
		require.Equal(t, payload.AvailableSeats, got.Seats)
	})

	// Given: invalid body (fromPoint too short)
	// When: POST
	// Then: 400
	t.Run("bad_request_when_from_point_too_short", func(t *testing.T) {
		t.Parallel()
		lockIT(t)
		fixtures.UseStubClientID(t, fixtures.NormalClientID)

		payload := createTripRequestModel()
		payload.FromPoint = "too-short"

		body, err := json.Marshal(payload)
		require.NoError(t, err)

		req, err := http.NewRequest(http.MethodPost, createTripDraftURL, bytes.NewReader(body))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set(fixtures.RefreshTokenHeader, fixtures.RefreshTokenValue)

		resp, err := testApp.Test(req, -1)
		require.NoError(t, err)
		defer func() {
			if err := resp.Body.Close(); err != nil {
				t.Errorf("close response body: %v", err)
			}
		}()

		require.Equal(t, http.StatusBadRequest, resp.StatusCode)

		respBody, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		var apiResp api.Response
		require.NoError(t, json.Unmarshal(respBody, &apiResp))
		require.False(t, apiResp.Success)
		require.Equal(t, api.RequestValidationError, apiResp.Message)
	})

	// Given: stub does not inject claims into Locals
	// When: POST createTripDraft
	// Then: 401 from RequireClientRole (before handler) — plain Fiber error, not ErrResponse JSON
	// Note: UseStubClientID(uuid.Nil) still injects claims → would be 201, not 401
	t.Run("unauthorized_when_claims_not_found_in_context", func(t *testing.T) {
		t.Parallel()
		lockIT(t)
		fixtures.UseStubNoClaims(t)

		payload := createTripRequestModel()
		body, err := json.Marshal(payload)
		require.NoError(t, err)

		req, err := http.NewRequest(http.MethodPost, createTripDraftURL, bytes.NewReader(body))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set(fixtures.RefreshTokenHeader, fixtures.RefreshTokenValue)

		resp, err := testApp.Test(req, -1)
		require.NoError(t, err)
		defer func() {
			if err := resp.Body.Close(); err != nil {
				t.Errorf("close response body: %v", err)
			}
		}()

		require.Equal(t, http.StatusUnauthorized, resp.StatusCode)

		respBody, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		// fiber.NewError → text/JSON Fiber shape, not api.ErrResponse
		require.Contains(t, string(respBody), "missing token claims")
	})
}

func createTripRequestModel() api.CreateTripDraftRequest {
	return api.CreateTripDraftRequest{
		FromPoint:      "Moscow city, st. Big Road, h.10O",
		ToPoint:        "Moscow city, st. Big Road, h.10",
		DepartureTime:  time.Now(),
		AvailableSeats: 1,
	}
}
