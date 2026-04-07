package outbox

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

//outbox_event
// В payload достаточно сохранить JSON с trip_id.

type EventNameEnum string

const (
	EventDraft     EventNameEnum = "trip_draft"
	EventPublished EventNameEnum = "trip_published"
	EventCancelled EventNameEnum = "trip_cancelled"
	EventCompleted EventNameEnum = "trip_completed"
)

// PayloadEvent - ваша payload-структура
type PayloadEvent struct {
	TripID uuid.UUID
}

type Entity struct {
	Id          int64        `db:"id"`
	EventName   string       `db:"event_name"`
	AggregateId uuid.UUID    `db:"aggregate_id"` //trip_id
	Payload     PayloadEvent `db:"payload"`      //	Payload     json.RawMessage `db:"payload"`
	CreatedAt   time.Time    `db:"created_at"`
}

type CreateOutboxRequest struct{}

func (req *Entity) ToEntity() *Entity {
	return &Entity{
		EventName:   req.EventName,
		AggregateId: req.AggregateId,
		Payload:     req.Payload,
		CreatedAt:   req.CreatedAt,
	}
}

// Value — для записи в БД (JSONB), для INSERT / UPDATE (запись в БД)
func (p *PayloadEvent) Value() (driver.Value, error) {
	return json.Marshal(p)
}

// Scan — для чтения из БД (JSONB) - для SELECT
func (p *PayloadEvent) Scan(src any) error {
	if src == nil {
		*p = PayloadEvent{} // или оставьте нулевым значением
		return nil
	}

	var data []byte
	switch v := src.(type) {
	case []byte:
		data = v
	case string:
		data = []byte(v)
	default:
		return errors.New("unsupported type for PayloadEvent")
	}

	return json.Unmarshal(data, p)
}
