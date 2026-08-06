// Package scenarioarchive готовит версии сценариев к сохранению в архиве.
//
// Содержимое сценариев остаётся собственностью JSON-файлов: пакет только
// читает их, приводит к каноническому виду и считает хэш. Ни формат файлов,
// ни доменная модель сценария здесь не меняются — архив нужен для того,
// чтобы уже начатая попытка доигрывалась на своей версии после выпуска
// нового контента.
package scenarioarchive

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"sort"

	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/scenario"
)

// filePattern повторяет маску загрузчика сценариев: архив обязан видеть
// ровно те же файлы, что и каталог.
const filePattern = "*.json"

// ErrContentChanged — содержимое версии изменено без поднятия номера версии.
//
// Это ошибка редактирования контента: уже сохранённые попытки ссылаются на
// эту версию и после подмены доигрывались бы по другому тексту.
var ErrContentChanged = errors.New("содержимое версии сценария изменено")

// ErrDuplicateVersion — набор синхронизации содержит один ключ дважды.
var ErrDuplicateVersion = errors.New("версия сценария повторяется в синхронизации")

// ErrMultipleActiveVersions — один сценарий имеет несколько текущих версий.
var ErrMultipleActiveVersions = errors.New("у сценария несколько активных версий")

// ErrUnknownScenarioFile — файл не соответствует ни одному проверенному сценарию.
var ErrUnknownScenarioFile = errors.New("файл сценария отсутствует в проверенном каталоге")

// Version — версия сценария вместе с её содержимым.
type Version struct {
	ScenarioID       scenario.ID
	Version          scenario.Version
	Slug             string
	Role             scenario.Role
	Title            string
	Description      string
	Difficulty       scenario.Difficulty
	EstimatedMinutes int
	IsActive         bool

	// Content — канонический JSON сценария, Hash — его отпечаток.
	Content []byte
	Hash    string
}

// Repository хранит версии сценариев.
//
// Интерфейс объявлен рядом с потребителем: реализация живёт в адаптере
// хранения и подменяется в тестах.
type Repository interface {
	Sync(ctx context.Context, versions []Version) error
}

// identity — минимум, по которому файл сопоставляется с проверенным сценарием.
type identity struct {
	ID      string `json:"id"`
	Version int    `json:"version"`
}

// Load читает файлы сценариев и собирает версии для архива.
//
// Метаданные берутся из уже проверенного каталога, а не из файла повторно:
// в архив не должен попасть сценарий, который не прошёл валидацию графа.
func Load(fsys fs.FS, catalog []scenario.Scenario) ([]Version, error) {
	names, err := fs.Glob(fsys, filePattern)
	if err != nil {
		return nil, fmt.Errorf("найти файлы сценариев: %w", err)
	}

	// Порядок обхода файловой системы не гарантирован, а архив должен
	// собираться одинаково при каждом запуске.
	sort.Strings(names)

	known := make(map[identity]scenario.Scenario, len(catalog))
	for _, found := range catalog {
		known[identity{ID: string(found.ID), Version: int(found.Version)}] = found
	}

	versions := make([]Version, 0, len(names))

	for _, name := range names {
		version, err := loadFile(fsys, name, known)
		if err != nil {
			return nil, err
		}

		versions = append(versions, version)
	}

	return versions, nil
}

func loadFile(fsys fs.FS, name string, known map[identity]scenario.Scenario) (Version, error) {
	raw, err := fs.ReadFile(fsys, name)
	if err != nil {
		return Version{}, fmt.Errorf("прочитать файл сценария %q: %w", name, err)
	}

	var key identity
	if err := json.Unmarshal(raw, &key); err != nil {
		return Version{}, fmt.Errorf("определить версию файла сценария %q: %w", name, err)
	}

	definition, found := known[key]
	if !found {
		return Version{}, fmt.Errorf("%w: %q версии %d в файле %q",
			ErrUnknownScenarioFile, key.ID, key.Version, name)
	}

	content, hash, err := canonicalize(raw)
	if err != nil {
		return Version{}, fmt.Errorf("подготовить содержимое сценария %q: %w", name, err)
	}

	return Version{
		ScenarioID:       definition.ID,
		Version:          definition.Version,
		Slug:             definition.Slug,
		Role:             definition.Role,
		Title:            definition.Title,
		Description:      definition.Description,
		Difficulty:       definition.Difficulty,
		EstimatedMinutes: definition.EstimatedMinutes,
		IsActive:         definition.IsActive,
		Content:          content,
		Hash:             hash,
	}, nil
}

// canonicalize приводит JSON к устойчивому виду и считает отпечаток.
//
// Разбор и повторная сборка убирают форматирование и порядок ключей:
// автор сценария может переформатировать файл, не ломая запуск, а любое
// смысловое изменение отпечаток меняет. Порядок элементов массивов
// сохраняется — для узлов сценария он значим.
func canonicalize(raw []byte) ([]byte, string, error) {
	var decoded map[string]any
	if err := json.Unmarshal(raw, &decoded); err != nil {
		return nil, "", fmt.Errorf("разобрать JSON: %w", err)
	}

	if decoded == nil {
		return nil, "", errors.New("сценарий должен быть JSON-объектом")
	}

	// json.Marshal сортирует ключи объектов, поэтому результат не зависит
	// от порядка обхода карт в Go.
	canonical, err := json.Marshal(decoded)
	if err != nil {
		return nil, "", fmt.Errorf("собрать канонический JSON: %w", err)
	}

	// isActive — изменяемая политика каталога, а не часть выпущенного графа.
	// Её смена не требует нового version и хранится отдельной колонкой.
	delete(decoded, "isActive")

	immutable, err := json.Marshal(decoded)
	if err != nil {
		return nil, "", fmt.Errorf("собрать содержимое для отпечатка: %w", err)
	}

	digest := sha256.Sum256(immutable)

	return canonical, hex.EncodeToString(digest[:]), nil
}

// Hash вычисляет тот же детерминированный отпечаток, который используется
// при синхронизации. Нужен адаптеру для проверки сохранённого JSON при чтении.
func Hash(raw []byte) (string, error) {
	_, hash, err := canonicalize(raw)
	if err != nil {
		return "", err
	}

	return hash, nil
}

// LegacyHash вычисляет отпечаток ранней реализации Phase 6, где isActive
// ошибочно входил в неизменяемое содержимое. Он используется только для
// безопасного обновления уже созданных локальных строк на новую семантику.
func LegacyHash(raw []byte) (string, error) {
	var decoded map[string]any
	if err := json.Unmarshal(raw, &decoded); err != nil {
		return "", fmt.Errorf("разобрать JSON: %w", err)
	}

	if decoded == nil {
		return "", errors.New("сценарий должен быть JSON-объектом")
	}

	canonical, err := json.Marshal(decoded)
	if err != nil {
		return "", fmt.Errorf("собрать канонический JSON: %w", err)
	}

	digest := sha256.Sum256(canonical)

	return hex.EncodeToString(digest[:]), nil
}
