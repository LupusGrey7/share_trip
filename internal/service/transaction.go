package service

import (
	"context"
	"fmt"
	"github.com/gofiber/fiber/v2/log"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func tx[T interface{}](
	ctx context.Context,
	pool *pgxpool.Pool,
	block func(tx pgx.Tx) (*T, error),
) (*T, error) {
	txBegin, err := pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	res, err := block(txBegin)
	if err != nil {
		return nil, fmt.Errorf("err block() txBegin with: %w", err)
	}

	if err = txBegin.Commit(ctx); err != nil {
		// Если коммит не удался, тоже пробуем откатить (хотя это может не сработать)
		if rbErr := txBegin.Rollback(ctx); rbErr != nil {
			log.Error("rollback error: %v", rbErr)
		}
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return res, nil
}
