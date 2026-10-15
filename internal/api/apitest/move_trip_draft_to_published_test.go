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

	"job4j.ru/share_trip/internal/trip/domain"

	"github.com/stretchr/testify/require"
)

const (
	testCompanyIDForPublish      = "acme01"
	createTripDraftURLForPublish = GroupPrefixV2 + "/trip/createTripDraft"
	moveTripDraftToPublishURL    = GroupPrefixV2 + "/trip/moveTripDraft-ToPublish/%s/company/%s"
)

func TestServer_MoveTripFromDraftToPublish(t *testing.T) {
	t.Parallel()

	// Given: trip draft owned by NormalClientID
	// When: owner publishes (identity = JWT sub, empty body)
	// Then: 200 + status published
	t.Run("success_when_caller_is_trip_owner", func(t *testing.T) {
		t.Parallel()
		lockIT(t)
		fixtures.UseStubClientID(t, fixtures.NormalClientID)

		created := mustCreateTripDraft(t, createTripDraftRequestModel())

		resp := mustPublishTripDraft(t, created.ID.String())
		defer func() {
			if err := resp.Body.Close(); err != nil {
				t.Errorf("close response body: %v", err)
			}
		}()

		require.Equal(t, http.StatusOK, resp.StatusCode)

		respBody, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		var got api.MoveTripFromDraftToPublishResponse
		require.NoError(t, json.Unmarshal(respBody, &got))

		want := api.MoveTripFromDraftToPublishResponse{
			ID:            got.ID,
			DriverID:      fixtures.NormalClientID,
			FromPoint:     got.FromPoint,
			ToPoint:       got.ToPoint,
			CreatedAt:     got.CreatedAt,
			DepartureTime: got.DepartureTime,
			Seats:         got.Seats,
			Status:        api.StatusEnum("published"),
		}
		require.Equal(t, want, got)

		var (
			aggregateType string
			eventType     string
			status        string
			attempts      int
			payload       []byte
		)
		err = testPool.QueryRow(testCtx, `
			select aggregate_type, event_type, status, attempts, payload
			from outbox_events
			where aggregate_id = $1
			order by created_at desc
			limit 1
		`, created.ID).Scan(&aggregateType, &eventType, &status, &attempts, &payload)
		require.NoError(t, err)
		require.Equal(t, "trip", aggregateType)
		require.Equal(t, "trip_published", eventType)
		require.Equal(t, "pending", status)
		require.Equal(t, 0, attempts)

		var payloadMap map[string]string
		require.NoError(t, json.Unmarshal(payload, &payloadMap))
		require.Equal(t, created.ID.String(), payloadMap["trip_id"])
		require.Equal(t, fixtures.NormalClientID.String(), payloadMap["driver_id"])
		require.Equal(t, testCompanyIDForPublish, payloadMap["company_id"])
	})

	// Given: trip draft owned by NormalClientID
	// When: another user (InvalidClientID) calls publish with their token
	// Then: 403 (use case ownership check)
	t.Run("forbidden_when_caller_is_not_trip_owner", func(t *testing.T) {
		t.Parallel()
		lockIT(t)
		fixtures.UseStubClientID(t, fixtures.NormalClientID)

		created := mustCreateTripDraft(t, createTripDraftRequestModel())

		fixtures.UseStubClientID(t, fixtures.InvalidClientID)
		resp := mustPublishTripDraft(t, created.ID.String())
		defer func() {
			if err := resp.Body.Close(); err != nil {
				t.Errorf("close response body: %v", err)
			}
		}()

		require.Equal(t, http.StatusForbidden, resp.StatusCode)

		respBody, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		var apiResp api.Response
		require.NoError(t, json.Unmarshal(respBody, &apiResp))
		require.False(t, apiResp.Success)
		wantMsg := fmt.Sprintf(
			"forbidden: client %s is not driver of trip %s",
			fixtures.InvalidClientID,
			created.ID,
		)
		require.Equal(t, wantMsg, apiResp.Message)
	})

	// Given: trip id does not exist
	// When: owner publishes
	// Then: 404
	t.Run("not_found_when_trip_does_not_exist", func(t *testing.T) {
		t.Parallel()
		lockIT(t)
		fixtures.UseStubClientID(t, fixtures.NormalClientID)

		nonExistentTripID := "00000000-0000-0000-0000-000000000001"
		resp := mustPublishTripDraft(t, nonExistentTripID)
		defer func() {
			if err := resp.Body.Close(); err != nil {
				t.Errorf("close response body: %v", err)
			}
		}()

		require.Equal(t, http.StatusNotFound, resp.StatusCode)

		respBody, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		var apiResp api.Response
		require.NoError(t, json.Unmarshal(respBody, &apiResp))
		require.False(t, apiResp.Success)
		require.Equal(t, api.StatusNotFound, apiResp.Message)
	})

	// Given: trip status forced to cancelled
	// When: owner publishes
	// Then: 409
	t.Run("conflict_when_trip_is_cancelled", func(t *testing.T) {
		t.Parallel()
		lockIT(t)
		fixtures.UseStubClientID(t, fixtures.NormalClientID)

		created := mustCreateTripDraft(t, createTripDraftRequestModel())

		_, err := testDB.ExecContext(testCtx,
			`UPDATE trips SET trip_status_id = (SELECT id FROM trip_status WHERE name = $1) WHERE id = $2`,
			"cancelled", created.ID,
		)
		require.NoError(t, err)

		resp := mustPublishTripDraft(t, created.ID.String())
		defer func() {
			if err := resp.Body.Close(); err != nil {
				t.Errorf("close response body: %v", err)
			}
		}()

		require.Equal(t, http.StatusConflict, resp.StatusCode)

		respBody, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		var apiResp api.Response
		require.NoError(t, json.Unmarshal(respBody, &apiResp))
		require.False(t, apiResp.Success)
		wantMsg := fmt.Sprintf("tx block: conflict: invalid entity status: expected %s", domain.StatusDraft)
		require.Equal(t, wantMsg, apiResp.Message)
	})

	// Given: no claims in context
	// When: publish
	// Then: 401
	t.Run("unauthorized_without_claims", func(t *testing.T) {
		t.Parallel()
		lockIT(t)
		fixtures.UseStubClientID(t, fixtures.NormalClientID)

		created := mustCreateTripDraft(t, createTripDraftRequestModel())

		fixtures.UseStubNoClaims(t)
		resp := mustPublishTripDraft(t, created.ID.String())
		defer func() {
			if err := resp.Body.Close(); err != nil {
				t.Errorf("close response body: %v", err)
			}
		}()

		require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	// Given: trip already published
	// When: owner publishes again
	// Then: 204 (ErrAlreadyDone), no body
	t.Run("no_content_when_trip_already_published", func(t *testing.T) {
		t.Parallel()
		lockIT(t)
		fixtures.UseStubClientID(t, fixtures.NormalClientID)

		created := mustCreateTripDraft(t, createTripDraftRequestModel())

		first := mustPublishTripDraft(t, created.ID.String())
		require.Equal(t, http.StatusOK, first.StatusCode)
		_ = first.Body.Close()

		second := mustPublishTripDraft(t, created.ID.String())
		defer func() {
			if err := second.Body.Close(); err != nil {
				t.Errorf("close response body: %v", err)
			}
		}()

		require.Equal(t, http.StatusNoContent, second.StatusCode)
	})
}

func createTripDraftRequestModel() api.CreateTripDraftRequest {
	return api.CreateTripDraftRequest{
		FromPoint:      "Mockov city, st. Big Street, h.101",
		ToPoint:        "Mockov city, st. Big Street, h.10O",
		DepartureTime:  time.Now(),
		AvailableSeats: 1,
	}
}

func mustCreateTripDraft(t *testing.T, payload api.CreateTripDraftRequest) api.CreateTripDraftResponse {
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

	var created api.CreateTripDraftResponse
	require.NoError(t, json.Unmarshal(respBody, &created))
	require.Equal(t, fixtures.CurrentStubClientID(), created.DriverID)
	return created
}

// mustPublishTripDraft — PATCH publish; identity from JWT stub, body empty (lead: no clientId in JSON).
func mustPublishTripDraft(t *testing.T, tripID string) *http.Response {
	t.Helper()

	req, err := http.NewRequest(
		http.MethodPatch,
		fmt.Sprintf(moveTripDraftToPublishURL, tripID, testCompanyIDForPublish),
		nil,
	)
	require.NoError(t, err)
	req.Header.Set(fixtures.RefreshTokenHeader, fixtures.RefreshTokenValue)

	resp, err := testApp.Test(req, -1)
	require.NoError(t, err)
	return resp
}
