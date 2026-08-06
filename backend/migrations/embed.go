// Package migrations хранит версионированные миграции схемы PostgreSQL.
//
// Файлы встроены в бинарник: production-образ собирается из одной статической
// сборки, поэтому отдельный слой с .sql в него просто не попал бы, и схема
// разъехалась бы с кодом.
package migrations

import "embed"

//go:embed *.sql
var files embed.FS

// Files возвращает файловую систему с миграциями.
func Files() embed.FS {
	return files
}
