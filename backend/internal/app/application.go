// Package app собирает зависимости приложения и управляет жизненным циклом HTTP-сервера.
package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/attempt"
	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/config"
	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/httpapi"
	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/platform/clock"
	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/platform/identifier"
	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/profile"
	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/progress"
	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/scenario"
	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/scenarioarchive"
	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/storage/memory"
	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/storage/postgres"
	"github.com/sonjiwu2/copypaste_antiscum/backend/scenarios"
)

// Application владеет собранным HTTP-сервером и подключением к базе.
type Application struct {
	config config.Config
	logger *slog.Logger
	server *http.Server

	// pool равен nil в режиме хранилища в памяти.
	pool *pgxpool.Pool
}

// storage — собранные хранилища и проверка готовности приложения.
type storage struct {
	attempts         attempt.Repository
	profiles         profile.Repository
	progress         progress.Repository
	scenarioVersions scenario.VersionRepository
	pool             *pgxpool.Pool
	ready            httpapi.ReadinessCheck
}

// New собирает приложение из конфигурации и логгера.
//
// Порядок сборки важен: каталог сценариев проверяется до подключения к базе,
// а версии сценариев синхронизируются до того, как появится первая попытка.
// Неисправная фикстура или изменённая без поднятия версии обязаны остановить
// запуск, а не всплыть при первом запросе пользователя.
func New(ctx context.Context, cfg config.Config, logger *slog.Logger) (*Application, error) {
	catalog, err := scenario.LoadFS(scenarios.Files())
	if err != nil {
		return nil, fmt.Errorf("загрузить сценарии: %w", err)
	}

	scenarioRepository, err := memory.NewScenarioRepository(catalog)
	if err != nil {
		return nil, fmt.Errorf("собрать каталог сценариев: %w", err)
	}

	logger.Info("каталог сценариев загружен", slog.Int("scenarios", len(catalog)))

	built, err := newStorage(ctx, cfg, logger, catalog, scenarioRepository)
	if err != nil {
		return nil, err
	}

	systemClock := clock.System{}
	scenarios := scenario.NewService(scenarioRepository)

	handler := httpapi.NewRouter(httpapi.RouterDeps{
		Logger:          logger,
		RequestIDs:      identifier.Random{},
		MaxRequestBytes: cfg.MaxRequestBytes,
		Ready:           built.ready,
		Cookie: httpapi.CookieSettings{
			Name:   cfg.Cookie.Name,
			MaxAge: int(cfg.Cookie.MaxAge.Seconds()),
			Secure: cfg.Cookie.Secure,
		},
		Profiles:  profile.NewService(built.profiles, systemClock, identifier.Random{}),
		Scenarios: scenarios,
		Progress:  progress.NewService(built.progress, scenarios),
		Attempts: attempt.NewService(
			scenarioRepository,
			built.scenarioVersions,
			built.attempts,
			systemClock,
			identifier.Random{},
		),
	})

	return &Application{
		config: cfg,
		logger: logger,
		pool:   built.pool,
		server: &http.Server{
			Addr:              cfg.HTTPAddr,
			Handler:           handler,
			ReadHeaderTimeout: cfg.ReadHeaderTimeout,
			ReadTimeout:       cfg.ReadTimeout,
			WriteTimeout:      cfg.WriteTimeout,
			IdleTimeout:       cfg.IdleTimeout,
		},
	}, nil
}

// newStorage выбирает хранилище попыток.
//
// Ветвление живёт только здесь: остальные пакеты не знают, где лежат данные.
// Отката на память после сбоя базы нет — иначе пользователь молча терял бы
// прохождения там, где ожидал их сохранения.
func newStorage(
	ctx context.Context,
	cfg config.Config,
	logger *slog.Logger,
	catalog []scenario.Scenario,
	memoryVersions scenario.VersionRepository,
) (storage, error) {
	if !cfg.Cookie.Secure {
		logger.Warn("cookie профиля отправляется без флага Secure: допустимо только для локального демо по HTTP")
	}

	if cfg.StorageDriver == config.StorageMemory {
		logger.Warn("хранилище попыток работает в памяти: данные не переживут перезапуск")

		attempts := memory.NewAttemptRepository()

		return storage{
			attempts:         attempts,
			profiles:         memory.NewProfileRepository(),
			progress:         memory.NewProgressRepository(attempts),
			scenarioVersions: memoryVersions,
			ready:            nil,
		}, nil
	}

	pool, err := postgres.NewPool(ctx, cfg.Database)
	if err != nil {
		return storage{}, fmt.Errorf("подключиться к базе данных: %w", err)
	}

	archive, err := syncScenarioArchive(ctx, pool, cfg, logger, catalog)
	if err != nil {
		// Пул закрывается здесь же: неудачный старт не должен оставлять
		// открытые соединения до завершения процесса.
		pool.Close()

		return storage{}, err
	}

	return storage{
		attempts:         postgres.NewAttemptRepository(pool, cfg.Database.QueryTimeout),
		profiles:         postgres.NewProfileRepository(pool, cfg.Database.QueryTimeout),
		progress:         postgres.NewProgressRepository(pool, cfg.Database.QueryTimeout),
		scenarioVersions: archive,
		pool:             pool,
		ready:            readinessOf(pool, cfg),
	}, nil
}

// syncScenarioArchive переносит версии сценариев в архив.
//
// Попытка ссылается на конкретную версию сценария, поэтому архив обязан
// содержать её раньше, чем появится первая попытка.
func syncScenarioArchive(
	ctx context.Context,
	pool *pgxpool.Pool,
	cfg config.Config,
	logger *slog.Logger,
	catalog []scenario.Scenario,
) (*postgres.ScenarioArchiveRepository, error) {
	versions, err := scenarioarchive.Load(scenarios.Files(), catalog)
	if err != nil {
		return nil, fmt.Errorf("подготовить версии сценариев: %w", err)
	}

	archive := postgres.NewScenarioArchiveRepository(pool, cfg.Database.QueryTimeout)

	if err := archive.Sync(ctx, versions); err != nil {
		return nil, fmt.Errorf("синхронизировать версии сценариев: %w", err)
	}

	logger.Info("версии сценариев синхронизированы", slog.Int("versions", len(versions)))

	return archive, nil
}

// readinessOf собирает проверку готовности к работе.
// Проверка намеренно дешёвая: её вызывает healthcheck контейнера.
func readinessOf(pool *pgxpool.Pool, cfg config.Config) httpapi.ReadinessCheck {
	return func(ctx context.Context) error {
		return postgres.Ping(ctx, pool, cfg.Database.HealthTimeout)
	}
}

// Handler возвращает HTTP-обработчик приложения. Нужен интеграционным тестам,
// которым не требуется поднимать реальный слушающий сокет.
func (a *Application) Handler() http.Handler {
	return a.server.Handler
}

// Close освобождает ресурсы приложения.
//
// Нужен тестам и вызовам, которые собрали приложение, но не запускали сервер:
// без него пул подключений остался бы открытым.
func (a *Application) Close() {
	if a.pool != nil {
		a.pool.Close()
	}
}

// Run запускает сервер и завершает его при отмене контекста.
// Возврат происходит только после остановки сервера.
func (a *Application) Run(ctx context.Context) error {
	defer a.Close()

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
// Пул закрывается после сервера: иначе последние запросы остались бы без базы.
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
