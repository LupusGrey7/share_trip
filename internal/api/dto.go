package api

import (
	"time"

	"github.com/google/uuid"
	"job4j.ru/share_trip/internal/domain/trip/model"
)

type GetTripByIdRequestModel struct {
	ID string `validate:"required,uuid"`
}

type MoveTripDraftToPublishRequestModel struct {
	ID       string    `validate:"required,uuid"`
	ClientID uuid.UUID `json:"clientId" validate:"required,uuid"` //"omitempty,uuid"
}

type CreateTripRequestModel struct {
	DriverID       uuid.UUID `json:"driverId" validate:"required,uuid"`
	FromPoint      string    `json:"fromPoint" validate:"required,min=20,max=155"`
	ToPoint        string    `json:"toPoint" validate:"required,min=20,max=155"`
	DepartureTime  time.Time `json:"departureTime" validate:"required"`
	AvailableSeats int       `json:"seats" validate:"required,min=1,max=3"`
}

func (req *MoveTripDraftToPublishRequestModel) ToRequest() model.MoveTripDraftToPublishModel {
	return model.MoveTripDraftToPublishModel{
		ID:       req.ID,
		ClientID: req.ClientID,
	}
}
