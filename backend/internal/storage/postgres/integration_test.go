//go:build integration

// Тесты этого файла требуют настоящий PostgreSQL и не собираются обычным
// `go test ./...`. Команда запуска описана в backend/README.md.
package postgres_test

import (
	"context"
	"database/sql"
	"io"
	"os"
	"testing"
	"time"

	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/storage/postgres"
)

// testDatabaseURLEnv указывает на отдельную тестовую базу.
// Схема в ней пересоздаётся, поэтому подставлять сюда рабочую базу нельзя.
const testDatabaseURLEnv = "TEST_DATABASE_URL"

// requireDatabase открывает соединение с тестовой базой и очищает схему.
//
// Тест падает, а не пропускается, если переменная задана, но база недоступна:
// молчаливый пропуск скрыл бы неработающую интеграцию в CI.
func requireDatabase(t *testing.T) *sql.DB {
	t.Helper()

	url, ok := os.LookupEnv(testDatabaseURLEnv)
	if !ok || url == "" {
		t.Skipf("%s не задан: интеграционные тесты пропущены", testDatabaseURLEnv)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db, err := postgres.OpenSQL(ctx, url)
	if err != nil {
		t.Fatalf("не удалось подключиться к тестовой базе: %v", err)
	}

	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Errorf("не удалось закрыть соединение: %v", err)
		}
	})

	resetSchema(t, db)

	return db
}

func resetSchema(t *testing.T, db *sql.DB) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	statements := []string{
		"DROP SCHEMA IF EXISTS public CASCADE",
		"CREATE SCHEMA public",
	}

	for _, statement := range statements {
		if _, err := db.ExecContext(ctx, statement); err != nil {
			t.Fatalf("не удалось очистить схему (%s): %v", statement, err)
		}
	}
}

func migrateUp(t *testing.T, db *sql.DB) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	if err := postgres.RunMigration(ctx, db, postgres.CommandUp, io.Discard); err != nil {
		t.Fatalf("не удалось применить миграции: %v", err)
	}
}

func TestMigrationsApplyFromEmptyDatabase(t *testing.T) {
	db := requireDatabase(t)
	migrateUp(t, db)

	for _, table := range []string{"profiles", "scenario_versions", "attempts", "attempt_decisions"} {
		if !tableExists(t, db, table) {
			t.Errorf("таблица %q не создана", table)
		}
	}

	version, err := postgres.SchemaVersion(context.Background(), db)
	if err != nil {
		t.Fatalf("не удалось получить версию схемы: %v", err)
	}

	if version != 4 {
		t.Errorf("версия схемы = %d, ожидалась 4", version)
	}
}

// Повторный запуск миграций обязан быть безопасным: контейнер migrate
// поднимается при каждом старте стека.
func TestMigrationsAreIdempotent(t *testing.T) {
	db := requireDatabase(t)

	migrateUp(t, db)
	migrateUp(t, db)

	version, err := postgres.SchemaVersion(context.Background(), db)
	if err != nil {
		t.Fatalf("не удалось получить версию схемы: %v", err)
	}

	if version != 4 {
		t.Errorf("версия схемы = %d, ожидалась 4", version)
	}
}

func TestMigrationsRollBackAndReapply(t *testing.T) {
	db := requireDatabase(t)
	migrateUp(t, db)

	ctx := context.Background()

	if err := postgres.RunMigration(ctx, db, postgres.CommandDown, io.Discard); err != nil {
		t.Fatalf("не удалось откатить миграцию: %v", err)
	}

	if tableExists(t, db, "attempt_decisions") {
		t.Error("после отката таблица attempt_decisions должна отсутствовать")
	}

	migrateUp(t, db)

	if !tableExists(t, db, "attempt_decisions") {
		t.Error("после повторного применения таблица attempt_decisions должна существовать")
	}
}

