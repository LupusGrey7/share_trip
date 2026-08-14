package model

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

const (
	StatusDraft     StatusEnum = "draft"
	StatusPublished StatusEnum = "published"
	StatusCancelled StatusEnum = "cancelled"
	StatusCompleted StatusEnum = "completed"
)

type Entity struct {
	ID            uuid.UUID  `db:"id"`
	DriverID      uuid.UUID  `db:"driver_id"`
	FromPoint     string     `db:"from_point"`
	ToPoint       string     `db:"to_point"`
	CreatedAt     time.Time  `db:"created_at"`
	DepartureTime time.Time  `db:"departure_time"`
	Seats         int        `db:"seats"`
	Status        StatusEnum `db:"status"`
}

// CreateTripDraftResponse model info
// @Description Trip information
// @Description with trip id, driverId, fromPoint, toPoint, createAt, departureTime, seats, status
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

type GetByIDModelRequest struct {
	ID string
}

type GetTripByIDModelResponse struct {
	ID            uuid.UUID  `json:"id"`
	DriverID      uuid.UUID  `json:"driverId"`
	FromPoint     string     `json:"fromPoint"`
	ToPoint       string     `json:"toPoint"`
	Seats         int        `json:"seats"`
	CreatedAt     time.Time  `json:"createdAt" validate:"required"`
	DepartureTime time.Time  `json:"departureTime" validate:"required"`
	Status        StatusEnum `json:"status"`
}

type CreateTripRequestModel struct {
	DriverID       uuid.UUID
	FromPoint      string
	ToPoint        string
	DepartureTime  time.Time
	AvailableSeats int
}

type MoveTripDraftToPublishRequestModel struct {
	ID       string
	ClientID uuid.UUID `json:"clientId" validate:"required,uuid"` //"omitempty,uuid"
}

type MoveTripDraftToPublishModel struct {
	ID       string
	ClientID uuid.UUID `json:"clientId" validate:"required,uuid"` //"omitempty,uuid"
}

type MoveTripDraftToPublishModelResponse struct {
	ID            uuid.UUID  `json:"id"`
	DriverID      uuid.UUID  `json:"driverId"`
	FromPoint     string     `json:"fromPoint"`
	ToPoint       string     `json:"toPoint"`
	Seats         int        `json:"seats"`
	CreatedAt     time.Time  `json:"createdAt" validate:"required"`
	DepartureTime time.Time  `json:"departureTime" validate:"required"`
	Status        StatusEnum `json:"status"`
}

type MoveTripPublishedToStartModel struct {
	ID          string
	ClientID    uuid.UUID
	CompanyID   string
	ServiceCode ServiceCodeEnum
}

type MoveTripPublishedToStartRequestModel struct {
	ID          string
	ClientID    uuid.UUID       `json:"clientId" validate:"required,uuid"`
	CompanyID   string          `json:"companyId" validate:"required"`
	ServiceCode ServiceCodeEnum `json:"serviceCode" validate:"required" enum:"trip_publication"`
}

type MoveTripPublishedToStartModelResponse struct {
	ID            uuid.UUID  `json:"id"`
	DriverID      uuid.UUID  `json:"driverId"`
	FromPoint     string     `json:"fromPoint"`
	ToPoint       string     `json:"toPoint"`
	Seats         int        `json:"seats"`
	CreatedAt     time.Time  `json:"createdAt" validate:"required"`
	DepartureTime time.Time  `json:"departureTime" validate:"required"`
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

func (req *CreateTripRequestModel) ToEntity() *Entity {
	return &Entity{
		DriverID:      req.DriverID,
		FromPoint:     req.FromPoint,
		ToPoint:       req.ToPoint,
		DepartureTime: req.DepartureTime,
		Seats:         req.AvailableSeats,
	}
}

func (req *Entity) ToResponse() CreateTripDraftResponse {
	return CreateTripDraftResponse{
		ID:            req.ID,
		DriverID:      req.DriverID,
		FromPoint:     req.FromPoint,
		ToPoint:       req.ToPoint,
		CreatedAt:     req.CreatedAt,
		DepartureTime: req.DepartureTime,
		Seats:         req.Seats,
		Status:        req.Status,
	}
}

func (req *Entity) ToCreateResponse() *CreateTripDraftResponse {
	return &CreateTripDraftResponse{
		ID:            req.ID,
		DriverID:      req.DriverID,
		FromPoint:     req.FromPoint,
		ToPoint:       req.ToPoint,
		CreatedAt:     req.CreatedAt,
		DepartureTime: req.DepartureTime,
		Seats:         req.Seats,
		Status:        req.Status,
	}
}

func (req *CreateTripDraftResponse) ToResponse(entity Entity) *CreateTripDraftResponse {
	return &CreateTripDraftResponse{
		ID:            entity.ID,
		DriverID:      entity.DriverID,
		FromPoint:     entity.FromPoint,
		ToPoint:       entity.ToPoint,
		CreatedAt:     entity.CreatedAt,
		DepartureTime: entity.DepartureTime,
		Seats:         entity.Seats,
		Status:        entity.Status,
	}
}

func (req *MoveTripDraftToPublishModelResponse) ToPublishModelResponse(entity Entity) *MoveTripDraftToPublishModelResponse {
	return &MoveTripDraftToPublishModelResponse{
		ID:            entity.ID,
		DriverID:      entity.DriverID,
		FromPoint:     entity.FromPoint,
		ToPoint:       entity.ToPoint,
		CreatedAt:     entity.CreatedAt,
		DepartureTime: entity.DepartureTime,
		Seats:         entity.Seats,
		Status:        entity.Status,
	}
}

func (req *Entity) ToUpdatedPublishModelResponse() *MoveTripDraftToPublishModelResponse {
	return &MoveTripDraftToPublishModelResponse{
		ID:            req.ID,
		DriverID:      req.DriverID,
		FromPoint:     req.FromPoint,
		ToPoint:       req.ToPoint,
		CreatedAt:     req.CreatedAt,
		DepartureTime: req.DepartureTime,
		Seats:         req.Seats,
		Status:        req.Status,
	}
}

func (e *Entity) ToPublishedStartModelResponse(allowed bool, reason string) *MoveTripPublishedToStartModelResponse {
	return &MoveTripPublishedToStartModelResponse{
		ID:            e.ID,
		DriverID:      e.DriverID,
		FromPoint:     e.FromPoint,
		ToPoint:       e.ToPoint,
		CreatedAt:     e.CreatedAt,
		DepartureTime: e.DepartureTime,
		Seats:         e.Seats,
		Status:        e.Status,
		Allowed:       allowed,
		Reason:        reason,
	}
}

func (req *Entity) ToGetByIdModelResponse() *GetTripByIDModelResponse {
	return &GetTripByIDModelResponse{
		ID:            req.ID,
		DriverID:      req.DriverID,
		FromPoint:     req.FromPoint,
		ToPoint:       req.ToPoint,
		CreatedAt:     req.CreatedAt,
		DepartureTime: req.DepartureTime,
		Seats:         req.Seats,
		Status:        req.Status,
	}
}
