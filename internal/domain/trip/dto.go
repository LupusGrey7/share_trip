package trip

import (
	"time"

	"github.com/google/uuid"
)

type StatusEnum string

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

// CreateTripResponse model info
// @Description Trip information
// @Description with trip id, driverId, fromPoint, toPoint, createAt, departureTime, seats, status
type CreateTripResponse struct {
	ID            uuid.UUID  `json:"id"`
	DriverID      uuid.UUID  `json:"driverId"`
	FromPoint     string     `json:"fromPoint"`
	ToPoint       string     `json:"toPoint"`
	CreatedAt     time.Time  `json:"createdAt"`
	DepartureTime time.Time  `json:"departureTime"`
	Seats         int        `json:"seats"`
	Status        StatusEnum `json:"status"`
}

type GetByIdModelRequest struct {
	ID uuid.UUID `json:"id" validate:"required,uuid"`
}

type GetTripByIdModelResponse struct {
	ID            uuid.UUID  `json:"id"`
	DriverID      uuid.UUID  `json:"driverId"`
	FromPoint     string     `json:"fromPoint"`
	ToPoint       string     `json:"toPoint"`
	Seats         int        `json:"seats"`
	CreatedAt     time.Time  `json:"createdAt" validate:"required"`
	DepartureTime time.Time  `json:"departureTime" validate:"required"`
	Status        StatusEnum `json:"status"`
}

type CreateTripRequest struct {
	DriverID       uuid.UUID `json:"driverId" validate:"required,uuid"`
	FromPoint      string    `json:"fromPoint" validate:"required,min=20,max=155"`
	ToPoint        string    `json:"toPoint" validate:"required,min=20,max=155"`
	DepartureTime  time.Time `json:"departureTime" validate:"required"`
	AvailableSeats int       `json:"seats" validate:"required,min=1,max=3"`
}

type MoveTripDraftToPublishModelRequest struct {
	ID       string
	ClientID uuid.UUID `json:"clientId" validate:"required,uuid"` //"omitempty,uuid"
}

type MoveTripDraftToPublishModel struct {
	ID       uuid.UUID
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

// PageResponse model info
// @Description Employee account information
// @Description with employee result, page_size, page_number, total
type PageResponse struct {
	Result     []CreateTripResponse `json:"result"`
	PageSize   int64                `json:"page_size" `
	PageNumber int64                `json:"page_number"`
	Total      int64                `json:"total"`
}

func (req *CreateTripRequest) ToEntity() *Entity {
	return &Entity{
		DriverID:      req.DriverID,
		FromPoint:     req.FromPoint,
		ToPoint:       req.ToPoint,
		DepartureTime: req.DepartureTime,
		Seats:         req.AvailableSeats,
	}
}

func (req *Entity) ToResponse() CreateTripResponse {
	return CreateTripResponse{
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

func (req *Entity) ToCreateResponse() *CreateTripResponse {
	return &CreateTripResponse{
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

func (req *CreateTripResponse) ToResponse(entity Entity) *CreateTripResponse {
	return &CreateTripResponse{
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

func (req *Entity) UpdateToPublishModelResponse() *MoveTripDraftToPublishModelResponse {
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

func (req *Entity) ToGetByIdModelResponse() *GetTripByIdModelResponse {
	return &GetTripByIdModelResponse{
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
