package repository

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
)

type InfoRepository interface {
	GetDbConnectInfo(ctx context.Context) (string, error)
}

type RepoPg struct {
	pool *pgxpool.Pool
}

func NewRepoPg(pool *pgxpool.Pool) *RepoPg {
	return &RepoPg{pool: pool}
}

func (r *RepoPg) GetDbConnectInfo(ctx context.Context) (string, error) {
	err := r.pool.Ping(ctx)
	if err != nil {
		return "", fmt.Errorf("database ping failed: %w", err)
	}

	return "OK", err
}
