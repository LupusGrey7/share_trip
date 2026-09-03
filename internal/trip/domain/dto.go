package domain

import (
	"time"

	"github.com/google/uuid"
	"job4j.ru/share_trip/internal/client/contracts"
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

// Status codes match trip_status.name in DB dictionary (not a PostgreSQL ENUM).
const (
	StatusDraft     StatusEnum = "draft"
	StatusPublished StatusEnum = "published"
	StatusStarted   StatusEnum = "started"
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
	Status        StatusEnum `db:"status"` // trip_status.name via JOIN / RETURNING subquery
}

type GetByIDInput struct {
	ID string
}

// GetTripByIDOutput — usecase/service boundary (no json tags).
type GetTripByIDOutput struct {
	ID            uuid.UUID
	DriverID      uuid.UUID
	FromPoint     string
	ToPoint       string
	Seats         int
	CreatedAt     time.Time
	DepartureTime time.Time
	Status        StatusEnum
}

type CreateTripInput struct {
	DriverID       uuid.UUID
	FromPoint      string
	ToPoint        string
	DepartureTime  time.Time
	AvailableSeats int
}

type CreateTripOutput struct {
	ID            uuid.UUID
	DriverID      uuid.UUID
	FromPoint     string
	ToPoint       string
	CreatedAt     time.Time
	DepartureTime time.Time
	Seats         int
	Status        StatusEnum
}

type MoveTripPublishedToStartedInput struct {
	ID            string
	ClientID      uuid.UUID
	CompanyID     string
	ServiceCode   ServiceCodeEnum
	ContractCheck *contracts.CheckResult // if result is nil, then contract check is not performed
}

type MoveTripDraftToPublishInput struct {
	ID       string
	ClientID uuid.UUID
}

type MoveTripDraftToPublishOutput struct {
	ID            uuid.UUID
	DriverID      uuid.UUID
	FromPoint     string
	ToPoint       string
	Seats         int
	CreatedAt     time.Time
	DepartureTime time.Time
	Status        StatusEnum
}

type MoveTripPublishedToStartedOutput struct {
	ID            uuid.UUID
	DriverID      uuid.UUID
	FromPoint     string
	ToPoint       string
	Seats         int
	CreatedAt     time.Time
	DepartureTime time.Time
	Status        StatusEnum
	Allowed       bool
	Reason        string
}

// PageResponse model info
// @Description Employee account information
// @Description with employee result, page_size, page_number, total
type PageResponse struct {
	Result     []CreateTripOutput `json:"result"`
	PageSize   int64              `json:"page_size" `
	PageNumber int64              `json:"page_number"`
	Total      int64              `json:"total"`
}

func (req *CreateTripInput) ToEntity() *Entity {
	return &Entity{
		DriverID:      req.DriverID,
		FromPoint:     req.FromPoint,
		ToPoint:       req.ToPoint,
		DepartureTime: req.DepartureTime,
		Seats:         req.AvailableSeats,
	}
}

func entityToCreateOutput(entity *Entity) *CreateTripOutput {
	if entity == nil {
		return nil
	}
	return &CreateTripOutput{
		ID:            entity.ID,
		DriverID:      entity.DriverID,
		FromPoint:     entity.FromPoint,
		ToPoint:       entity.ToPoint,
		Seats:         entity.Seats,
		CreatedAt:     entity.CreatedAt,
		DepartureTime: entity.DepartureTime,
		Status:        entity.Status,
	}
}

func (e *Entity) ToCreateTripOutput() *CreateTripOutput {
	return entityToCreateOutput(e)
}

func entityToPublishOutput(entity *Entity) *MoveTripDraftToPublishOutput {
	if entity == nil {
		return nil
	}
	return &MoveTripDraftToPublishOutput{
		ID:            entity.ID,
		DriverID:      entity.DriverID,
		FromPoint:     entity.FromPoint,
		ToPoint:       entity.ToPoint,
		Seats:         entity.Seats,
		CreatedAt:     entity.CreatedAt,
		DepartureTime: entity.DepartureTime,
		Status:        entity.Status,
	}
}

func (e *Entity) ToMoveTripDraftToPublishOutput() *MoveTripDraftToPublishOutput {
	return entityToPublishOutput(e)
}

func (e *Entity) ToMoveTripPublishedToStartedOutput(allowed bool, reason string) *MoveTripPublishedToStartedOutput {
	out := entityToPublishOutput(e)
	if out == nil {
		return nil
	}
	return &MoveTripPublishedToStartedOutput{
		ID:            out.ID,
		DriverID:      out.DriverID,
		FromPoint:     out.FromPoint,
		ToPoint:       out.ToPoint,
		Seats:         out.Seats,
		CreatedAt:     out.CreatedAt,
		DepartureTime: out.DepartureTime,
		Status:        out.Status,
		Allowed:       allowed,
		Reason:        reason,
	}
}

// TripEntityToOutput closes the entity boundary inside usecase (Contract-style).
func TripEntityToOutput(entity *Entity) *GetTripByIDOutput {
	if entity == nil {
		return nil
	}
	return &GetTripByIDOutput{
		ID:            entity.ID,
		DriverID:      entity.DriverID,
		FromPoint:     entity.FromPoint,
		ToPoint:       entity.ToPoint,
		Seats:         entity.Seats,
		CreatedAt:     entity.CreatedAt,
		DepartureTime: entity.DepartureTime,
		Status:        entity.Status,
	}
}
