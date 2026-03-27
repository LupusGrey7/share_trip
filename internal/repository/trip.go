package repository

import (
	"context"
	"fmt"
	"github.com/gofiber/fiber/v2/log"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"job4j.ru/share_trip/internal/domain/trip"
	"time"
)

const (
	getTripByID = `select * from trips where id = $1`

	createNewTrip = `
insert into trips(id, driver_id, from_point, to_point, departure_time, seats, status, created_at) 
values($1, $2, $3, $4, $5, $6, $7, $8) 
RETURNING 
id, driver_id, from_point, to_point, departure_time, seats, status, created_at`

	createTripHistory = `
insert into trip_history(id, trip_id, from_status, to_status, created_at) 
values($1, $2, $3, $4, $5)`

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

type TripRepository struct {
	pool *pgxpool.Pool
}

func NewTripRepository(pool *pgxpool.Pool) *TripRepository {
	return &TripRepository{pool: pool}
}

func (r *TripRepository) CreateTrip(ctx context.Context, t *trip.Entity) (trip.Entity, error) {
	var entity trip.Entity

	query := createNewTrip
	id := uuid.New()
	args := []interface{}{id, t.DriverID, t.FromPoint, t.ToPoint, time.Now(), t.Seats, t.Status, time.Now()}
	argsRslRow := []interface{}{&entity.ID, &entity.DriverID, &entity.FromPoint, &entity.ToPoint, &entity.DepartureTime, &entity.Seats, &entity.Status, &entity.CreatedAt}

	err := r.pool.QueryRow(ctx, query, args...).Scan(argsRslRow...)
	if err != nil {
		return trip.Entity{}, fmt.Errorf("ошибка при вставке: %w", err)
	}

	id = uuid.New()
	query = createTripHistory
	argsTHistory := []interface{}{id, entity.ID, trip.StatusDraft, entity.Status, time.Now()}

	_, err = r.pool.Query(ctx, query, argsTHistory...)
	if err != nil {
		return trip.Entity{}, fmt.Errorf("ошибка при вставке trip_history: %w", err)
	}

	return entity, nil
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

func (r *TripRepository) GetById(ctx context.Context, id string) (trip.Entity, error) {
	var entity trip.Entity

	query := getTripByID

	rows, err := r.pool.Query(ctx, query, id)
	if err != nil {
		return trip.Entity{}, fmt.Errorf("error while query: %w", err)
	}
	defer rows.Close()

	//Критически важно: переходим на первую строку
	if !rows.Next() {
		return trip.Entity{}, fmt.Errorf("trip with id %s not found", id)
	}

	argsRslRow := []interface{}{
		&entity.ID, &entity.DriverID, &entity.FromPoint, &entity.ToPoint,
		&entity.DepartureTime, &entity.Seats, &entity.Status, &entity.CreatedAt,
	}
	err = rows.Scan(argsRslRow...)
	if err != nil {
		log.Error("error : ", err)
		return trip.Entity{}, err
	}

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

func (r *TripRepository) GetForUpdateByID(
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

// BeginWithTx - create DB transactional, выполняет переданную функцию внутри транзакции.
func (r *TripRepository) BeginWithTx(
	ctx context.Context,
	fn func(tx pgx.Tx) error, //функция
) error {
	// 2. Начинаем транзакцию
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		log.Fatal(err)
		return err
	}

	//3. Гарантируем откат, если не вызван Commit
	defer func() {
		if err != nil {
			err := tx.Rollback(ctx)
			if err != nil {
				log.Fatal(err)
			}
		}
	}()

	// 4. Передаем tx в метод
	err = fn(tx)
	if err != nil {
		return err
	}

	// 5. Фиксируем транзакцию
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}
	return nil
}

// Tx - выполняет переданную функцию в транзакции.
// Если функция вернёт ошибку – откат, иначе коммит.
func (r *TripRepository) Tx(ctx context.Context, block func(tx pgx.Tx) error) error {

	txBegin, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	err = block(txBegin)
	if err != nil {
		// Если блок вернул ошибку — откатываем
		if rbErr := txBegin.Rollback(ctx); rbErr != nil {
			log.Error("rollback error: %v", rbErr)
		}
		return fmt.Errorf("error func block() TX: %w", err)
	}

	if err = txBegin.Commit(ctx); err != nil {
		// Если коммит не удался, тоже пробуем откатить (хотя это может не сработать)
		if rbErr := txBegin.Rollback(ctx); rbErr != nil {
			log.Error("rollback error: %v", rbErr)
		}
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
