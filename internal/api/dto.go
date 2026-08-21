package api

import (
	"time"

	"github.com/google/uuid"
	"job4j.ru/share_trip/internal/trip/domain"
)

type ServiceCodeEnum string

const (
	ServiceCodeTripPublication ServiceCodeEnum = "trip_publication" // ServiceCode for trip publication
	ServiceCodeTripStart       ServiceCodeEnum = "trip_start"       // ServiceCode for trip start
)

type GetTripByIDRequestModel struct {
	ID string `validate:"required,uuid"`
}

func (req GetTripByIDRequestModel) ToGetByIDRequestModel() domain.GetByIDModelRequest {
	return domain.GetByIDModelRequest{ID: req.ID}
}

type MoveTripDraftToPublishRequestModel struct {
	ID       string    `validate:"required,uuid"`
	ClientID uuid.UUID `json:"clientId" validate:"required,uuid"`
}

type MoveTripPublishedToStartedRequestModel struct {
	ID          string          `validate:"required,uuid"`             // trip id (path)
	CompanyID   string          `validate:"required,min=2,max=10"`     // company id (path)
	ServiceCode ServiceCodeEnum `validate:"required,oneof=trip_start"` // service code (path)
}

func (req MoveTripPublishedToStartedRequestModel) ToMoveTripPublishedToStartedModel() domain.MoveTripPublishedToStartedModel {
	return domain.MoveTripPublishedToStartedModel{
		ID:          req.ID,
		CompanyID:   req.CompanyID,
		ServiceCode: domain.ServiceCodeEnum(req.ServiceCode),
	}
}

type CreateTripRequestModel struct {
	FromPoint      string    `json:"fromPoint" validate:"required,min=20,max=155"`
	ToPoint        string    `json:"toPoint" validate:"required,min=20,max=155"`
	DepartureTime  time.Time `json:"departureTime" validate:"required"`
	AvailableSeats int       `json:"seats" validate:"required,min=1,max=4"`
}

// ToCreateTripRequestModel maps HTTP body → domain.
// DriverID is NOT taken from body — handler sets it from Keycloak sub.
func (req *CreateTripRequestModel) ToCreateTripRequestModel() domain.CreateTripRequestModel {
	return domain.CreateTripRequestModel{
		FromPoint:      req.FromPoint,
		ToPoint:        req.ToPoint,
		DepartureTime:  req.DepartureTime,
		AvailableSeats: req.AvailableSeats,
	}
}

func (req *MoveTripDraftToPublishRequestModel) ToMoveTripDraftToPublishModel() domain.MoveTripDraftToPublishModel {
	return domain.MoveTripDraftToPublishModel{
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
