package httpapi

import (
	"context"
	"log/slog"
)

type contextKey int

const (
	requestIDKey contextKey = iota
	loggerKey
)

// RequestIDFrom возвращает идентификатор запроса или пустую строку.
func RequestIDFrom(ctx context.Context) string {
	requestID, _ := ctx.Value(requestIDKey).(string)

	return requestID
}

func withRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey, requestID)
}

// loggerFrom возвращает логгер запроса. Если его нет (вызов вне цепочки
// middleware), отдаём логгер по умолчанию, чтобы не терять запись.
func loggerFrom(ctx context.Context) *slog.Logger {
	logger, ok := ctx.Value(loggerKey).(*slog.Logger)
	if !ok {
		return slog.Default()
	}

	return logger
}

func withLogger(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}
