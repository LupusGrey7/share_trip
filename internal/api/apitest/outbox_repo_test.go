package apitest

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
	"job4j.ru/share_trip/internal/observability/metrics"
	"job4j.ru/share_trip/internal/outbox/domain"
	"job4j.ru/share_trip/internal/storage"
)

func TestOutboxRepo_CreateLockMark(t *testing.T) {
	t.Parallel()
	lockIT(t)

	reg := prometheus.NewRegistry()
	m := metrics.New(reg)
	repo := storage.NewOutboxEventRepository(m)

	eventID := uuid.New()
	tripID := uuid.New()
	driverID := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	at := time.Now().UTC().Truncate(time.Millisecond)

	event := domain.NewTripPublishedEvent(eventID, tripID, driverID, "acme01", at)

	ctx := context.Background()
	tx, err := testPool.Begin(ctx)
	require.NoError(t, err)
	defer func() { _ = tx.Rollback(ctx) }()

	require.NoError(t, repo.CreateTx(ctx, tx, &event))
	require.NoError(t, tx.Commit(ctx))

	t.Cleanup(func() {
		_, _ = testPool.Exec(context.Background(), `delete from outbox_events where id = $1`, eventID)
	})

	txLock, err := testPool.Begin(ctx)
	require.NoError(t, err)
	defer func() { _ = txLock.Rollback(ctx) }()

	pending, err := repo.LockPendingTx(ctx, txLock, 50)
	require.NoError(t, err)

	var found *domain.Entity
	for _, e := range pending {
		if e.ID == eventID {
			found = e
			break
		}
	}
	require.NotNil(t, found, "created event must appear in LockPendingTx")
	require.Equal(t, domain.StatusPending, found.Status)
	require.Equal(t, "acme01", found.Payload.CompanyID)
	require.Equal(t, tripID.String(), found.Payload.TripID)

	require.NoError(t, repo.MarkFailedTx(ctx, txLock, eventID, errors.New("kafka timeout")))
	require.NoError(t, txLock.Commit(ctx))

	var (
		status    string
		attempts  int
		lastError *string
	)
	err = testPool.QueryRow(ctx, `
		select status, attempts, last_error from outbox_events where id = $1
	`, eventID).Scan(&status, &attempts, &lastError)
	require.NoError(t, err)
	require.Equal(t, "pending", status, "below max attempts stays pending for retry")
	require.Equal(t, 1, attempts)
	require.NotNil(t, lastError)
	require.Contains(t, *lastError, "kafka timeout")

	txSent, err := testPool.Begin(ctx)
	require.NoError(t, err)
	defer func() { _ = txSent.Rollback(ctx) }()

	// status still pending — can lock again after failed attempt
	require.NoError(t, repo.MarkSentTx(ctx, txSent, eventID))
	require.NoError(t, txSent.Commit(ctx))

	var sentAt *time.Time
	err = testPool.QueryRow(ctx, `
		select status, attempts, sent_at, last_error from outbox_events where id = $1
	`, eventID).Scan(&status, &attempts, &sentAt, &lastError)
	require.NoError(t, err)
	require.Equal(t, "sent", status)
	require.NotNil(t, sentAt)
	require.Nil(t, lastError)
}
