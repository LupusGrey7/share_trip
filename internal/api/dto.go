package api

import (
	"time"

	"github.com/google/uuid"
)

type StatusEnum string
type ServiceCodeEnum string

const (
	ServiceCodeTripPublication  ServiceCodeEnum = "trip_publication"
	ServiceCodeTripCancellation ServiceCodeEnum = "trip_cancellation"
	ServiceCodeTripCompletion   ServiceCodeEnum = "trip_completion"
	ServiceCodeTripStart        ServiceCodeEnum = "trip_start"
	ServiceCodeTripEnd          ServiceCodeEnum = "trip_end"
)

type GetTripByIDRequest struct {
	ID string `validate:"required,uuid" param:"tripId"`
}

type MoveTripDraftToPublishRequest struct {
	ID       string    `validate:"required,uuid"`
	ClientID uuid.UUID `json:"clientId" validate:"required,uuid"`
}

type MoveTripPublishedToStartedRequest struct {
	ID          string          `param:"tripId" validate:"required,uuid"`
	CompanyID   string          `param:"companyId" validate:"required,min=2,max=10"`
	ServiceCode ServiceCodeEnum `param:"serviceCode" validate:"required,oneof=trip_start"`
	DriverID    uuid.UUID       `validate:"required,uuid"` // из Keycloak middleware, не из path
}

type CreateTripDraftRequest struct {
	FromPoint      string    `json:"fromPoint" validate:"required,min=20,max=155"`
	ToPoint        string    `json:"toPoint" validate:"required,min=20,max=155"`
	DepartureTime  time.Time `json:"departureTime" validate:"required"`
	AvailableSeats int       `json:"seats" validate:"required,min=1,max=4"`
}

type CreateTripDraftResponse struct {
	ID            uuid.UUID  `json:"id"`
	DriverID      uuid.UUID  `json:"driverId"`
	FromPoint     string     `json:"fromPoint"`
	ToPoint       string     `json:"toPoint"`
	CreatedAt     time.Time  `json:"createdAt"`
	DepartureTime time.Time  `json:"departureTime"`
	Seats         int        `json:"seats"`
	Status        StatusEnum `json:"status"`
}

type GetTripByIDResponse struct {
	ID            uuid.UUID  `json:"id"`
	DriverID      uuid.UUID  `json:"driverId"`
	FromPoint     string     `json:"fromPoint"`
	ToPoint       string     `json:"toPoint"`
	Seats         int        `json:"seats"`
	CreatedAt     time.Time  `json:"createdAt"`
	DepartureTime time.Time  `json:"departureTime"`
	Status        StatusEnum `json:"status"`
}

type MoveTripDraftToPublishResponse struct {
	ID            uuid.UUID  `json:"id"`
	DriverID      uuid.UUID  `json:"driverId"`
	FromPoint     string     `json:"fromPoint"`
	ToPoint       string     `json:"toPoint"`
	Seats         int        `json:"seats"`
	CreatedAt     time.Time  `json:"createdAt"`
	DepartureTime time.Time  `json:"departureTime"`
	Status        StatusEnum `json:"status"`
}

type MoveTripPublishedToStartedResponse struct {
	ID            uuid.UUID  `json:"id"`
	DriverID      uuid.UUID  `json:"driverId"`
	FromPoint     string     `json:"fromPoint"`
	ToPoint       string     `json:"toPoint"`
	Seats         int        `json:"seats"`
	CreatedAt     time.Time  `json:"createdAt"`
	DepartureTime time.Time  `json:"departureTime"`
	Status        StatusEnum `json:"status"`
	Allowed       bool       `json:"allowed"`
	Reason        string     `json:"reason"`
}

// PageResponse model info
// @Description Employee account information
// @Description with employee result, page_size, page_number, total
type PageResponse struct {
	Result     []CreateTripDraftResponse `json:"result"`
	PageSize   int64                     `json:"page_size" `
	PageNumber int64                     `json:"page_number"`
	Total      int64                     `json:"total"`
}
