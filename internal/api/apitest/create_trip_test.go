package apitest

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"time"

	"job4j.ru/share_trip/internal/api"
	"job4j.ru/share_trip/internal/api/apierr"
	"job4j.ru/share_trip/internal/api/apitest/fixtures"
	domainhttp "job4j.ru/share_trip/internal/domain/http"
	"job4j.ru/share_trip/internal/domain/trip/model"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

const createTripDraftURL = GroupPrefixV2 + "/trip/createTripDraft"

func TestServer_CreateTrip(t *testing.T) {

	// Given: valid request
	// When: sending request
	// Then: return 201 and created trip
	t.Run("success_create_trip_draft", func(t *testing.T) {
		fixtures.UseStubClientID(t, fixtures.NormalClientID)
		payload := createTripRequestModel()

		body, err := json.Marshal(payload)
		require.NoError(t, err)

		url := createTripDraftURL
		req, err := http.NewRequest(
			http.MethodPost,
			url,
			bytes.NewReader(body),
		)
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

		var got model.CreateTripDraftResponse
		err = json.Unmarshal(respBody, &got)
		require.NoError(t, err)
		response := model.CreateTripDraftResponse{
			ID:            got.ID,
			DriverID:      payload.DriverID,
			FromPoint:     got.FromPoint,
			ToPoint:       got.ToPoint,
			CreatedAt:     got.CreatedAt,
			DepartureTime: got.DepartureTime,
			Seats:         got.Seats,
			Status:        got.Status,
		}
		require.Equal(t, response, got)
	})

	// Given: invalid request
	// When: sending request
	// Then: return 400 and validation error
	t.Run("bad_request_when_driver_id_is_not_valid_uuid", func(t *testing.T) {
		fixtures.UseStubClientID(t, fixtures.NormalClientID)

		payload := createTripRequestModel()
		payload.DriverID = uuid.Nil

		body, err := json.Marshal(payload)
		require.NoError(t, err)

		url := createTripDraftURL
		req, err := http.NewRequest(
			http.MethodPost,
			url,
			bytes.NewReader(body),
		)
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

		var apiResp domainhttp.Response
		require.NoError(t, json.Unmarshal(respBody, &apiResp))
		require.False(t, apiResp.Success)
		require.Equal(t, apierr.RequestValidationError, apiResp.Message)
	})

	// Given: JWT sub = InvalidClientID, body.driverId = another user
	// When: POST createTripDraft
	// Then: 403 IDOR (driverId must match authenticated user)
	t.Run("forbidden_when_keycloak_client_id_does_not_match_authenticated_user", func(t *testing.T) {
		fixtures.UseStubClientID(t, fixtures.InvalidClientID)

		payload := createTripRequestModel()         // DriverID starts as NormalClientID

		body, err := json.Marshal(payload)
		require.NoError(t, err)
		req, err := http.NewRequest(
			http.MethodPost,
			createTripDraftURL,
			bytes.NewReader(body),
		)
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

		require.Equal(t, http.StatusForbidden, resp.StatusCode)

		respBody, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		var apiResp domainhttp.Response
		require.NoError(t, json.Unmarshal(respBody, &apiResp))
		require.False(t, apiResp.Success)
		require.Equal(t, apierr.ErrorForbiddenIDMismatch, apiResp.Message)
	})

}

func createTripRequestModel() api.CreateTripRequestModel {
	return api.CreateTripRequestModel{
		DriverID:       fixtures.NormalClientID, // must match Keycloak stub sub
		FromPoint:      "Mockov city, st. Big Road, h.10О",
		ToPoint:        "Mockov city, st. Big Road, h.10",
		DepartureTime:  time.Now(),
		AvailableSeats: 1,
	}
}
