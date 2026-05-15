// trip repository

package repository

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"job4j.ru/share_trip/internal/domain/trip/model"
	"job4j.ru/share_trip/internal/observability/logctx"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

const (
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
	GetByID(ctx context.Context, tx pgx.Tx, id string) (*model.Entity, error)
	GetForUpdateByIDTx(ctx context.Context, tx pgx.Tx, id string) (*model.Entity, error)
	UpdateTripTx(ctx context.Context, tx pgx.Tx, t *model.Entity) (*model.Entity, error)
	CreateTripTx(ctx context.Context, tx pgx.Tx, t *model.Entity) (*model.Entity, error)
}

func (r *TripRepository) GetByID(
	ctx context.Context,
	tx pgx.Tx,
	id string,
) (*model.Entity, error) {
	//getting custom logger context
	logger := logctx.Logger(ctx).With(
		slog.String("layer", "repository"),
		slog.String("repository", "TripRepository"),
		slog.String("operation", "GetByID"),
		slog.String("trip_id", id),
	)

	logger.Info("select trip started")

	var entity model.Entity

	query := getTripByID
	rows, err := r.pool.Query(ctx, query, id)
	if err != nil {
		logger.Error(
			"select trip failed",
			slog.Any("error", err),
		)
		return &model.Entity{}, fmt.Errorf("error while query: %w", err)
	}
	defer rows.Close()

	//Critical: Jump to the first line
	if !rows.Next() {
		logger.Error(
			"select trip failed",
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
			"select trip failed",
			slog.Any("error", err),
		)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrTripNotFound
		}

		return nil, fmt.Errorf(errQueryByID, id, err)
	}

	logger.Info("get trip completed")
	return &entity, nil
}

func (r *TripRepository) CreateTripTx(
	ctx context.Context,
	tx pgx.Tx, // транзакция
	t *model.Entity,
) (*model.Entity, error) {
	//getting custom logger context
	logger := logctx.Logger(ctx).With(
		slog.String("layer", "repository"),
		slog.String("repository", "TripRepository"),
		slog.String("operation", "CreateTripTx"),
		slog.String("trip_id", t.ID.String()),
		slog.String("client_id", t.DriverID.String()),
	)

	logger.Info("insert trip started")

	entity := &model.Entity{} // Create an empty structure on the stack

	query := createNewTrip
	id := uuid.New()
	args := []interface{}{id, t.DriverID, t.FromPoint, t.ToPoint, time.Now(), t.Seats, t.Status, time.Now()}
	argsRslRow := []interface{}{&entity.ID, &entity.DriverID, &entity.FromPoint, &entity.ToPoint, &entity.DepartureTime, &entity.Seats, &entity.Status, &entity.CreatedAt}

	err := tx.QueryRow(ctx, query, args...).Scan(argsRslRow...)
	if err != nil {
		logger.Error(
			"insert trip failed",
			slog.Any("error", err),
		)
		return &model.Entity{}, fmt.Errorf("ошибка при вставке: %w", err)
	}

	id = uuid.New()
	query = createTripHistory
	argsTHistory := []interface{}{id, entity.ID, model.StatusDraft, entity.Status, time.Now()}

	rows, err := tx.Query(ctx, query, argsTHistory...)
	if err != nil {
		logger.Error(
			"insert trip_history failed",
			slog.Any("error", err),
		)
		return &model.Entity{}, fmt.Errorf("ошибка при вставке trip_history: %w", err)
	}
	defer rows.Close() // process rows

	logger.Info("insert trip completed")
	return entity, nil
}

func (r *TripRepository) UpdateTripTx(
	ctx context.Context,
	tx pgx.Tx,
	t *model.Entity,
) (*model.Entity, error) {
	var entity model.Entity // Create an empty structure on the stack
	//getting custom logger context
	logger := logctx.Logger(ctx).With(
		slog.String("layer", "repository"),
		slog.String("repository", "TripRepository"),
		slog.String("operation", "UpdateTripTx"),
		slog.String("trip_id", t.ID.String()),
	)
	logger.Info("update trip started")

	query := updateTrip
	args := []interface{}{t.Status, t.ID}

	err := tx.QueryRow(ctx, query, args...).Scan(
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
	rows, err := tx.Query(ctx, query, args...)
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
	defer rows.Close() // process rows

	logger.Info("update trip completed")
	return &entity, nil
}

func (r *TripRepository) GetForUpdateByIDTx(
	ctx context.Context,
	tx pgx.Tx, //transaction
	id string,
) (*model.Entity, error) {
	tp := model.Entity{} // We create an empty structure on the stack (analogous to - var tp trip.Entity)
	//getting custom logger context
	logger := logctx.Logger(ctx).With(
		slog.String("layer", "repository"),
		slog.String("repository", "TripRepository"),
		slog.String("operation", "GetForUpdateByIDTx"),
		slog.String("trip_id", id),
	)

	logger.Info("select trip started")

	query := forUpdateTrip
	err := tx.QueryRow(ctx, query, id).Scan(
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
				"select trip failed",
				slog.Any("error", err),
			)
			return nil, ErrTripNotFound
		}
		return nil, fmt.Errorf(errQueryByID, id, err)
	}

	logger.Info("select trip completed")
	return &tp, nil // Return a pointer to the filled structure
}
