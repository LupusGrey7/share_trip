package repository

import (
	"context"
	"errors"
	"fmt"
	"job4j.ru/share_trip/internal/domain/trip/model"
	"job4j.ru/share_trip/internal/observability/logctx"
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v2/log"
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
	GetByID(ctx context.Context, tx pgx.Tx, id uuid.UUID) (*model.Entity, error)
	GetForUpdateByIDTx(ctx context.Context, tx pgx.Tx, id uuid.UUID) (*model.Entity, error)
	UpdateTripTx(ctx context.Context, tx pgx.Tx, t *model.Entity) (*model.Entity, error)
	CreateTripTx(ctx context.Context, tx pgx.Tx, t *model.Entity) (*model.Entity, error)
}

func (r *TripRepository) GetByID(
	ctx context.Context,
	tx pgx.Tx,
	id uuid.UUID,
) (*model.Entity, error) {
	var entity model.Entity

	query := getTripByID
	rows, err := r.pool.Query(ctx, query, id)
	if err != nil {
		return &model.Entity{}, fmt.Errorf("error while query: %w", err)
	}
	defer rows.Close()

	//Критически важно: переходим на первую строку
	if !rows.Next() {
		return &model.Entity{}, fmt.Errorf("trip with id %s not found", id)
	}

	argsRslRow := []interface{}{
		&entity.ID, &entity.DriverID, &entity.FromPoint, &entity.ToPoint,
		&entity.DepartureTime, &entity.Seats, &entity.Status, &entity.CreatedAt,
	}
	err = rows.Scan(argsRslRow...)
	if err != nil {
		log.Error("error : ", err)
		return &model.Entity{}, err
	}

	return &entity, nil
}

func (r *TripRepository) CreateTripTx(
	ctx context.Context,
	tx pgx.Tx, // транзакция
	t *model.Entity,
) (*model.Entity, error) {
	logger := logctx.Logger(ctx).With(
		slog.String("layer", "repository"),
		slog.String("repository", "TripRepository"),
		slog.String("operation", "Create"),
		slog.String("trip_id", t.ID.String()),
		slog.String("client_id", t.DriverID.String()),
	)

	logger.Info("insert trip started")

	entity := &model.Entity{} // Создаем пустую структуру в стеке

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
	defer rows.Close() // обработать rows

	logger.Info("insert trip completed")

	return entity, nil
}

func (r *TripRepository) UpdateTripTx(
	ctx context.Context,
	tx pgx.Tx,
	t *model.Entity,
) (*model.Entity, error) {
	var entity model.Entity // Создаем пустую структуру в стеке
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
			return nil, ErrTripNotFound
		}
		return nil, fmt.Errorf("query trip by id %s: %w", t.ID, err)
	}

	args = []interface{}{t.Status, t.ID}
	query = updateTripHistory
	rows, err := tx.Query(ctx, query, args...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrTripNotFound
		}
		return nil, fmt.Errorf("query trip_history by id %s: %w", t.ID, err)
	}
	defer rows.Close() // обработать rows

	return &entity, nil
}

func (r *TripRepository) GetForUpdateByIDTx(
	ctx context.Context,
	tx pgx.Tx, // транзакция
	id uuid.UUID,
) (*model.Entity, error) {
	tp := model.Entity{} // Создаем пустую структуру в стеке (аналог - var tp trip.Entity)

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
			return nil, ErrTripNotFound
		}
		return nil, fmt.Errorf("query trip by id %s: %w", id, err)
	}
	return &tp, nil // Возвращаем указатель на заполненную структуру
}
