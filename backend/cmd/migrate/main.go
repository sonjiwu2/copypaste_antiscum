// Команда migrate применяет миграции схемы PostgreSQL.
//
// Отдельный бинарник нужен, чтобы схема менялась ровно один раз перед
// стартом сервиса: миграции не выполняются ни на каждый HTTP-запрос,
// ни неявно при подъёме нескольких экземпляров api.
//
//	migrate up        применить все миграции
//	migrate down      откатить последнюю миграцию
//	migrate status    показать применённые и ожидающие миграции
//	migrate version   показать текущую версию схемы
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/config"
	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/storage/postgres"
)

// defaultCommand выбран так, чтобы контейнер миграций в Compose запускался
// без аргументов и делал самое ожидаемое действие.
const defaultCommand = postgres.CommandUp

func main() {
	if err := run(); err != nil {
		slog.Error("миграция завершилась с ошибкой", slog.String("error", err.Error()))
		os.Exit(1)
	}
}

// run вынесен из main, чтобы отложенные вызовы отрабатывали до os.Exit.
func run() error {
	command := defaultCommand
	if len(os.Args) > 1 {
		command = os.Args[1]
	}

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("загрузить конфигурацию: %w", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	connectCtx, cancel := postgres.WithTimeout(ctx, cfg.Database.ConnectTimeout)
	defer cancel()

	db, err := postgres.OpenSQL(connectCtx, cfg.Database.URL)
	if err != nil {
		return err
	}

	defer func() {
		if err := db.Close(); err != nil {
			slog.Warn("не удалось закрыть соединение с базой", slog.String("error", err.Error()))
		}
	}()

	return postgres.RunMigration(ctx, db, command, os.Stdout)
}
