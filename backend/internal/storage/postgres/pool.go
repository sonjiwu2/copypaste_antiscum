// Package postgres содержит адаптеры хранения на PostgreSQL.
//
// Пакет знает о доменных типах и о драйвере, но ничего не знает о HTTP:
// перевод ошибок в статусы остаётся в httpapi.
package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/config"
)

// ErrEmptyDatabaseURL сообщает, что адрес базы не задан.
//
// Тихий запуск без базы недопустим: приложение обязано упасть на старте,
// а не отвечать пользователю ошибками на каждый запрос.
var ErrEmptyDatabaseURL = errors.New("DATABASE_URL обязателен для хранилища PostgreSQL")

// NewPool открывает пул подключений и проверяет базу одним ping.
//
// Пул закрывается здесь же при любой ошибке проверки: иначе неудачный старт
// оставил бы установленные соединения висеть до завершения процесса.
func NewPool(ctx context.Context, settings config.Database) (*pgxpool.Pool, error) {
	if settings.URL == "" {
		return nil, ErrEmptyDatabaseURL
	}

	poolConfig, err := pgxpool.ParseConfig(settings.URL)
	if err != nil {
		// Текст ошибки pgx не содержит пароль, но и сам URL сюда не подставляется.
		return nil, fmt.Errorf("разобрать адрес базы данных: %w", err)
	}

	poolConfig.MaxConns = settings.MaxConns
	poolConfig.MinConns = settings.MinConns
	poolConfig.ConnConfig.ConnectTimeout = settings.ConnectTimeout

	connectCtx, cancel := WithTimeout(ctx, settings.ConnectTimeout)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(connectCtx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("создать пул подключений: %w", err)
	}

	if err := Ping(connectCtx, pool, settings.ConnectTimeout); err != nil {
		pool.Close()

		return nil, err
	}

	return pool, nil
}

// Pinger — минимум, нужный проверке доступности базы.
// Интерфейс объявлен рядом с потребителем, чтобы readiness не тянул весь пул.
type Pinger interface {
	Ping(ctx context.Context) error
}

// Ping проверяет доступность базы в пределах таймаута.
func Ping(ctx context.Context, pinger Pinger, timeout time.Duration) error {
	pingCtx, cancel := WithTimeout(ctx, timeout)
	defer cancel()

	if err := pinger.Ping(pingCtx); err != nil {
		return fmt.Errorf("проверить доступность базы данных: %w", MapError(err))
	}

	return nil
}

// WithTimeout ограничивает операцию таймаутом, но никогда не продлевает
// уже назначенный вызывающим дедлайн: отменённый клиентом запрос не должен
// продолжать занимать соединение с базой.
func WithTimeout(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	if timeout <= 0 {
		return context.WithCancel(ctx)
	}

	deadline, ok := ctx.Deadline()
	if ok && time.Until(deadline) <= timeout {
		return context.WithCancel(ctx)
	}

	return context.WithTimeout(ctx, timeout)
}
