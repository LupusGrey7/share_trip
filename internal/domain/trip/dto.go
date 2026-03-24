package trip

import (
	"github.com/google/uuid"
	"time"
)

type StatusEnum string

const (
	Draft     StatusEnum = "draft"
	Published StatusEnum = "published"
	Cancelled StatusEnum = "cancelled"
	Completed StatusEnum = "completed"
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

// Response model info
// @Description Trip information
// @Description with trip id, driverId, fromPoint, toPoint, createAt, departureTime, seats, status
type Response struct {
	ID            uuid.UUID  `json:"id"`
	DriverID      uuid.UUID  `json:"driverId"`
	FromPoint     string     `json:"fromPoint"`
	ToPoint       string     `json:"toPoint"`
	CreatedAt     time.Time  `json:"createdAt"`
	DepartureTime time.Time  `json:"departureTime"`
	Seats         int        `json:"seats"`
	Status        StatusEnum `json:"status"`
}

type GetByIdRequest struct {
	ID uuid.UUID `json:"id" validate:"required,uuid"`
}

type CreateTripCommand struct {
	DriverID       uuid.UUID `json:"driverId" validate:"required,uuid"`
	FromPoint      string    `json:"fromPoint" validate:"required,min=20,max=155"`
	ToPoint        string    `json:"toPoint" validate:"required,min=20,max=155"`
	DepartureTime  time.Time `json:"departureTime" validate:"required"`
	AvailableSeats int       `json:"seats" validate:"required,min=1,max=3"`
}

type UpdateTripCommand struct {
	DriverID      uuid.UUID  `json:"driverId" validate:"required,uuid"` //"omitempty,uuid"
	FromPoint     string     `json:"fromPoint" validate:"required,min=20,max=155"`
	ToPoint       string     `json:"toPoint" validate:"required,min=20,max=155"`
	CreatedAt     time.Time  `json:"createdAt" validate:"required"`
	DepartureTime time.Time  `json:"departureTime" validate:"required"`
	Seats         int        `json:"seats" validate:"required,min=1,max=3"`
	Status        StatusEnum `json:"status" validate:"required,oneof=draft published canceled completed"`
}

// PageResponse model info
// @Description Employee account information
// @Description with employee result, page_size, page_number, total
type PageResponse struct {
	Result     []Response `json:"result"`
	PageSize   int64      `json:"page_size" `
	PageNumber int64      `json:"page_number"`
	Total      int64      `json:"total"`
}

func (req *CreateTripCommand) ToEntity() *Entity {
	return &Entity{
		DriverID:      req.DriverID,
		FromPoint:     req.FromPoint,
		ToPoint:       req.ToPoint,
		DepartureTime: req.DepartureTime,
		Seats:         req.AvailableSeats,
	}
}

func (req *Entity) ToResponse() Response {
	return Response{
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
