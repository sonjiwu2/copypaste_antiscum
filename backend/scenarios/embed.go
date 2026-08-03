// Package scenarios хранит файлы учебных сценариев.
//
// Файлы встроены в бинарник: приложение запускается без внешнего каталога,
// а Backend №3 заменяет содержимое, не трогая код.
package scenarios

import "embed"

//go:embed *.json
var files embed.FS

// Files возвращает файловую систему со сценариями.
func Files() embed.FS {
	return files
}
