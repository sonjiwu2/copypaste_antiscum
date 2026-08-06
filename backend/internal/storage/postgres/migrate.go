package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"

	"github.com/sonjiwu2/copypaste_antiscum/backend/migrations"
)

// Команды миграций, доступные из CLI.
const (
	CommandUp      = "up"
	CommandDown    = "down"
	CommandStatus  = "status"
	CommandVersion = "version"
)

// migrationsDir — корень встроенной файловой системы миграций.
const migrationsDir = "."

// ErrUnknownCommand сообщает о неизвестной команде миграции.
var ErrUnknownCommand = errors.New("неизвестная команда миграции")

// gooseMu защищает глобальные настройки goose: диалект, файловую систему и
// логгер задаются на уровне пакета, поэтому параллельные вызовы недопустимы.
var gooseMu sync.Mutex

// OpenSQL открывает соединение database/sql поверх драйвера pgx.
//
// goose работает через database/sql, поэтому пул pgxpool здесь не подходит.
// Соединение живёт только на время миграции и закрывается вызывающим.
func OpenSQL(ctx context.Context, databaseURL string) (*sql.DB, error) {
	if databaseURL == "" {
		return nil, ErrEmptyDatabaseURL
	}

	connConfig, err := pgx.ParseConfig(databaseURL)
	if err != nil {
		// Адрес не подставляется в текст: он содержит пароль.
		return nil, fmt.Errorf("разобрать адрес базы данных: %w", err)
	}

	db := stdlib.OpenDB(*connConfig)

	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()

		return nil, fmt.Errorf("проверить доступность базы данных: %w", MapError(err))
	}

	return db, nil
}

// RunMigration выполняет команду миграции на уже открытом соединении.
//
// Функция вынесена из CLI, чтобы интеграционный тест применял ровно те же
// миграции тем же способом, что и запуск в Compose.
func RunMigration(ctx context.Context, db *sql.DB, command string, output io.Writer) error {
	gooseMu.Lock()
	defer gooseMu.Unlock()

	if err := prepareGoose(output); err != nil {
		return err
	}

	switch command {
	case CommandUp:
		if err := goose.UpContext(ctx, db, migrationsDir); err != nil {
			return fmt.Errorf("применить миграции: %w", err)
		}
	case CommandDown:
		if err := goose.DownContext(ctx, db, migrationsDir); err != nil {
			return fmt.Errorf("откатить миграцию: %w", err)
		}
	case CommandStatus:
		if err := goose.StatusContext(ctx, db, migrationsDir); err != nil {
			return fmt.Errorf("получить статус миграций: %w", err)
		}
	case CommandVersion:
		version, err := goose.GetDBVersionContext(ctx, db)
		if err != nil {
			return fmt.Errorf("получить версию схемы: %w", err)
		}

		if _, err := fmt.Fprintf(output, "версия схемы: %d\n", version); err != nil {
			return fmt.Errorf("записать версию схемы: %w", err)
		}
	default:
		return fmt.Errorf("%w: %q", ErrUnknownCommand, command)
	}

	return nil
}

// SchemaVersion возвращает номер последней применённой миграции.
func SchemaVersion(ctx context.Context, db *sql.DB) (int64, error) {
	gooseMu.Lock()
	defer gooseMu.Unlock()

	if err := prepareGoose(io.Discard); err != nil {
		return 0, err
	}

	version, err := goose.GetDBVersionContext(ctx, db)
	if err != nil {
		return 0, fmt.Errorf("получить версию схемы: %w", err)
	}

	return version, nil
}

// prepareGoose настраивает goose на встроенные миграции и диалект PostgreSQL.
func prepareGoose(output io.Writer) error {
	goose.SetBaseFS(migrations.Files())
	goose.SetLogger(migrationLogger{output: output})

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("выбрать диалект миграций: %w", err)
	}

	return nil
}

// migrationLogger направляет вывод goose в переданный writer.
//
// Fatalf намеренно не завершает процесс: неудачная миграция обязана вернуться
// ошибкой вызывающему, который сам решит, останавливать ли запуск.
type migrationLogger struct {
	output io.Writer
}

func (l migrationLogger) Printf(format string, v ...any) {
	l.write(format, v...)
}

func (l migrationLogger) Fatalf(format string, v ...any) {
	l.write(format, v...)
}

// write повторяет поведение log.Logger: goose передаёт строки без перевода
// строки и рассчитывает, что его добавит логгер.
func (l migrationLogger) write(format string, v ...any) {
	if !strings.HasSuffix(format, "\n") {
		format += "\n"
	}

	_, _ = fmt.Fprintf(l.output, format, v...)
}
