package outbox

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"job4j.ru/share_trip/internal/clients/kafka"
	"job4j.ru/share_trip/internal/outbox/domain"
)

type fakeProducer struct {
	err   error
	calls []kafka.TripPublished
}

func (f *fakeProducer) PublishTripPublished(_ context.Context, event kafka.TripPublished) error {
	f.calls = append(f.calls, event)
	return f.err
}

func TestDeliverBatch_SuccessMarksSent(t *testing.T) {
	t.Parallel()

	id1 := uuid.New()
	id2 := uuid.New()
	events := []*domain.Entity{pendingEvent(id1), pendingEvent(id2)}
	prod := &fakeProducer{}
	var sent []uuid.UUID

	err := deliverBatch(
		context.Background(),
		events,
		prod,
		func(_ context.Context, id uuid.UUID) error {
			sent = append(sent, id)
			return nil
		},
		func(context.Context, uuid.UUID, error) error {
			t.Fatal("MarkFailed must not be called")
			return nil
		},
	)
	require.NoError(t, err)
	require.Len(t, prod.calls, 2)
	require.Equal(t, []uuid.UUID{id1, id2}, sent)
	require.Equal(t, id1.String(), prod.calls[0].EventID)
}

func TestDeliverBatch_ProduceErrorMarksFailedAndContinues(t *testing.T) {
	t.Parallel()

	id1 := uuid.New()
	id2 := uuid.New()
	events := []*domain.Entity{pendingEvent(id1), pendingEvent(id2)}
	prod := &fakeProducer{err: errors.New("broker down")}
	var failed []uuid.UUID

	err := deliverBatch(
		context.Background(),
		events,
		prod,
		func(context.Context, uuid.UUID) error {
			t.Fatal("MarkSent must not be called when produce fails")
			return nil
		},
		func(_ context.Context, id uuid.UUID, produceErr error) error {
			failed = append(failed, id)
			require.ErrorContains(t, produceErr, "broker down")
			return nil
		},
	)
	require.NoError(t, err)
	require.Equal(t, []uuid.UUID{id1, id2}, failed)
}

func TestDeliverBatch_MarkSentErrorStopsBatch(t *testing.T) {
	t.Parallel()

	id := uuid.New()
	prod := &fakeProducer{}
	markErr := errors.New("db mark sent failed")

	err := deliverBatch(
		context.Background(),
		[]*domain.Entity{pendingEvent(id)},
		prod,
		func(context.Context, uuid.UUID) error { return markErr },
		func(context.Context, uuid.UUID, error) error {
			t.Fatal("MarkFailed must not be called")
			return nil
		},
	)
	require.ErrorIs(t, err, markErr)
}

func TestDeliverBatch_MarkFailedErrorStopsBatch(t *testing.T) {
	t.Parallel()

	id := uuid.New()
	prod := &fakeProducer{err: errors.New("kafka")}
	markErr := errors.New("db mark failed")

	err := deliverBatch(
		context.Background(),
		[]*domain.Entity{pendingEvent(id)},
		prod,
		func(context.Context, uuid.UUID) error {
			t.Fatal("MarkSent must not be called")
			return nil
		},
		func(context.Context, uuid.UUID, error) error { return markErr },
	)
	require.ErrorIs(t, err, markErr)
}

func pendingEvent(id uuid.UUID) *domain.Entity {
	e := domain.NewTripPublishedEvent(
		id,
		uuid.MustParse("22222222-2222-2222-2222-222222222222"),
		uuid.MustParse("33333333-3333-3333-3333-333333333333"),
		"acme01",
		time.Date(2026, 9, 17, 12, 0, 0, 0, time.UTC),
	)
	return &e
}
