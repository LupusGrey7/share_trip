package api

import (
	"github.com/google/uuid"
	"job4j.ru/share_trip/internal/domain/trip"
	"time"
)

type MoveTripDraftToPublishModelRequest struct {
	ID       string    `validate:"required,uuid"`
	ClientID uuid.UUID `json:"clientId" validate:"required,uuid"` //"omitempty,uuid"
}

func (req *MoveTripDraftToPublishModelRequest) ToRequest(id uuid.UUID) trip.MoveTripDraftToPublishModel {
	return trip.MoveTripDraftToPublishModel{
		ID:       id,
		ClientID: req.ClientID,
	}
}

type GetByIDModelRequest struct {
	ID string `validate:"required,uuid"`
}

type CreateTripRequest struct {
	DriverID       uuid.UUID `json:"driverId" validate:"required,uuid"`
	FromPoint      string    `json:"fromPoint" validate:"required,min=20,max=155"`
	ToPoint        string    `json:"toPoint" validate:"required,min=20,max=155"`
	DepartureTime  time.Time `json:"departureTime" validate:"required"`
	AvailableSeats int       `json:"seats" validate:"required,min=1,max=3"`
}

func (req *CreateTripRequest) ToCreateTripDomainRequest() trip.CreateTripRequest {
	return trip.CreateTripRequest{
		DriverID:       req.DriverID,
		FromPoint:      req.FromPoint,
		ToPoint:        req.ToPoint,
		DepartureTime:  req.DepartureTime,
		AvailableSeats: req.AvailableSeats,
	}
}
