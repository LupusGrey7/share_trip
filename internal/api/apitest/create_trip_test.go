package apitest

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"job4j.ru/share_trip/internal/domain/trip"
)

/*
*
=== Registered Routes ===
GET    /ready
GET    /trip/:tripId
HEAD   /ready
HEAD   /trip/:tripId
POST   /trip/createTripDraft
PATCH  /trip/moveTripDraft-ToPublish/:tripId
*/
func TestServer_CreateTrip(t *testing.T) {

	t.Run("success - создание поездки", func(t *testing.T) {
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
			"/trip/createTripDraft",
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
	})
}
