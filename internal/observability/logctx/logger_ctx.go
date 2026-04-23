//Контекст логирования

package logctx

import (
	"context"
	"log/slog"
)

type loggerKey struct{}    // key for Context for Logger
type requestIDKey struct{} //key for Context for RequestID

func WithLogger(ctx context.Context, logger *slog.Logger) context.Context {
	//In the derived context, the value associated with key is val.
	return context.WithValue(ctx, loggerKey{}, logger)
}

func Logger(ctx context.Context) *slog.Logger {
	// ctx.Value - returns the value associated with this context for key,
	logger, ok := ctx.Value(loggerKey{}).(*slog.Logger)
	if !ok || logger == nil {
		return slog.Default() //by default
	}
	return logger
}

func WithRequestID(ctx context.Context, requestID string) context.Context {
	//In the derived context, the value associated with key is val.
	return context.WithValue(ctx, requestIDKey{}, requestID)
}

func RequestID(ctx context.Context) string {
	// ctx.Value - returns the value associated with this context for key,
	value, ok := ctx.Value(requestIDKey{}).(string)
	if !ok {
		return "" //by default
	}
	return value
}
