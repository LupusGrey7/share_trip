package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"job4j.ru/share_trip/internal/domain/trip"
)

const (
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
where id = $1 FOR UPDATE
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
	GetForUpdateByIDTx(ctx context.Context, tx pgx.Tx, id uuid.UUID) (trip.Entity, error)
	UpdateTripTx(ctx context.Context, tx pgx.Tx, tp trip.Entity) (trip.Entity, error)
	CreateTripTx(ctx context.Context, tx pgx.Tx, t *trip.Entity) (trip.Entity, error)
}

func (r *TripRepository) CreateTripTx(
	ctx context.Context,
	tx pgx.Tx, // транзакция
	t *trip.Entity,
) (trip.Entity, error) {
	var entity trip.Entity

	query := createNewTrip
	id := uuid.New()
	args := []interface{}{id, t.DriverID, t.FromPoint, t.ToPoint, time.Now(), t.Seats, t.Status, time.Now()}
	argsRslRow := []interface{}{&entity.ID, &entity.DriverID, &entity.FromPoint, &entity.ToPoint, &entity.DepartureTime, &entity.Seats, &entity.Status, &entity.CreatedAt}

	err := tx.QueryRow(ctx, query, args...).Scan(argsRslRow...)
	if err != nil {
		return trip.Entity{}, fmt.Errorf("ошибка при вставке: %w", err)
	}

	id = uuid.New()
	query = createTripHistory
	argsTHistory := []interface{}{id, entity.ID, trip.StatusDraft, entity.Status, time.Now()}

	rows, err := tx.Query(ctx, query, argsTHistory...)
	if err != nil {
		return trip.Entity{}, fmt.Errorf("ошибка при вставке trip_history: %w", err)
	}
	defer rows.Close() // обработать rows

	return entity, nil
}

func (r *TripRepository) UpdateTripTx(
	ctx context.Context,
	tx pgx.Tx,
	tp trip.Entity,
) (trip.Entity, error) {
	var entity trip.Entity
	query := updateTrip
	args := []interface{}{tp.Status, tp.ID}

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
		return trip.Entity{}, fmt.Errorf(
			"error tx.QueryRow trip by id for update: %w", err,
		)
	}

	args = []interface{}{tp.Status, tp.ID}
	query = updateTripHistory
	rows, err := tx.Query(ctx, query, args...)
	if err != nil {
		return trip.Entity{}, fmt.Errorf("error by update trip_history: %w", err)
	}
	defer rows.Close() // обработать rows

	return entity, nil
}

func (r *TripRepository) GetForUpdateByIDTx(
	ctx context.Context,
	tx pgx.Tx, // транзакция
	id uuid.UUID,
) (trip.Entity, error) {
	var tp trip.Entity

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
		return trip.Entity{}, fmt.Errorf(
			"error tx.QueryRow get trip by id for update: %w", err,
		)
	}
	return tp, nil
}
