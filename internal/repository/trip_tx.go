// trip repository

package repository

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"go.opentelemetry.io/otel"

	"job4j.ru/share_trip/internal/domain/trip/model"
	"job4j.ru/share_trip/internal/observability/logctx"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

const (
	ErrSelectEntityFailed = "select trip failed"
	MetricsResultSuccess  = "success"

	getTripByID = `
select *
from trips 
where id = $1
`

	forUpdateTrip = `
select
	id,
	driver_id,
	from_point,
	to_point,
	departure_time,
	seats,
	status,
	created_at
from trips
where id = $1 
FOR UPDATE
`
	updateTrip = `
update trips
set status = $1
where id = $2
RETURNING 
id, driver_id, from_point, to_point, departure_time, seats, status, created_at
`
	updateTripHistory = `
update trip_history
set to_status = $1
where trip_id = $2
`
)

type BaseTxTripRepository interface {
	GetTripByID(ctx context.Context, tx pgx.Tx, id string) (*model.Entity, error)
	GetForUpdateByIDTx(ctx context.Context, tx pgx.Tx, id string) (*model.Entity, error)
	UpdateTripTx(ctx context.Context, tx pgx.Tx, t *model.Entity) (*model.Entity, error)
	CreateTripDraftTx(ctx context.Context, tx pgx.Tx, t *model.Entity) (*model.Entity, error)
}

func (r *TripRepository) GetTripByID(
	ctx context.Context,
	tx pgx.Tx,
	id string,
) (*model.Entity, error) {
	//tracing Jaeger
	tracer := otel.Tracer("TripRepository")
	ctxSpc, span := tracer.Start(ctx, "TripRepository.GetByID")
	defer span.End()

	//getting custom logger context
	logger := logctx.Logger(ctxSpc).With(
		slog.String("layer", "repository"),
		slog.String("repository", "TripRepository"),
		slog.String("operation", "GetByID"),
		slog.String("trip_id", id),
	)
	logger.Debug("select trip by ID started")

	var entity model.Entity

	query := getTripByID
	rows, err := r.pool.Query(ctxSpc, query, id)
	if err != nil {
		logger.Error(
			ErrSelectEntityFailed,
			slog.Any("error", err),
		)
		return &model.Entity{}, fmt.Errorf("error while query: %w", err)
	}
	defer rows.Close()

	//Critical: Jump to the first line
	if !rows.Next() {
		logger.Error(
			ErrSelectEntityFailed,
			slog.Any("error", err),
		)
		return nil, ErrTripNotFound
	}

	argsRslRow := []interface{}{
		&entity.ID, &entity.DriverID, &entity.FromPoint, &entity.ToPoint,
		&entity.DepartureTime, &entity.Seats, &entity.Status, &entity.CreatedAt,
	}
	err = rows.Scan(argsRslRow...)
	if err != nil {
		logger.Error(
			ErrSelectEntityFailed,
			slog.Any("error", err),
		)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrTripNotFound
		}

		return nil, fmt.Errorf(errQueryByID, id, err)
	}

	logger.Debug("get trip by ID completed")
	return &entity, nil
}

func (r *TripRepository) CreateTripDraftTx(
	ctx context.Context,
	tx pgx.Tx,
	t *model.Entity,
) (*model.Entity, error) {
	//tracing Jaeger
	tracer := otel.Tracer("TripRepository")
	ctxSpc, span := tracer.Start(ctx, "TripRepository.CreateTripDraftTx")

	// prometheus
	started := time.Now()
	name := "repo_trips_create_duration_seconds" //metric name
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
		slog.String("repository", "TripRepository"),
		slog.String("operation", "CreateTripDraftTx"),
		slog.String("trip_id", t.ID.String()),
		slog.String("client_id", t.DriverID.String()),
	)
	logger.Debug("create trip draft started")

	entity := &model.Entity{} // Create an empty structure on the stack
	query := createNewTrip
	id := uuid.New()
	args := []interface{}{id, t.DriverID, t.FromPoint, t.ToPoint, t.DepartureTime, t.Seats, t.Status, time.Now()}
	argsRslRow := []interface{}{&entity.ID, &entity.DriverID, &entity.FromPoint, &entity.ToPoint, &entity.DepartureTime, &entity.Seats, &entity.Status, &entity.CreatedAt}

	err := tx.QueryRow(ctxSpc, query, args...).Scan(argsRslRow...)
	if err != nil {
		logger.Error("create trip draft failed", slog.Any("error", err))
		return &model.Entity{}, fmt.Errorf("insert new trip failed: %w", err)
	}

	id = uuid.New()
	query = createTripHistory
	argsTHistory := []interface{}{id, entity.ID, model.StatusDraft, entity.Status, time.Now()}

	rows, err = tx.Query(ctxSpc, query, argsTHistory...)
	if err != nil {
		logger.Error("create trip_history repository failed", slog.Any("error", err))
		return &model.Entity{}, fmt.Errorf("error while create trip_history: %w", err)
	}
	defer rows.Close() // process rows sql

	logger.Debug("create trip draft completed", slog.String("trip_id", entity.ID.String()))
	return entity, nil
}

