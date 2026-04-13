package apitest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"io"
	"job4j.ru/share_trip/internal/domain/trip"
	"net/http"
	"testing"
	"time"
)

func TestServer_UpdateTrip(t *testing.T) {

	t.Run("success - обновление поездки", func(t *testing.T) {
		payload := trip.CreateTripRequest{
			DriverID:       uuid.New(),
			FromPoint:      "Mockov city, st. Big Star, h.10О",
			ToPoint:        "Mockov city, st. Dig Star, h.10",
			DepartureTime:  time.Now(),
			AvailableSeats: 1,
		}

		body, err := json.Marshal(payload)
		require.NoError(t, err)

		req, err := http.NewRequest(
			http.MethodPost,
			"/trip/create",
			bytes.NewReader(body),
		)
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

		var got trip.CreateTripResponse
		err = json.Unmarshal(respBody, &got)
		require.NoError(t, err)
		response := trip.CreateTripResponse{
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

		//---update
		payload1 := trip.MoveTripDraftToPublishModelRequest{
			ClientID: payload.DriverID,
		}

		body1, err1 := json.Marshal(payload1)
		require.NoError(t, err1)

		req, err1 = http.NewRequest(
			http.MethodPatch,
			fmt.Sprintf("/trip/%s", got.ID),
			bytes.NewReader(body1),
		)
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

		var got1 trip.MoveTripDraftToPublishModelResponse
		err1 = json.Unmarshal(respBody, &got1)
		require.NoError(t, err1)
		response1 := trip.MoveTripDraftToPublishModelResponse{
			ID:            got1.ID,
			DriverID:      payload.DriverID,
			FromPoint:     got1.FromPoint,
			ToPoint:       got1.ToPoint,
			CreatedAt:     got1.CreatedAt,
			DepartureTime: got1.DepartureTime,
			Seats:         got1.Seats,
			Status:        trip.StatusPublished,
		}

		require.Equal(t, response1, got1)
	})
}
