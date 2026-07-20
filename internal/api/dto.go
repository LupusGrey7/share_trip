package api

import (
	"time"

	"github.com/google/uuid"
	"job4j.ru/share_trip/internal/domain/trip/model"
)

type GetTripByIDRequestModel struct {
	ID string `validate:"required,uuid"`
}

func (req GetTripByIDRequestModel) ToGetByIDRequestModel() model.GetByIDModelRequest {
	return model.GetByIDModelRequest{ID: req.ID}
}

type MoveTripDraftToPublishRequestModel struct {
	ID       string    `validate:"required,uuid"`
	ClientID uuid.UUID `json:"clientId" validate:"required,uuid"`
}

type CreateTripRequestModel struct {
	DriverID       uuid.UUID `json:"driverId" validate:"required,uuid"`
	FromPoint      string    `json:"fromPoint" validate:"required,min=20,max=155"`
	ToPoint        string    `json:"toPoint" validate:"required,min=20,max=155"`
	DepartureTime  time.Time `json:"departureTime" validate:"required"`
	AvailableSeats int       `json:"seats" validate:"required,min=1,max=4"`
}

func (req *CreateTripRequestModel) ToCreateTripRequestModel() model.CreateTripRequestModel {
	return model.CreateTripRequestModel{
		DriverID:       req.DriverID, // source of truth = Keycloak sub (same as body after check)
		FromPoint:      req.FromPoint,
		ToPoint:        req.ToPoint,
		DepartureTime:  req.DepartureTime,
		AvailableSeats: req.AvailableSeats,
	}
}

func (req *MoveTripDraftToPublishRequestModel) ToMoveTripDraftToPublishModel() model.MoveTripDraftToPublishModel {
	return model.MoveTripDraftToPublishModel{
		ID:       req.ID,
		ClientID: req.ClientID,
	}
}

type CreateTripRequest struct {
	DriverID       uuid.UUID `json:"driverId" validate:"required,uuid"`
	FromPoint      string    `json:"fromPoint" validate:"required,min=20,max=155"`
	ToPoint        string    `json:"toPoint" validate:"required,min=20,max=155"`
	DepartureTime  time.Time `json:"departureTime" validate:"required"`
	AvailableSeats int       `json:"seats" validate:"required,min=1,max=3"`
}
