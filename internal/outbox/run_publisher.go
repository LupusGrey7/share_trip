package outbox

import (
	"context"
	"log/slog"
	"time"
)

func (p *Publisher) Run(ctx context.Context) error {
	slog.Info("outbox publisher started",
		slog.Duration("interval", p.interval),
		slog.Int("batch_size", p.batchSize),
	)

	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			slog.Info("outbox publisher stopped", slog.Any("reason", ctx.Err()))
			return ctx.Err()
		case <-ticker.C:
			if err := p.PublishBatch(ctx); err != nil {
				slog.Error("publish outbox batch", slog.Any("error", err))
			}
		}
	}
}
