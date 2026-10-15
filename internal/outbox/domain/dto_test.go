package domain_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"job4j.ru/share_trip/internal/clients/kafka"
	"job4j.ru/share_trip/internal/outbox/domain"
)

func TestNewTripPublishedEvent_AndToTripPublished(t *testing.T) {
	t.Parallel()

	eventID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	tripID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	driverID := uuid.MustParse("33333333-3333-3333-3333-333333333333")
	at := time.Date(2026, 9, 17, 12, 0, 0, 0, time.UTC)

	e := domain.NewTripPublishedEvent(eventID, tripID, driverID, "acme01", at)

	require.Equal(t, eventID, e.ID)
	require.Equal(t, domain.AggregateTypeTrip, e.AggregateType)
	require.Equal(t, tripID, e.AggregateID)
	require.Equal(t, string(domain.EventPublished), e.EventType)
	require.Equal(t, domain.StatusPending, e.Status)
	require.Equal(t, 0, e.Attempts)
	require.Equal(t, tripID.String(), e.Payload.TripID)
	require.Equal(t, driverID.String(), e.Payload.DriverID)
	require.Equal(t, "acme01", e.Payload.CompanyID)
	require.Equal(t, at, e.CreatedAt)

	got := e.ToTripPublished()
	require.Equal(t, kafka.TripPublished{
		EventID:    eventID.String(),
		EventType:  kafka.EventTypePublished,
		OccurredAt: at,
		Payload: kafka.PayloadEvent{
			TripID:    tripID.String(),
			DriverID:  driverID.String(),
			CompanyID: "acme01",
		},
	}, got)
}
