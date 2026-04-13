package apitest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"job4j.ru/share_trip/internal/api"
	"job4j.ru/share_trip/internal/domain/trip"
)

func TestServer_MoveTripDraftToPublish(t *testing.T) {

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
			"/trip/createTripDraft",
			bytes.NewReader(body),
		)
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		//Отправляем запрос в приложение
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
		fmt.Println("-00->", got.ID)

		//---update
		publishModelRequest := api.MoveTripDraftToPublishModelRequest{
			ClientID: payload.DriverID,
		}

		marshalBody, err1 := json.Marshal(publishModelRequest)
		require.NoError(t, err1)
		fmt.Println("-->", got.ID)
		fmt.Printf("/trip/moveTripDraft-ToPublish/%s\n", got.ID)

		req, err1 = http.NewRequest(
			http.MethodPatch,
			fmt.Sprintf("/trip/moveTripDraft-ToPublish/%s", got.ID),
			bytes.NewReader(marshalBody),
		)
		require.NoError(t, err1)
		req.Header.Set("Content-Type", "application/json")

		//Отправляем запрос в приложение
		resp2, err2 := testApp.Test(req, -1)
		require.NoError(t, err2)
		defer func() {
			if err := resp2.Body.Close(); err != nil {
				t.Errorf("close response body: %v", err)
			}
		}()

		t.Logf("Response body: %v", resp2)
		require.Equal(t, http.StatusOK, resp2.StatusCode)

		respBody, err2 = io.ReadAll(resp2.Body)
		require.NoError(t, err2)

		var got1 trip.MoveTripDraftToPublishModelResponse
		err1 = json.Unmarshal(respBody, &got1)
		require.NoError(t, err1)

		t.Logf("Response body2: %s", string(respBody)) // Выведем тело ответа для диагностики

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
