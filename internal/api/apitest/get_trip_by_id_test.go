package apitest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"job4j.ru/share_trip/internal/api"
	"job4j.ru/share_trip/internal/api/apitest/fixtures"
	"job4j.ru/share_trip/internal/domain/trip/model"

	domainhttp "job4j.ru/share_trip/internal/domain/http"
	"job4j.ru/share_trip/internal/api/apierr"

	"github.com/stretchr/testify/require"
)

func TestServer_GetTripById(t *testing.T) {

	t.Run("success_get_trip_by_id", func(t *testing.T) {
		fixtures.UseStubClientID(t, fixtures.NormalClientID)

		payload := createTripDraft()

		body, err := json.Marshal(payload)
		require.NoError(t, err)

		url := GroupPrefixV2 + "/trip/createTripDraft"
		req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

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

		urlCheck := GroupPrefixV2 + fmt.Sprintf("/trip/%s", got.ID)
		req, err1 := http.NewRequest(http.MethodGet, urlCheck, nil)
		require.NoError(t, err1)
		req.Header.Set("Content-Type", "application/json")

		resp, err1 = testApp.Test(req, -1)
		require.NoError(t, err1)
		defer func() {
			if err := resp.Body.Close(); err != nil {
				t.Errorf("close response body: %v", err)
			}
		}()

		require.Equal(t, http.StatusOK, resp.StatusCode)

		respBody, err1 = io.ReadAll(resp.Body)
		require.NoError(t, err1)

		var got1 model.CreateTripDraftResponse
		err1 = json.Unmarshal(respBody, &got1)
		require.NoError(t, err1)
		response1 := model.CreateTripDraftResponse{
			ID:            got1.ID,
			DriverID:      got.DriverID,
			FromPoint:     got1.FromPoint,
			ToPoint:       got1.ToPoint,
			CreatedAt:     got1.CreatedAt,
			DepartureTime: got1.DepartureTime,
			Seats:         got1.Seats,
			Status:        model.StatusDraft,
		}

		require.Equal(t, response1, got1)
	})

	// Given: trip created by NormalClientID
	// When: another user (InvalidClientID) calls GET
	// Then: 403 ownership / IDOR guard
	t.Run("forbidden_when_caller_is_not_trip_owner", func(t *testing.T) {
		fixtures.UseStubClientID(t, fixtures.NormalClientID)

		createTripPayload := createTripDraft()
		body, err := json.Marshal(createTripPayload)
		require.NoError(t, err)

		url := GroupPrefixV2 + "/trip/createTripDraft"
		req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

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

		var created model.CreateTripDraftResponse
		require.NoError(t, json.Unmarshal(respBody, &created))
		require.Equal(t, fixtures.NormalClientID, created.DriverID)

		// Switch "logged-in user" without rebuilding Fiber / routes.
		// Middleware already registered in TestMain; it reads stubSubject on each request.
		fixtures.SetStubClientID(fixtures.InvalidClientID)

		urlCheck := GroupPrefixV2 + fmt.Sprintf("/trip/%s", created.ID)
		getReq, err := http.NewRequest(http.MethodGet, urlCheck, nil)
		require.NoError(t, err)
		getReq.Header.Set("Content-Type", "application/json")

		getResp, err := testApp.Test(getReq, -1)
		require.NoError(t, err)
		defer func() {
			if err := getResp.Body.Close(); err != nil {
				t.Errorf("close response body: %v", err)
			}
		}()

		require.Equal(t, http.StatusForbidden, getResp.StatusCode)

		getBody, err := io.ReadAll(getResp.Body)
		require.NoError(t, err)

		var apiResp domainhttp.Response
		require.NoError(t, json.Unmarshal(getBody, &apiResp))
		require.False(t, apiResp.Success)
		require.Equal(t, apierr.ErrorForbiddenIDMismatch, apiResp.Message)
	})
}

func createTripDraft() api.CreateTripRequestModel {
	return api.CreateTripRequestModel{
		DriverID:       fixtures.NormalClientID, // must match stub sub on create
		FromPoint:      "Mockov city, st. Big Star, h.101О",
		ToPoint:        "Mockov city, st. Dig Star, h.101",
		DepartureTime:  time.Now(),
		AvailableSeats: 1,
	}
}
