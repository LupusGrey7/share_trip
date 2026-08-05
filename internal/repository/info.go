package repository

import (
	"context"
	"fmt"

	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel"
	"job4j.ru/share_trip/internal/observability/logctx"
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
	ctxSpc, span := otel.Tracer("Repository").Start(ctx, "Repository.GetDbConnectInfo")
	defer span.End()

	// getting custom logger context
	logger := logctx.Logger(ctxSpc).With(
		slog.String("layer", "repository"),
		slog.String("repository", "RepoPg.GetDbConnectInfo"),
	)
	logger.Debug("database ping started")

	err := r.pool.Ping(ctxSpc)
	if err != nil {
		logger.Error("database ping failed", slog.Any("error", err))
		return "", fmt.Errorf("database ping failed: %w", err)
	}

	logger.Debug("database ping completed")
	return "OK", nil
}