func (r *TripRepository) UpdateTripTx(
	ctx context.Context,
	tx pgx.Tx,
	t *model.Entity,
) (*model.Entity, error) {
	//tracing Jaeger
	tracer := otel.Tracer("TripRepository")
	ctxSpc, span := tracer.Start(ctx, "TripRepository.UpdateTripTx")

	//prometheus log
	started := time.Now()                        //metric time
	name := "repo_trips_update_duration_seconds" //metric name
	result := MetricsResultSuccess               //metric result
	var rows pgx.Rows                            // for history to defer

	defer func() {
		rows.Close() // process rows

		r.metrics.RepositoryQueryTotal.
			WithLabelValues(name, result).
			Inc()
		r.metrics.RepositoryQueryDuration.
			WithLabelValues(name, result).
			Observe(time.Since(started).Seconds())

		span.End() // always in the end
	}()

	//getting custom logger context
	logger := logctx.Logger(ctxSpc).With(
		slog.String("layer", "repository"),
		slog.String("repository", "TripRepository"),
		slog.String("operation", "UpdateTripTx"),
		slog.String("trip_id", t.ID.String()),
	)
	logger.Debug("update trip started")

	var entity model.Entity // Create an empty structure on the stack
	query := updateTrip
	args := []interface{}{t.Status, t.ID}

	err := tx.QueryRow(ctxSpc, query, args...).Scan(
		&entity.ID,
		&entity.DriverID,
		&entity.FromPoint,
		&entity.ToPoint,
		&entity.DepartureTime,
		&entity.Seats,
		&entity.Status,
		&entity.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			logger.Error(
				"update trip by id failed",
				slog.Any("error", err),
			)
			return nil, ErrTripNotFound
		}
		return nil, fmt.Errorf(errQueryByID, t.ID, err)
	}

	args = []interface{}{t.Status, t.ID}
	query = updateTripHistory

	rows, err = tx.Query(ctxSpc, query, args...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			logger.Error(
				"update trip history by id failed",
				slog.Any("error", err),
			)
			return nil, ErrTripNotFound
		}
		return nil, fmt.Errorf(errQueryTripHistoryByID, t.ID, err)
	}

	logger.Debug("update trip completed")
	return &entity, nil
}

func (r *TripRepository) GetForUpdateByIDTx(
	ctx context.Context,
	tx pgx.Tx, //transaction
	id string,
) (*model.Entity, error) {
	//tracing Jaeger
	tracer := otel.Tracer("TripRepository")
	ctxSpc, span := tracer.Start(ctx, "TripRepository.GetForUpdateByIDTx")

	//prometheus log
	started := time.Now()                      //metric time
	name := "repo_trips_lock_duration_seconds" //metric name
	result := MetricsResultSuccess             //metric result

	defer func() {
		r.metrics.RepositoryQueryTotal.
			WithLabelValues(name, result).
			Inc()
		r.metrics.RepositoryQueryDuration.
			WithLabelValues(name, result).
			Observe(time.Since(started).Seconds())

		span.End() // always in the end
	}()

	//getting custom logger context
	logger := logctx.Logger(ctxSpc).With(
		slog.String("layer", "repository"),
		slog.String("repository", "TripRepository"),
		slog.String("operation", "GetForUpdateByIDTx"),
		slog.String("trip_id", id),
	)
	logger.Debug("select trip for update started")

	tp := model.Entity{} // We create an empty structure on the stack (analogous to - var tp trip.Entity)
	query := forUpdateTrip

	err := tx.QueryRow(ctxSpc, query, id).Scan(
		&tp.ID,
		&tp.DriverID,
		&tp.FromPoint,
		&tp.ToPoint,
		&tp.DepartureTime,
		&tp.Seats,
		&tp.Status,
		&tp.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			logger.Error(
				ErrSelectEntityFailed,
				slog.Any("error", err),
			)
			return nil, ErrTripNotFound
		}
		return nil, fmt.Errorf(errQueryByID, id, err)
	}

	logger.Debug("select trip for update completed")
	return &tp, nil // Return a pointer to the filled structure
}
