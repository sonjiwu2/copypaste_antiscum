// Package logging настраивает структурированный логгер приложения.
package logging

import (
	"io"
	"log/slog"
	"strings"
)

// New создаёт JSON-логгер с указанным уровнем. Неизвестный уровень
// интерпретируется как info, чтобы опечатка в окружении не гасила логи.
func New(output io.Writer, level string) *slog.Logger {
	return slog.New(slog.NewJSONHandler(output, &slog.HandlerOptions{
		Level: parseLevel(level),
	}))
}

func parseLevel(level string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
