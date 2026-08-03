// Package app собирает зависимости приложения и управляет жизненным циклом HTTP-сервера.
package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/config"
	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/httpapi"
	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/platform/identifier"
)

// Application владеет собранным HTTP-сервером.
type Application struct {
	config config.Config
	logger *slog.Logger
	server *http.Server
}

// New собирает приложение из конфигурации и логгера.
func New(cfg config.Config, logger *slog.Logger) *Application {
	handler := httpapi.NewRouter(httpapi.RouterDeps{
		Logger:          logger,
		RequestIDs:      identifier.Random{},
		MaxRequestBytes: cfg.MaxRequestBytes,
	})

	return &Application{
		config: cfg,
		logger: logger,
		server: &http.Server{
			Addr:              cfg.HTTPAddr,
			Handler:           handler,
			ReadHeaderTimeout: cfg.ReadHeaderTimeout,
			ReadTimeout:       cfg.ReadTimeout,
			WriteTimeout:      cfg.WriteTimeout,
			IdleTimeout:       cfg.IdleTimeout,
		},
	}
}

// Handler возвращает HTTP-обработчик приложения. Нужен интеграционным тестам,
// которым не требуется поднимать реальный слушающий сокет.
func (a *Application) Handler() http.Handler {
	return a.server.Handler
}

// Run запускает сервер и завершает его при отмене контекста.
// Возврат происходит только после остановки сервера.
func (a *Application) Run(ctx context.Context) error {
	serverFailed := make(chan error, 1)

	go func() {
		a.logger.InfoContext(ctx, "http-сервер запущен", slog.String("addr", a.config.HTTPAddr))

		if err := a.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverFailed <- fmt.Errorf("запустить http-сервер: %w", err)

			return
		}

		serverFailed <- nil
	}()

	select {
	case err := <-serverFailed:
		return err
	case <-ctx.Done():
		return a.shutdown()
	}
}

// shutdown даёт активным запросам завершиться в пределах таймаута.
func (a *Application) shutdown() error {
	a.logger.Info("остановка http-сервера")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), a.config.ShutdownTimeout)
	defer cancel()

	if err := a.server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("остановить http-сервер: %w", err)
	}

	a.logger.Info("http-сервер остановлен")

	return nil
}
