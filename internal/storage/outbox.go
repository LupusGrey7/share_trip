// outbox_events repository (create + poller lock/mark).

package storage

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"go.opentelemetry.io/otel"
	"job4j.ru/share_trip/internal/observability/logctx"
	"job4j.ru/share_trip/internal/observability/metrics"
	"job4j.ru/share_trip/internal/outbox/domain"
)

const (
	// OutboxMaxAttempts — after this many failed Produce attempts status becomes failed.
	OutboxMaxAttempts = 10

	createEvent = `
insert into outbox_events(id, aggregate_type, aggregate_id, event_type, payload, status, attempts, created_at)
values($1, $2, $3, $4, $5, $6, $7, $8)
`

	lockPending = `
select
	id,
	aggregate_type,
	aggregate_id,
	event_type,
	payload,
	status,
	attempts,
	last_error,
	created_at,
	sent_at
from outbox_events
where status = 'pending'
order by created_at
for update skip locked
limit $1
`

	markSent = `
update outbox_events
set status = 'sent',
    sent_at = now(),
    last_error = null
where id = $1
`

	markFailed = `
update outbox_events
set attempts = attempts + 1,
    last_error = $2,
    status = case
        when attempts + 1 >= $3 then 'failed'
        else status
    end
where id = $1
`
)

type OutboxRepository interface {
	CreateTx(ctx context.Context, tx pgx.Tx, o *domain.Entity) error
	LockPendingTx(ctx context.Context, tx pgx.Tx, limit int) ([]*domain.Entity, error)
	MarkSentTx(ctx context.Context, tx pgx.Tx, id uuid.UUID) error
	MarkFailedTx(ctx context.Context, tx pgx.Tx, id uuid.UUID, produceErr error) error
}

type OutboxEventRepository struct {
	metrics *metrics.Metrics
}

func NewOutboxEventRepository(m *metrics.Metrics) *OutboxEventRepository {
	return &OutboxEventRepository{
		metrics: m,
	}
}

func (r *OutboxEventRepository) CreateTx(
	ctx context.Context,
	tx pgx.Tx,
	o *domain.Entity,
) error {
	tracer := otel.Tracer("OutboxEventRepository")
	ctxSpc, span := tracer.Start(ctx, "OutboxEventRepository.CreateTx")

	started := time.Now()
	op := "outbox_create"
	result := MetricsResultSuccess

	defer func() {
		r.metrics.RepositoryQueryTotal.WithLabelValues(op, result).Inc()
		r.metrics.RepositoryQueryDuration.WithLabelValues(op, result).
			Observe(time.Since(started).Seconds())
		span.End()
	}()

	logger := logctx.Logger(ctxSpc).With(
		slog.String("layer", "repository"),
		slog.String("repository", "OutboxEventRepository"),
		slog.String("operation", "CreateTx"),
		slog.String("aggregate_id", o.AggregateID.String()),
		slog.String("event_id", o.ID.String()),
	)
	logger.Debug("outbox create started")

	createdAt := o.CreatedAt
	if createdAt.IsZero() {
		createdAt = time.Now()
	}

	status := o.Status
	if status == "" {
		status = domain.StatusPending
	}

	args := []interface{}{
		o.ID,
		o.AggregateType,
		o.AggregateID,
		o.EventType,
		o.Payload,
		status,
		o.Attempts,
		createdAt,
	}

	if _, err := tx.Exec(ctxSpc, createEvent, args...); err != nil {
		result = "error"
		return fmt.Errorf("err when create outbox event trip publish: %w", err)
	}

	logger.Debug("outbox create completed")
	return nil
}

func (r *OutboxEventRepository) LockPendingTx(
	ctx context.Context,
	tx pgx.Tx,
	limit int,
) ([]*domain.Entity, error) {
	tracer := otel.Tracer("OutboxEventRepository")
	ctxSpc, span := tracer.Start(ctx, "OutboxEventRepository.LockPendingTx")

	started := time.Now()
	op := "outbox_lock_pending"
	result := MetricsResultSuccess

	defer func() {
		r.metrics.RepositoryQueryTotal.WithLabelValues(op, result).Inc()
		r.metrics.RepositoryQueryDuration.WithLabelValues(op, result).
			Observe(time.Since(started).Seconds())
		span.End()
	}()

	if limit <= 0 {
		return nil, nil
	}

	rows, err := tx.Query(ctxSpc, lockPending, limit)
	if err != nil {
		result = "error"
		return nil, fmt.Errorf("lock pending outbox events: %w", err)
	}
	defer rows.Close()

	events := make([]*domain.Entity, 0, limit)
	for rows.Next() {
		var (
			e          domain.Entity
			status     string
			lastError  *string
			sentAt     *time.Time
		)
		if err := rows.Scan(
			&e.ID,
			&e.AggregateType,
			&e.AggregateID,
			&e.EventType,
			&e.Payload,
			&status,
			&e.Attempts,
			&lastError,
			&e.CreatedAt,
			&sentAt,
		); err != nil {
			result = "error"
			return nil, fmt.Errorf("scan pending outbox event: %w", err)
		}
		e.Status = domain.Status(status)
		e.LastError = lastError
		e.SentAt = sentAt
		events = append(events, &e)
	}
	if err := rows.Err(); err != nil {
		result = "error"
		return nil, fmt.Errorf("iterate pending outbox events: %w", err)
	}

	slog.Info("lock pending outbox events completed", "events", len(events))
	return events, nil
}

func (r *OutboxEventRepository) MarkSentTx(
	ctx context.Context,
	tx pgx.Tx,
	id uuid.UUID,
) error {
	tracer := otel.Tracer("OutboxEventRepository")
	ctxSpc, span := tracer.Start(ctx, "OutboxEventRepository.MarkSentTx")

	started := time.Now()
	op := "outbox_mark_sent"
	result := MetricsResultSuccess

	defer func() {
		r.metrics.RepositoryQueryTotal.WithLabelValues(op, result).Inc()
		r.metrics.RepositoryQueryDuration.WithLabelValues(op, result).
			Observe(time.Since(started).Seconds())
		span.End()
	}()

	tag, err := tx.Exec(ctxSpc, markSent, id)
	if err != nil {
		result = "error"
		return fmt.Errorf("mark outbox event sent: %w", err)
	}
	if tag.RowsAffected() == 0 {
		result = "error"
		return fmt.Errorf("mark outbox event sent: event %s not found", id)
	}
	return nil
}

func (r *OutboxEventRepository) MarkFailedTx(
	ctx context.Context,
	tx pgx.Tx,
	id uuid.UUID,
	produceErr error,
) error {
	tracer := otel.Tracer("OutboxEventRepository")
	ctxSpc, span := tracer.Start(ctx, "OutboxEventRepository.MarkFailedTx")

	started := time.Now()
	op := "outbox_mark_failed"
	result := MetricsResultSuccess

	defer func() {
		r.metrics.RepositoryQueryTotal.WithLabelValues(op, result).Inc()
		r.metrics.RepositoryQueryDuration.WithLabelValues(op, result).
			Observe(time.Since(started).Seconds())
		span.End()
	}()

	errMsg := ""
	if produceErr != nil {
		errMsg = produceErr.Error()
	}

	tag, err := tx.Exec(ctxSpc, markFailed, id, errMsg, OutboxMaxAttempts)
	if err != nil {
		result = "error"
		return fmt.Errorf("mark outbox event failed: %w", err)
	}
	if tag.RowsAffected() == 0 {
		result = "error"
		return fmt.Errorf("mark outbox event failed: event %s not found", id)
	}
	return nil
}
