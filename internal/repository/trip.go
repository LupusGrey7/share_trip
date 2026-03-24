package repository

import (
	"context"
	"fmt"
	"github.com/gofiber/fiber/v2/log"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"job4j.ru/share_trip/internal/domain/trip"
	"time"
)

type TripRepository struct {
	pool *pgxpool.Pool
}

func NewTripRepository(pool *pgxpool.Pool) *TripRepository {
	return &TripRepository{pool: pool}
}

func (r *TripRepository) CreateTrip(ctx context.Context, t *trip.Entity) (trip.Entity, error) {
	var entity trip.Entity

	query := `insert into trips(id, driver_id, from_point, to_point, departure_time, seats, status, created_at) 
values($1, $2, $3, $4, $5, $6, $7, $8) 
RETURNING 
id, driver_id, from_point, to_point, departure_time, seats, status, created_at`
	id := uuid.New()
	args := []interface{}{id, t.DriverID, t.FromPoint, t.ToPoint, time.Now(), t.Seats, t.Status, time.Now()}
	argsRslRow := []interface{}{&entity.ID, &entity.DriverID, &entity.FromPoint, &entity.ToPoint, &entity.DepartureTime, &entity.Seats, &entity.Status, &entity.CreatedAt}

	err := r.pool.QueryRow(ctx, query, args...).Scan(argsRslRow...)
	if err != nil {
		return trip.Entity{}, fmt.Errorf("ошибка при вставке: %w", err)
	}

	id = uuid.New()
	query = `insert into trip_history(id, trip_id, from_status, to_status, created_at) values($1, $2, $3, $4, $5)`
	argsTHistory := []interface{}{id, entity.ID, trip.Draft, entity.Status, time.Now()}

	_, err = r.pool.Query(ctx, query, argsTHistory...)
	if err != nil {
		return trip.Entity{}, fmt.Errorf("ошибка при вставке trip_history: %w", err)
	}

	return entity, nil
}

func (r *TripRepository) GetById(ctx context.Context, id string) (trip.Entity, error) {
	var entity trip.Entity

	query := `select * from trips where id = $1`

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
