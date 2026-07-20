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
	"job4j.ru/share_trip/internal/api/apierr"
	"job4j.ru/share_trip/internal/api/apitest/fixtures"
	domainhttp "job4j.ru/share_trip/internal/domain/http"
	"job4j.ru/share_trip/internal/domain/trip/model"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

const (
	createTripDraftURLForPublish = GroupPrefixV2 + "/trip/createTripDraft"
	moveTripDraftToPublishURL    = GroupPrefixV2 + "/trip/moveTripDraft-ToPublish/%s"
)

func TestServer_MoveTripDraftToPublish(t *testing.T) {
	// Given: trip draft owned by NormalClientID
	// When: owner publishes
	// Then: 200 + status published
	t.Run("success_when_caller_is_trip_owner", func(t *testing.T) {
		fixtures.UseStubClientID(t, fixtures.NormalClientID)

		created := mustCreateTripDraft(t, createTripDraftRequestModel())

		publishBody := api.MoveTripDraftToPublishRequestModel{
			ClientID: fixtures.NormalClientID,
		}
		resp := mustPublishTripDraft(t, created.ID.String(), publishBody)
		defer func() {
			if err := resp.Body.Close(); err != nil {
				t.Errorf("close response body: %v", err)
			}
		}()

		require.Equal(t, http.StatusOK, resp.StatusCode)

		respBody, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		var got model.MoveTripDraftToPublishModelResponse
		require.NoError(t, json.Unmarshal(respBody, &got))

		want := model.MoveTripDraftToPublishModelResponse{
			ID:            got.ID,
			DriverID:      fixtures.NormalClientID,
			FromPoint:     got.FromPoint,
			ToPoint:       got.ToPoint,
			CreatedAt:     got.CreatedAt,
			DepartureTime: got.DepartureTime,
			Seats:         got.Seats,
			Status:        model.StatusPublished,
		}
		require.Equal(t, want, got)
	})

	// Given: trip draft owned by NormalClientID
	// When: body.clientId is another user
	// Then: 403 (use case ownership check)
	t.Run("forbidden_when_client_id_is_not_trip_owner", func(t *testing.T) {
		fixtures.UseStubClientID(t, fixtures.NormalClientID)

		created := mustCreateTripDraft(t, createTripDraftRequestModel())

		publishBody := api.MoveTripDraftToPublishRequestModel{
			ClientID: fixtures.InvalidClientID,
		}
		resp := mustPublishTripDraft(t, created.ID.String(), publishBody)
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
		// HandleError Unwrap'ит tx-обёртку → текст use case:
		// fmt.Errorf("%w: client %s is not driver of trip %s", ErrForbidden, ...)
		wantMsg := fmt.Sprintf(
			"forbidden: client %s is not driver of trip %s",
			fixtures.InvalidClientID,
			created.ID,
		)
		require.Equal(t, wantMsg, apiResp.Message)
	})

	// Given: valid clientId, trip id does not exist
	// When: publish
	// Then: 404
	t.Run("not_found_when_trip_does_not_exist", func(t *testing.T) {
		fixtures.UseStubClientID(t, fixtures.NormalClientID)

		nonExistentTripID := "00000000-0000-0000-0000-000000000001"
		publishBody := api.MoveTripDraftToPublishRequestModel{
			ClientID: fixtures.NormalClientID,
		}
		resp := mustPublishTripDraft(t, nonExistentTripID, publishBody)
		defer func() {
			if err := resp.Body.Close(); err != nil {
				t.Errorf("close response body: %v", err)
			}
		}()

		require.Equal(t, http.StatusNotFound, resp.StatusCode)

		respBody, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		var apiResp domainhttp.Response
		require.NoError(t, json.Unmarshal(respBody, &apiResp))
		require.False(t, apiResp.Success)
		// ErrTripNotFound → константа StatusNotFound (без id в message)
		require.Equal(t, apierr.StatusNotFound, apiResp.Message)
	})

	// Given: trip status forced to cancelled
	// When: owner publishes
	// Then: 409
	t.Run("conflict_when_trip_is_cancelled", func(t *testing.T) {
		fixtures.UseStubClientID(t, fixtures.NormalClientID)

		created := mustCreateTripDraft(t, createTripDraftRequestModel())

		_, err := testDB.ExecContext(testCtx, "UPDATE trips SET status = $1 WHERE id = $2", "cancelled", created.ID)
		require.NoError(t, err)

		publishBody := api.MoveTripDraftToPublishRequestModel{
			ClientID: fixtures.NormalClientID,
		}
		resp := mustPublishTripDraft(t, created.ID.String(), publishBody)
		defer func() {
			if err := resp.Body.Close(); err != nil {
				t.Errorf("close response body: %v", err)
			}
		}()

		require.Equal(t, http.StatusConflict, resp.StatusCode)

		respBody, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		var apiResp domainhttp.Response
		require.NoError(t, json.Unmarshal(respBody, &apiResp))
		require.False(t, apiResp.Success)
		// Unwrap tx → use case: "%w: invalid entity status: expected %s"
		wantMsg := fmt.Sprintf("conflict: invalid entity status: expected %s", model.StatusDraft)
		require.Equal(t, wantMsg, apiResp.Message)
	})

	// Given: clientId = uuid.Nil (fails validate required,uuid)
	// When: publish
	// Then: 400 (HandleError ErrInvalidValidate), not 500
	t.Run("bad_request_when_client_id_is_nil_uuid", func(t *testing.T) {
		fixtures.UseStubClientID(t, fixtures.NormalClientID)

		created := mustCreateTripDraft(t, createTripDraftRequestModel())

		publishBody := api.MoveTripDraftToPublishRequestModel{
			ClientID: uuid.Nil,
		}
		resp := mustPublishTripDraft(t, created.ID.String(), publishBody)
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
}

func createTripDraftRequestModel() api.CreateTripRequestModel {
	return api.CreateTripRequestModel{
		DriverID:       fixtures.NormalClientID, // must match Keycloak stub sub
		FromPoint:      "Mockov city, st. Big Street, h.101",
		ToPoint:        "Mockov city, st. Big Street, h.10O",
		DepartureTime:  time.Now(),
		AvailableSeats: 1,
	}
}

func mustCreateTripDraft(t *testing.T, payload api.CreateTripRequestModel) model.CreateTripDraftResponse {
	t.Helper()

	body, err := json.Marshal(payload)
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPost, createTripDraftURLForPublish, bytes.NewReader(body))
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

	var created model.CreateTripDraftResponse
	require.NoError(t, json.Unmarshal(respBody, &created))
	require.Equal(t, payload.DriverID, created.DriverID)
	return created
}

func mustPublishTripDraft(
	t *testing.T,
	tripID string,
	body api.MoveTripDraftToPublishRequestModel,
) *http.Response {
	t.Helper()

	marshalBody, err := json.Marshal(body)
	require.NoError(t, err)

	req, err := http.NewRequest(
		http.MethodPatch,
		fmt.Sprintf(moveTripDraftToPublishURL, tripID),
		bytes.NewReader(marshalBody),
	)
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(fixtures.RefreshTokenHeader, fixtures.RefreshTokenValue)

	resp, err := testApp.Test(req, -1)
	require.NoError(t, err)
	return resp
}
