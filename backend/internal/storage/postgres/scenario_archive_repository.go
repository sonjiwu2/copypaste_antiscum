package postgres

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/scenario"
	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/scenarioarchive"
)

// ScenarioArchiveRepository хранит версии сценариев в PostgreSQL.
type ScenarioArchiveRepository struct {
	pool         *pgxpool.Pool
	queryTimeout time.Duration
}

// NewScenarioArchiveRepository создаёт архив версий сценариев.
func NewScenarioArchiveRepository(pool *pgxpool.Pool, queryTimeout time.Duration) *ScenarioArchiveRepository {
	return &ScenarioArchiveRepository{pool: pool, queryTimeout: queryTimeout}
}

// insertScenarioVersionQuery добавляет версию, если её ещё нет.
//
// Существующая строка не обновляется: содержимое выпущенной версии
// неизменяемо, иначе уже начатые попытки увидели бы другой текст.
const insertScenarioVersionQuery = `
INSERT INTO scenario_versions (
    scenario_id, version, slug, role, title, description, difficulty,
    estimated_minutes, is_active, content, content_hash, created_at
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, now())
ON CONFLICT (scenario_id, version) DO NOTHING`

const selectScenarioVersionHashQuery = `
SELECT content_hash, content FROM scenario_versions WHERE scenario_id = $1 AND version = $2`

const updateScenarioVersionHashQuery = `
UPDATE scenario_versions SET content_hash = $3 WHERE scenario_id = $1 AND version = $2`

const selectScenarioVersionQuery = `
SELECT content, content_hash, is_active
  FROM scenario_versions
 WHERE scenario_id = $1 AND version = $2`

// deactivateScenarioVersionsQuery сначала скрывает прежний каталог целиком.
// Транзакция затем включает только активные embedded-версии; при любой
// ошибке откат сохраняет прежнюю согласованную политику каталога.
const deactivateScenarioVersionsQuery = `
UPDATE scenario_versions SET is_active = false WHERE is_active`

// updateScenarioVersionFlagQuery обновляет только признак доступности.
//
// Отключение сценария — решение каталога, а не изменение содержимого,
// поэтому этот флаг единственный, который версии разрешено менять.
const updateScenarioVersionFlagQuery = `
UPDATE scenario_versions
   SET is_active = $3
 WHERE scenario_id = $1 AND version = $2 AND is_active <> $3`

// Sync переносит проверенные версии сценариев в архив.
//
// Новая версия добавляется, уже известная сверяется по отпечатку. Расхождение
// отпечатка означает, что содержимое изменили без поднятия версии, и запуск
// обязан остановиться: молча подменять текст начатых попыток нельзя.
func (r *ScenarioArchiveRepository) Sync(ctx context.Context, versions []scenarioarchive.Version) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	if err := validateScenarioVersions(versions); err != nil {
		return err
	}

	queryCtx, cancel := WithTimeout(ctx, r.queryTimeout)
	defer cancel()

	tx, err := r.pool.BeginTx(queryCtx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return fmt.Errorf("начать транзакцию архива сценариев: %w", MapError(err))
	}

	ordered := append([]scenarioarchive.Version(nil), versions...)
	sort.Slice(ordered, func(i, j int) bool {
		if ordered[i].ScenarioID == ordered[j].ScenarioID {
			return ordered[i].Version < ordered[j].Version
		}

		return ordered[i].ScenarioID < ordered[j].ScenarioID
	})

	if _, err := tx.Exec(queryCtx, deactivateScenarioVersionsQuery); err != nil {
		_ = tx.Rollback(queryCtx)

		return fmt.Errorf("отключить прежние версии сценариев: %w", MapError(err))
	}

	for _, version := range ordered {
		if err := r.syncVersion(queryCtx, tx, version); err != nil {
			_ = tx.Rollback(queryCtx)

			return err
		}
	}

	if err := tx.Commit(queryCtx); err != nil {
		return fmt.Errorf("зафиксировать архив сценариев: %w", MapError(err))
	}

	return nil
}

func validateScenarioVersions(versions []scenarioarchive.Version) error {
	type key struct {
		id      scenario.ID
		version scenario.Version
	}

	seen := make(map[key]struct{}, len(versions))
	active := make(map[scenario.ID]scenario.Version, len(versions))

	for _, version := range versions {
		identity := key{id: version.ScenarioID, version: version.Version}
		if _, exists := seen[identity]; exists {
			return fmt.Errorf("%w: сценарий %q версии %d",
				scenarioarchive.ErrDuplicateVersion, version.ScenarioID, version.Version)
		}

		seen[identity] = struct{}{}

		if !version.IsActive {
			continue
		}

		if previous, exists := active[version.ScenarioID]; exists {
			return fmt.Errorf("%w: сценарий %q, версии %d и %d",
				scenarioarchive.ErrMultipleActiveVersions,
				version.ScenarioID, previous, version.Version)
		}

		active[version.ScenarioID] = version.Version
	}

	return nil
}

