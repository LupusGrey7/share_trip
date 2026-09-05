// repository/outbox_repo.go

package storage

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"go.opentelemetry.io/otel"
	"job4j.ru/share_trip/internal/observability/logctx"
	"job4j.ru/share_trip/internal/outbox/domain"

	"github.com/jackc/pgx/v5"
	"job4j.ru/share_trip/internal/observability/metrics"
)

const (
	createEvent = `
insert into outbox_event(event_id, event_name, aggregate_id, payload, created_at)
values($1, $2, $3, $4, $5)
`
)

type OutboxRepository interface {
	CreateOutboxEventTripPublishTx(ctx context.Context, tx pgx.Tx, o *domain.Entity) error
}

type OutboxEventRepository struct {
	metrics *metrics.Metrics
}

func NewOutboxEventRepository(m *metrics.Metrics) *OutboxEventRepository {
	return &OutboxEventRepository{
		metrics: m,
	}
}

func (r *OutboxEventRepository) CreateOutboxEventTripPublishTx(ctx context.Context, tx pgx.Tx, o *domain.Entity) error {
	//tracing Jaeger
	tracer := otel.Tracer("OutboxEventRepository")
	ctxSpc, span := tracer.Start(ctx, "OutboxEventRepository.CreateOutboxEventTripPublishTx")

	// prometheus
	started := time.Now()
	name := "repo_event_create_duration_seconds" //metric name
	result := MetricsResultSuccess               //metric result
	var rows pgx.Rows                            // for history to defer

	defer func() {
		rows.Close() // process rows sql

		r.metrics.RepositoryQueryTotal.
			WithLabelValues(name, result).
			Inc() // Increment the counter for the result
		r.metrics.RepositoryQueryDuration.
			WithLabelValues(name, result).
			Observe(time.Since(started).Seconds()) // Observe the duration of the operation

		span.End() // Span always ends in the end. Jaeger will measure the time between Start and End!
	}()

	//getting custom logger context
	logger := logctx.Logger(ctxSpc).With(
		slog.String("layer", "repository"),
		slog.String("repository", "OutboxEventRepository"),
		slog.String("operation", "CreateOutboxEventTripPublishTx"),
		slog.String("client_id", o.AggregateId.String()),
	)
	logger.Debug("create outbox event trip publish repository started")

	createdAt := o.CreatedAt
	if createdAt.IsZero() {
		createdAt = time.Now()
	}

	query := createEvent
	args := []interface{}{o.EventID, o.EventName, o.AggregateId, o.Payload, createdAt}

	rows, err := tx.Query(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("err when create outbox event trip publish: %w", err)
	}
	defer rows.Close()

	logger.Debug("create outbox event trip publish repository completed")
	return nil
}
