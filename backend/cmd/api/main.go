// Команда api запускает HTTP-сервер антискам-тренажёра.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/app"
	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/config"
	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/platform/logging"
)

func main() {
	if err := run(); err != nil {
		slog.Error("приложение завершилось с ошибкой", slog.String("error", err.Error()))
		os.Exit(1)
	}
}

// run вынесен из main, чтобы отложенные вызовы отрабатывали до os.Exit.
func run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("загрузить конфигурацию: %w", err)
	}

	logger := logging.New(os.Stdout, cfg.LogLevel)

	application, err := app.New(cfg, logger)
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	return application.Run(ctx)
}
