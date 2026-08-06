package postgres_test

import (
	"context"
	"errors"
	"io"
	"testing"

	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/storage/postgres"
	"github.com/sonjiwu2/copypaste_antiscum/backend/migrations"
)

// Ожидаемые файлы миграций. Тест защищает от главного риска доставки:
// схема разъезжается с кодом, если миграция не попала в бинарник.
var expectedMigrations = []string{
	"00001_profiles.sql",
	"00002_scenario_versions.sql",
	"00003_attempts.sql",
	"00004_attempt_decisions.sql",
}

func TestEmbeddedMigrationsArePresent(t *testing.T) {
	entries, err := migrations.Files().ReadDir(".")
	if err != nil {
		t.Fatalf("не удалось прочитать встроенные миграции: %v", err)
	}

	found := make(map[string]struct{}, len(entries))
	for _, entry := range entries {
		found[entry.Name()] = struct{}{}
	}

	for _, name := range expectedMigrations {
		if _, exists := found[name]; !exists {
			t.Errorf("миграция %q не встроена в бинарник", name)
		}
	}

	if len(found) != len(expectedMigrations) {
		t.Errorf("встроено файлов = %d, ожидалось %d: %v", len(found), len(expectedMigrations), found)
	}
}

// Порядок применения задан именами файлов и обязан быть возрастающим:
// внешние ключи не позволят создать attempts раньше profiles.
func TestEmbeddedMigrationsAreOrdered(t *testing.T) {
	entries, err := migrations.Files().ReadDir(".")
	if err != nil {
		t.Fatalf("не удалось прочитать встроенные миграции: %v", err)
	}

	for i, entry := range entries {
		if entry.Name() != expectedMigrations[i] {
			t.Fatalf("миграция %d = %q, ожидалась %q", i, entry.Name(), expectedMigrations[i])
		}
	}
}

func TestOpenSQLRequiresDatabaseURL(t *testing.T) {
	_, err := postgres.OpenSQL(context.Background(), "")
	if !errors.Is(err, postgres.ErrEmptyDatabaseURL) {
		t.Fatalf("ошибка = %v, ожидалась ErrEmptyDatabaseURL", err)
	}
}

func TestRunMigrationRejectsUnknownCommand(t *testing.T) {
	err := postgres.RunMigration(context.Background(), nil, "нет-такой-команды", io.Discard)
	if !errors.Is(err, postgres.ErrUnknownCommand) {
		t.Fatalf("ошибка = %v, ожидалась ErrUnknownCommand", err)
	}
}
