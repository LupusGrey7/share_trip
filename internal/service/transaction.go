package service

import (
	"context"
	"fmt"
	"job4j.ru/share_trip/internal/observability/logctx"
	"log/slog"

	"github.com/gofiber/fiber/v2/log"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func tx[T interface{}](
	ctx context.Context,
	pool *pgxpool.Pool,
	block func(tx pgx.Tx) (*T, error),
) (*T, error) {
	logger := logctx.Logger(ctx).With(
		slog.String("layer", "transaction"),
	)

	logger.Info("begin transaction")

	txBegin, err := pool.Begin(ctx)
	if err != nil {
		logger.Error(
			"failed to begin transaction",
			slog.Any("error", err),
		)

		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	//defer func() {
	//	err := txBegin.Rollback(ctx)
	//	if err != nil && !errors.Is(err, pgx.ErrTxClosed) {
	//		logger.Error(
	//			"rollback transaction failed",
	//			slog.Any("error", err),
	//		)
	//	}
	//}()

	res, err := block(txBegin)
	if err != nil {
		logger.Error(
			"transaction block failed",
			slog.Any("error", err),
		)

		return nil, fmt.Errorf("err block() txBegin with: %w", err)
	}

	if err = txBegin.Commit(ctx); err != nil {
		// Если коммит не удался, тоже пробуем откатить (хотя это может не сработать)
		logger.Error(
			"failed to commit transaction",
			slog.Any("error", err),
		)

		if rbErr := txBegin.Rollback(ctx); rbErr != nil {
			log.Error("rollback error: %v", rbErr)
			logger.Error(
				"rollback transaction failed",
				slog.Any("error", err),
			)
		}
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	logger.Info("commit transaction")

	return res, nil
}