func TestSchemaCreatesExpectedIndexes(t *testing.T) {
	db := requireDatabase(t)
	migrateUp(t, db)

	expected := []string{
		"attempts_profile_updated_idx",
		"attempts_profile_scenario_completed_idx",
		"attempts_profile_active_idx",
		"attempts_scenario_version_idx",
		"scenario_versions_active_idx",
	}

	for _, index := range expected {
		var found bool

		err := db.QueryRowContext(context.Background(),
			`SELECT EXISTS (SELECT 1 FROM pg_indexes WHERE schemaname = 'public' AND indexname = $1)`,
			index).Scan(&found)
		if err != nil {
			t.Fatalf("не удалось проверить индекс %q: %v", index, err)
		}

		if !found {
			t.Errorf("индекс %q не создан", index)
		}
	}
}

// Ограничения схемы — второй рубеж после доменных проверок.
// Тест доказывает, что они действительно работают, а не только объявлены.
func TestSchemaRejectsInvalidRows(t *testing.T) {
	db := requireDatabase(t)
	migrateUp(t, db)

	ctx := context.Background()
	seedScenarioVersion(t, db)

	testCases := []struct {
		name   string
		insert string
		args   []any
	}{
		{
			name:   "результат вне диапазона 0..100",
			insert: insertAttempt,
			args:   attemptArgs("attempt-score", "in_progress", 101, nil, nil),
		},
		{
			name:   "неизвестный статус",
			insert: insertAttempt,
			args:   attemptArgs("attempt-status", "paused", 100, nil, nil),
		},
		{
			// Завершённая попытка без времени завершения сломала бы прогресс.
			name:   "завершение без времени и итога",
			insert: insertAttempt,
			args:   attemptArgs("attempt-completed", "completed", 100, nil, nil),
		},
		{
			name:   "незавершённая попытка с итогом",
			insert: insertAttempt,
			args:   attemptArgs("attempt-outcome", "in_progress", 100, ptrTime(time.Now()), ptrString("safe")),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			if _, err := db.ExecContext(ctx, testCase.insert, testCase.args...); err == nil {
				t.Fatal("база приняла строку, нарушающую ограничение")
			}
		})
	}
}

// Служебный профиль нужен, пока анонимная идентификация не реализована.
func TestDefaultProfileExists(t *testing.T) {
	db := requireDatabase(t)
	migrateUp(t, db)

	var count int

	err := db.QueryRowContext(context.Background(),
		`SELECT count(*) FROM profiles WHERE id = 'profile_default_demo'`).Scan(&count)
	if err != nil {
		t.Fatalf("не удалось прочитать профили: %v", err)
	}

	if count != 1 {
		t.Errorf("служебных профилей = %d, ожидался 1", count)
	}
}

const insertAttempt = `
INSERT INTO attempts (
    id, profile_id, scenario_id, scenario_version, current_node_id, status, score,
    outcome, started_at, updated_at, completed_at, revealed_node_ids,
    applied_skill_effects, version
) VALUES ($1, 'profile_default_demo', 'test-scenario', 1, 'start', $2, $3,
    $5, now(), now(), $4, ARRAY['start'], '{}'::jsonb, 1)`

func attemptArgs(id, status string, score int, completedAt *time.Time, outcome *string) []any {
	return []any{id, status, score, completedAt, outcome}
}

func ptrTime(moment time.Time) *time.Time { return &moment }

func ptrString(value string) *string { return &value }

func seedScenarioVersion(t *testing.T, db *sql.DB) {
	t.Helper()

	_, err := db.ExecContext(context.Background(), `
INSERT INTO scenario_versions (
    scenario_id, version, slug, role, title, description, difficulty,
    estimated_minutes, is_active, content, content_hash, created_at
) VALUES ('test-scenario', 1, 'test-scenario', 'buyer', 'Тест', '', 'easy',
    4, true, '{}'::jsonb, 'hash', now())`)
	if err != nil {
		t.Fatalf("не удалось подготовить версию сценария: %v", err)
	}
}

func tableExists(t *testing.T, db *sql.DB, table string) bool {
	t.Helper()

	var found bool

	err := db.QueryRowContext(context.Background(),
		`SELECT EXISTS (
             SELECT 1 FROM information_schema.tables
              WHERE table_schema = 'public' AND table_name = $1
         )`, table).Scan(&found)
	if err != nil {
		t.Fatalf("не удалось проверить таблицу %q: %v", table, err)
	}

	return found
}
