package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2/log"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"job4j.ru/share_trip/internal/domain/trip"
)

const (
	createNewTrip = `
insert into trips(id, driver_id, from_point, to_point, departure_time, seats, status, created_at) 
values($1, $2, $3, $4, $5, $6, $7, $8) 
RETURNING 
id, driver_id, from_point, to_point, departure_time, seats, status, created_at`

	createTripHistory = `
insert into trip_history(id, trip_id, from_status, to_status, created_at) 
values($1, $2, $3, $4, $5)`
)

type BaseTripRepository interface {
	CreateTrip(context.Context, *trip.Entity) (*trip.Entity, error)
	//GetByID(ctx context.Context, id string) (*trip.Entity, error)
	Tx(ctx context.Context, block func(tx pgx.Tx) error) error
}

type TripRepository struct {
	pool *pgxpool.Pool
}

func NewTripRepository(pool *pgxpool.Pool) *TripRepository {
	return &TripRepository{pool: pool}
}

func (r *TripRepository) CreateTrip(ctx context.Context, t *trip.Entity) (*trip.Entity, error) {
	var entity = &trip.Entity{}

	query := createNewTrip
	id := uuid.New()
	args := []interface{}{id, t.DriverID, t.FromPoint, t.ToPoint, time.Now(), t.Seats, t.Status, time.Now()}
	argsRslRow := []interface{}{&entity.ID, &entity.DriverID, &entity.FromPoint, &entity.ToPoint, &entity.DepartureTime, &entity.Seats, &entity.Status, &entity.CreatedAt}

	err := r.pool.QueryRow(ctx, query, args...).Scan(argsRslRow...)
	if err != nil {
		return &trip.Entity{}, fmt.Errorf("ошибка при вставке: %w", err)
	}

	id = uuid.New()
	query = createTripHistory
	argsTHistory := []interface{}{id, entity.ID, trip.StatusDraft, entity.Status, time.Now()}

	_, err = r.pool.Query(ctx, query, argsTHistory...)
	if err != nil {
		return &trip.Entity{}, fmt.Errorf("ошибка при вставке trip_history: %w", err)
	}

	return entity, nil
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
