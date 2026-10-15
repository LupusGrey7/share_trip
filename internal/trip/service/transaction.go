// transaction function - package service

package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"job4j.ru/share_trip/internal/observability/logctx"
)

func tx[T interface{}](
	ctx context.Context,
	pool *pgxpool.Pool,
	block func(tx pgx.Tx) (*T, error),
) (*T, error) {
	txBegin, err := pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		// Rollback after a successful Commit returns pgx.ErrTxClosed.
		if rbErr := txBegin.Rollback(ctx); rbErr != nil && !errors.Is(rbErr, pgx.ErrTxClosed) {
			logctx.Logger(ctx).Error("rollback failed", slog.Any("error", rbErr)) // ignored error
		}
	}()

	res, err := block(txBegin)
	if err != nil {
		return nil, fmt.Errorf("tx block: %w", err)
	}
	if err = txBegin.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}
	return res, nil
}
