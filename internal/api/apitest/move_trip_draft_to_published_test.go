package apitest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"job4j.ru/share_trip/internal/domain/trip/model"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"job4j.ru/share_trip/internal/api"
)

func TestServer_MoveTripDraftToPublish(t *testing.T) {

	t.Run("success - обновление поездки", func(t *testing.T) {
		payload := api.CreateTripRequestModel{
			DriverID:       uuid.New(),
			FromPoint:      "Mockov city, st. Big Star, h.10О",
			ToPoint:        "Mockov city, st. Dig Star, h.10",
			DepartureTime:  time.Now(),
			AvailableSeats: 1,
		}

		body, err := json.Marshal(payload)
		require.NoError(t, err)

		url := GroupPrefixV2 + "/trip/createTripDraft"
		req, err := http.NewRequest(
			http.MethodPost,
			url,
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

		//---update
		publishModelRequest := api.MoveTripDraftToPublishRequestModel{
			ClientID: payload.DriverID,
		}

		marshalBody, err1 := json.Marshal(publishModelRequest)
		require.NoError(t, err1)

		urlUpdate := GroupPrefixV2 + fmt.Sprintf("/trip/moveTripDraft-ToPublish/%s", got.ID)
		req, err1 = http.NewRequest(
			http.MethodPatch,
			urlUpdate,
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

		var got1 model.MoveTripDraftToPublishModelResponse
		err1 = json.Unmarshal(respBody, &got1)
		require.NoError(t, err1)

		t.Logf("Response body2: %s", string(respBody)) // Выведем тело ответа для диагностики

		response1 := model.MoveTripDraftToPublishModelResponse{
			ID:            got1.ID,
			DriverID:      payload.DriverID,
			FromPoint:     got1.FromPoint,
			ToPoint:       got1.ToPoint,
			CreatedAt:     got1.CreatedAt,
			DepartureTime: got1.DepartureTime,
			Seats:         got1.Seats,
			Status:        model.StatusPublished,
		}

		require.Equal(t, response1, got1)
	})
	t.Run("forbidden - обновление поездки", func(t *testing.T) { //403
		payload := api.CreateTripRequestModel{
			DriverID:       uuid.New(),
			FromPoint:      "Mockov city, st. Big Star, h.10О",
			ToPoint:        "Mockov city, st. Dig Star, h.10",
			DepartureTime:  time.Now(),
			AvailableSeats: 1,
		}

		body, err := json.Marshal(payload)
		require.NoError(t, err)

		url := GroupPrefixV2 + "/trip/createTripDraft"
		req, err := http.NewRequest(
			http.MethodPost,
			url,
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
		t.Logf("Response ID: %v", got.ID)

		//---update
		uuID, err := uuid.Parse("d4733715-0fc7-42fa-b13a-f068e33c6d80")
		if err != nil {
			t.Errorf("err parse uuid: %v", err)
		}
		publishModelRequest := api.MoveTripDraftToPublishRequestModel{
			ClientID: uuID,
		}

		marshalBody, err1 := json.Marshal(publishModelRequest)
		require.NoError(t, err1)

		t.Logf("/trip/moveTripDraft-ToPublish/%s\n", got.ID)

		urlUpdate := GroupPrefixV2 + fmt.Sprintf("/trip/moveTripDraft-ToPublish/%s", got.ID)
		req, err1 = http.NewRequest(
			http.MethodPatch,
			urlUpdate,
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
		require.Equal(t, http.StatusForbidden, resp2.StatusCode)

	})
	t.Run("statusNotFound - обновление поездки", func(t *testing.T) { //404
		payload := api.CreateTripRequestModel{
			DriverID:       uuid.New(),
			FromPoint:      "Mockov city, st. Big Star, h.10О",
			ToPoint:        "Mockov city, st. Dig Star, h.10",
			DepartureTime:  time.Now(),
			AvailableSeats: 1,
		}

		body, err := json.Marshal(payload)
		require.NoError(t, err)

		url := GroupPrefixV2 + "/trip/createTripDraft"
		req, err := http.NewRequest(
			http.MethodPost,
			url,
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
		t.Logf("Response ID: %v", got.ID)

		//---update
		uuID := "d4733715-0fc7-42fa-b13a-f068e33c6d80"
		publishModelRequest := api.MoveTripDraftToPublishRequestModel{
			ClientID: payload.DriverID,
		}
		marshalBody, err1 := json.Marshal(publishModelRequest)

		require.NoError(t, err1)

		urlUpdate := GroupPrefixV2 + fmt.Sprintf("/trip/moveTripDraft-ToPublish/%s", uuID)
		req, err1 = http.NewRequest(
			http.MethodPatch,
			urlUpdate,
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
		require.Equal(t, http.StatusNotFound, resp2.StatusCode)
	})
	t.Run("StatusConflict - обновление поездки", func(t *testing.T) { //409
		payload := api.CreateTripRequestModel{
			DriverID:       uuid.New(),
			FromPoint:      "Mockov city, st. Big Star, h.10О",
			ToPoint:        "Mockov city, st. Dig Star, h.10",
			DepartureTime:  time.Now(),
			AvailableSeats: 1,
		}

		body, err := json.Marshal(payload)
		require.NoError(t, err)

		url := GroupPrefixV2 + "/trip/createTripDraft"
		req, err := http.NewRequest(
			http.MethodPost,
			url,
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
		t.Logf("Response ID: %v", got.ID)

		//---update
		// Принудительно меняем статус созданной поездки на "cancelled"
		_, err = testDB.ExecContext(testCtx, "UPDATE trips SET status = $1 WHERE id = $2", "cancelled", got.ID)
		require.NoError(t, err)
		publishModelRequest := api.MoveTripDraftToPublishRequestModel{
			ClientID: payload.DriverID,
		}

		marshalBody, err1 := json.Marshal(publishModelRequest)
		require.NoError(t, err1)

		urlUpdate := GroupPrefixV2 + fmt.Sprintf("/trip/moveTripDraft-ToPublish/%s", got.ID)
		req, err1 = http.NewRequest(
			http.MethodPatch,
			urlUpdate,
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
		require.Equal(t, http.StatusConflict, resp2.StatusCode)
	})
	t.Run("internalServerError - обновление поездки", func(t *testing.T) { // 500
		payload := api.CreateTripRequestModel{
			DriverID:       uuid.New(),
			FromPoint:      "Mockov city, st. Big Star, h.10О",
			ToPoint:        "Mockov city, st. Dig Star, h.10",
			DepartureTime:  time.Now(),
			AvailableSeats: 1,
		}

		body, err := json.Marshal(payload)
		require.NoError(t, err)

		url := GroupPrefixV2 + "/trip/createTripDraft"
		req, err := http.NewRequest(
			http.MethodPost,
			url,
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
		t.Logf("Response ID: %v", got.ID)

		//---update
		publishModelRequest := api.MoveTripDraftToPublishRequestModel{
			ClientID: uuid.Nil,
		}

		marshalBody, err1 := json.Marshal(publishModelRequest)
		require.NoError(t, err1)

		urlUpdate := GroupPrefixV2 + fmt.Sprintf("/trip/moveTripDraft-ToPublish/%s", got.ID)
		req, err1 = http.NewRequest(
			http.MethodPatch,
			urlUpdate,
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
		require.Equal(t, http.StatusInternalServerError, resp2.StatusCode)

	})
}
