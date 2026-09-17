package domain

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"job4j.ru/share_trip/internal/clients/kafka"
)

// Status of an outbox_events row.
type Status string

const (
	StatusPending Status = "pending"
	StatusSent    Status = "sent"
	StatusFailed  Status = "failed"
)

// EventType matches outbox_events.event_type / Kafka event_type.
type EventType string

const (
	EventPublished EventType = "trip_published"
	EventCancelled EventType = "trip_cancelled"
	EventCompleted EventType = "trip_completed"
)

const AggregateTypeTrip = "trip"

// PayloadEvent is stored in outbox_events.payload (JSONB) and maps to Kafka envelope.payload.
type PayloadEvent struct {
	TripID    string `json:"trip_id"`
	DriverID  string `json:"driver_id"`
	CompanyID string `json:"company_id"`
}

// Entity mirrors table outbox_events (1:1).
type Entity struct {
	ID            uuid.UUID    `db:"id"`
	AggregateType string       `db:"aggregate_type"`
	AggregateID   uuid.UUID    `db:"aggregate_id"`
	EventType     string       `db:"event_type"`
	Payload       PayloadEvent `db:"payload"`
	Status        Status       `db:"status"`
	Attempts      int          `db:"attempts"`
	LastError     *string      `db:"last_error"`
	CreatedAt     time.Time    `db:"created_at"`
	SentAt        *time.Time   `db:"sent_at"`
}

// NewTripPublishedEvent builds a pending outbox row for the publish TX.
// Call only inside the same transaction that updates the trip.
func NewTripPublishedEvent(
	eventID uuid.UUID,
	tripID uuid.UUID,
	driverID uuid.UUID,
	companyID string,
	at time.Time,
) Entity {
	if at.IsZero() {
		at = time.Now()
	}
	return Entity{
		ID:            eventID,
		AggregateType: AggregateTypeTrip,
		AggregateID:   tripID,
		EventType:     string(EventPublished),
		Payload: PayloadEvent{
			TripID:    tripID.String(),
			DriverID:  driverID.String(),
			CompanyID: companyID,
		},
		Status:    StatusPending,
		Attempts:  0,
		CreatedAt: at,
	}
}

// ToTripPublished builds the Kafka envelope from an outbox row.
func (e *Entity) ToTripPublished() kafka.TripPublished {
	return kafka.TripPublished{
		EventID:    e.ID.String(),
		EventType:  kafka.EventType(e.EventType),
		OccurredAt: e.CreatedAt,
		Payload: kafka.PayloadEvent{
			TripID:    e.Payload.TripID,
			DriverID:  e.Payload.DriverID,
			CompanyID: e.Payload.CompanyID,
		},
	}
}

// Value — JSONB write (INSERT / UPDATE).
func (p PayloadEvent) Value() (driver.Value, error) {
	return json.Marshal(p)
}

// Scan — JSONB read (SELECT).
func (p *PayloadEvent) Scan(src any) error {
	if src == nil {
		*p = PayloadEvent{}
		return nil
	}

	var data []byte
	switch v := src.(type) {
	case []byte:
		data = v
	case string:
		data = []byte(v)
	default:
		return fmt.Errorf("unsupported type for PayloadEvent: %T", src)
	}

	return json.Unmarshal(data, p)
}