// GetVersion возвращает точную сохранённую версию, даже если она больше не
// показывается в текущем каталоге.
func (r *ScenarioArchiveRepository) GetVersion(
	ctx context.Context,
	id scenario.ID,
	version scenario.Version,
) (scenario.Scenario, error) {
	if err := ctx.Err(); err != nil {
		return scenario.Scenario{}, err
	}

	queryCtx, cancel := WithTimeout(ctx, r.queryTimeout)
	defer cancel()

	var (
		content    []byte
		storedHash string
		isActive   bool
	)

	err := r.pool.QueryRow(queryCtx, selectScenarioVersionQuery,
		string(id), int32(version)).Scan(&content, &storedHash, &isActive)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return scenario.Scenario{}, scenario.ErrNotFound
		}

		return scenario.Scenario{}, fmt.Errorf("прочитать версию сценария %q %d: %w",
			id, version, MapError(err))
	}

	actualHash, err := scenarioarchive.Hash(content)
	if err != nil {
		return scenario.Scenario{}, fmt.Errorf("%w: сценарий %q версии %d: %v",
			scenario.ErrVersionCorrupt, id, version, err)
	}

	if actualHash != storedHash {
		return scenario.Scenario{}, fmt.Errorf(
			"%w: сценарий %q версии %d; сохранённый хэш %q, вычисленный %q",
			scenario.ErrVersionCorrupt, id, version, storedHash, actualHash)
	}

	definition, err := scenario.DecodeJSON(content)
	if err != nil {
		return scenario.Scenario{}, fmt.Errorf("%w: сценарий %q версии %d: %v",
			scenario.ErrVersionCorrupt, id, version, err)
	}

	if definition.ID != id || definition.Version != version {
		return scenario.Scenario{}, fmt.Errorf(
			"%w: запрошен сценарий %q версии %d, содержимое описывает %q версии %d",
			scenario.ErrVersionCorrupt, id, version,
			definition.ID, definition.Version)
	}

	// Признак активности является отдельной изменяемой политикой каталога и
	// поэтому берётся из колонки, а не из неизменяемого JSON первой вставки.
	definition.IsActive = isActive

	return definition, nil
}

func (r *ScenarioArchiveRepository) syncVersion(
	ctx context.Context,
	tx pgx.Tx,
	version scenarioarchive.Version,
) error {
	_, err := tx.Exec(ctx, insertScenarioVersionQuery,
		string(version.ScenarioID), int32(version.Version), version.Slug, string(version.Role),
		version.Title, version.Description, string(version.Difficulty), int32(version.EstimatedMinutes),
		version.IsActive, version.Content, version.Hash)
	if err != nil {
		return fmt.Errorf("сохранить версию сценария %q: %w", version.ScenarioID, MapError(err))
	}

	var (
		storedHash    string
		storedContent []byte
	)

	err = tx.QueryRow(ctx, selectScenarioVersionHashQuery,
		string(version.ScenarioID), int32(version.Version)).Scan(&storedHash, &storedContent)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("версия сценария %q %d не сохранена", version.ScenarioID, version.Version)
		}

		return fmt.Errorf("прочитать отпечаток версии сценария %q: %w", version.ScenarioID, MapError(err))
	}

	if storedHash != version.Hash {
		upgraded, err := r.upgradeLegacyHash(ctx, tx, version, storedContent, storedHash)
		if err != nil {
			return err
		}

		if upgraded {
			storedHash = version.Hash
		}
	}

	if storedHash != version.Hash {
		return fmt.Errorf(
			"%w: сценарий %q версии %d; сохранённый хэш %q, входящий %q; поднимите номер версии",
			scenarioarchive.ErrContentChanged, version.ScenarioID, version.Version,
			storedHash, version.Hash)
	}

	_, err = tx.Exec(ctx, updateScenarioVersionFlagQuery,
		string(version.ScenarioID), int32(version.Version), version.IsActive)
	if err != nil {
		return fmt.Errorf("обновить доступность версии сценария %q: %w", version.ScenarioID, MapError(err))
	}

	return nil
}

// upgradeLegacyHash меняет только алгоритм отпечатка ранней реализации.
// Содержимое строки не обновляется, а совпадение нового immutable-хэша
// доказывает, что бизнес-контент остался тем же.
func (r *ScenarioArchiveRepository) upgradeLegacyHash(
	ctx context.Context,
	tx pgx.Tx,
	version scenarioarchive.Version,
	storedContent []byte,
	storedHash string,
) (bool, error) {
	legacyHash, err := scenarioarchive.LegacyHash(storedContent)
	if err != nil {
		return false, fmt.Errorf("проверить прежний отпечаток сценария %q: %w",
			version.ScenarioID, err)
	}

	immutableHash, err := scenarioarchive.Hash(storedContent)
	if err != nil {
		return false, fmt.Errorf("проверить содержимое сценария %q: %w",
			version.ScenarioID, err)
	}

	if storedHash != legacyHash || immutableHash != version.Hash {
		return false, nil
	}

	_, err = tx.Exec(ctx, updateScenarioVersionHashQuery,
		string(version.ScenarioID), int32(version.Version), version.Hash)
	if err != nil {
		return false, fmt.Errorf("обновить алгоритм отпечатка сценария %q: %w",
			version.ScenarioID, MapError(err))
	}

	return true, nil
}
