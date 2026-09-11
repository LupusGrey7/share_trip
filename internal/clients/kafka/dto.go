package kafka

import (
	"time"
)

type EventType string

const (
	EventTypePublished EventType = "trip_published"
	EventTypeCancelled EventType = "trip_cancelled"
	EventTypeCompleted EventType = "trip_completed"
	EventTypeStarted   EventType = "trip_started"
)

// TripPublishedPayload — business fields inside envelope.payload.
type TripPublishedPayload struct {
	TripID    string `json:"trip_id"`
	DriverID  string `json:"driver_id"`
	CompanyID string `json:"company_id"`
}

// TripPublished — Kafka envelope (envelope + nested payload), not a flat JSON.
type TripPublished struct {
	EventID    string               `json:"event_id" validate:"required,uuid"`
	EventType  EventType            `json:"event_type" validate:"required,oneof=trip_published trip_cancellation trip_completion trip_start"`
	OccurredAt time.Time            `json:"occurred_at"`
	Payload    TripPublishedPayload `json:"payload"`
}
