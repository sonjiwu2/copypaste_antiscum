//go:build integration

package postgres_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/attempt"
	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/platform/clock"
	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/platform/identifier"
	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/profile"
	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/progress"
	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/scenario"
	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/scenarioarchive"
	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/storage/memory"
	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/storage/postgres"
)

func archiveVersion(hash string) scenarioarchive.Version {
	return scenarioarchive.Version{
		ScenarioID:       "archive-demo",
		Version:          1,
		Slug:             "archive-demo",
		Role:             scenario.RoleBuyer,
		Title:            "Демонстрация архива",
		Description:      "",
		Difficulty:       scenario.DifficultyEasy,
		EstimatedMinutes: 1,
		IsActive:         true,
		Content:          []byte(`{"id":"archive-demo"}`),
		Hash:             hash,
	}
}

func TestScenarioArchiveSyncInsertsAndIsIdempotent(t *testing.T) {
	prepareSchema(t)

	pool := requirePool(t)
	archive := postgres.NewScenarioArchiveRepository(pool, queryTimeout)
	ctx := context.Background()

	versions := []scenarioarchive.Version{archiveVersion("hash-1")}

	if err := archive.Sync(ctx, versions); err != nil {
		t.Fatalf("первая синхронизация вернула ошибку: %v", err)
	}

	// Повторный запуск приложения не должен ничего менять.
	if err := archive.Sync(ctx, versions); err != nil {
		t.Fatalf("повторная синхронизация вернула ошибку: %v", err)
	}

	var count int

	err := pool.QueryRow(ctx,
		"SELECT count(*) FROM scenario_versions WHERE scenario_id = 'archive-demo'").Scan(&count)
	if err != nil {
		t.Fatalf("не удалось прочитать архив: %v", err)
	}

	if count != 1 {
		t.Errorf("версий = %d, ожидалась 1", count)
	}
}

// Ключевая защита долговечности: содержимое выпущенной версии изменять нельзя,
// иначе уже начатые попытки доигрывались бы по другому тексту.
func TestScenarioArchiveRejectsChangedContent(t *testing.T) {
	prepareSchema(t)

	pool := requirePool(t)
	archive := postgres.NewScenarioArchiveRepository(pool, queryTimeout)
	ctx := context.Background()

	if err := archive.Sync(ctx, []scenarioarchive.Version{archiveVersion("hash-1")}); err != nil {
		t.Fatalf("первая синхронизация вернула ошибку: %v", err)
	}

	changed := archiveVersion("hash-2")
	changed.Content = []byte(`{"id":"archive-demo","changed":true}`)

	err := archive.Sync(ctx, []scenarioarchive.Version{changed})
	if !errors.Is(err, scenarioarchive.ErrContentChanged) {
		t.Fatalf("ошибка = %v, ожидалась ErrContentChanged", err)
	}

	// Сохранённое содержимое остаётся прежним.
	var storedHash string

	if err := pool.QueryRow(ctx,
		"SELECT content_hash FROM scenario_versions WHERE scenario_id = 'archive-demo' AND version = 1",
	).Scan(&storedHash); err != nil {
		t.Fatalf("не удалось прочитать отпечаток: %v", err)
	}

	if storedHash != "hash-1" {
		t.Errorf("отпечаток = %q, ожидался hash-1", storedHash)
	}

	var isActive bool
	if err := pool.QueryRow(ctx,
		"SELECT is_active FROM scenario_versions WHERE scenario_id = 'archive-demo' AND version = 1",
	).Scan(&isActive); err != nil {
		t.Fatalf("не удалось прочитать доступность: %v", err)
	}

	if !isActive {
		t.Error("откат синхронизации должен сохранить прежнюю активную версию")
	}
}

// Новая версия добавляется рядом со старой, а не вместо неё.
func TestScenarioArchiveKeepsOldVersions(t *testing.T) {
	prepareSchema(t)

	pool := requirePool(t)
	archive := postgres.NewScenarioArchiveRepository(pool, queryTimeout)
	ctx := context.Background()

	if err := archive.Sync(ctx, []scenarioarchive.Version{archiveVersion("hash-1")}); err != nil {
		t.Fatalf("первая синхронизация вернула ошибку: %v", err)
	}

	next := archiveVersion("hash-2")
	next.Version = 2

	if err := archive.Sync(ctx, []scenarioarchive.Version{next}); err != nil {
		t.Fatalf("синхронизация новой версии вернула ошибку: %v", err)
	}

	rows, err := pool.Query(ctx,
		"SELECT version, is_active FROM scenario_versions WHERE scenario_id = 'archive-demo' ORDER BY version")
	if err != nil {
		t.Fatalf("не удалось прочитать архив: %v", err)
	}

	defer rows.Close()

	type storedVersion struct {
		version  int32
		isActive bool
	}

	var found []storedVersion

	for rows.Next() {
		var version storedVersion
		if err := rows.Scan(&version.version, &version.isActive); err != nil {
			t.Fatalf("не удалось прочитать версию: %v", err)
		}

		found = append(found, version)
	}

	if len(found) != 2 || found[0].version != 1 || found[0].isActive ||
		found[1].version != 2 || !found[1].isActive {
		t.Errorf("версии = %+v, ожидались 1/inactive и 2/active", found)
	}
}

// Отключение сценария — решение каталога, а не изменение содержимого:
// признак доступности обновляется, отпечаток остаётся прежним.
func TestScenarioArchiveUpdatesActivityFlag(t *testing.T) {
	prepareSchema(t)

	pool := requirePool(t)
	archive := postgres.NewScenarioArchiveRepository(pool, queryTimeout)
	ctx := context.Background()

	if err := archive.Sync(ctx, []scenarioarchive.Version{archiveVersion("hash-1")}); err != nil {
		t.Fatalf("первая синхронизация вернула ошибку: %v", err)
	}

	disabled := archiveVersion("hash-1")
	disabled.IsActive = false

	if err := archive.Sync(ctx, []scenarioarchive.Version{disabled}); err != nil {
		t.Fatalf("синхронизация отключённого сценария вернула ошибку: %v", err)
	}

	var isActive bool

	if err := pool.QueryRow(ctx,
		"SELECT is_active FROM scenario_versions WHERE scenario_id = 'archive-demo' AND version = 1",
	).Scan(&isActive); err != nil {
		t.Fatalf("не удалось прочитать доступность: %v", err)
	}

	if isActive {
		t.Error("сценарий должен быть отключён")
	}
}

const validArchivedScenario = `{
  "id": "archive-exact",
  "version": 1,
  "slug": "archive-exact",
  "role": "buyer",
  "title": "Точная версия",
  "description": "Проверка чтения архива",
  "difficulty": "easy",
  "estimatedMinutes": 1,
  "startNodeId": "decision",
  "isActive": true,
  "nodes": [
    {
      "id": "decision",
      "type": "decision",
      "decisionPrompt": "Что сделать?",
      "choices": [
        {
          "id": "safe",
          "label": "Безопасно",
          "nextNodeId": "safe-ending",
          "safetyScore": 0,
          "criticality": "low",
          "consequence": {"severity":"safe","title":"Верно","explanation":"Безопасно"}
        },
        {
          "id": "unsafe",
          "label": "Опасно",
          "nextNodeId": "unsafe-ending",
          "safetyScore": -20,
          "criticality": "high",
          "consequence": {"severity":"dangerous","title":"Опасно","explanation":"Риск"}
        }
      ]
    },
    {"id":"safe-ending","type":"terminal","outcome":{"type":"safe","title":"Безопасно","explanation":"Конец"}},
    {"id":"unsafe-ending","type":"terminal","outcome":{"type":"unsafe","title":"Опасно","explanation":"Конец"}}
  ]
}`

func validArchiveVersion(t *testing.T) scenarioarchive.Version {
	t.Helper()

	definition, err := scenario.DecodeJSON([]byte(validArchivedScenario))
	if err != nil {
		t.Fatalf("не удалось разобрать тестовый сценарий: %v", err)
	}

	hash, err := scenarioarchive.Hash([]byte(validArchivedScenario))
	if err != nil {
		t.Fatalf("не удалось вычислить отпечаток: %v", err)
	}

	return scenarioarchive.Version{
		ScenarioID: definition.ID, Version: definition.Version, Slug: definition.Slug,
		Role: definition.Role, Title: definition.Title, Description: definition.Description,
		Difficulty: definition.Difficulty, EstimatedMinutes: definition.EstimatedMinutes,
		IsActive: definition.IsActive, Content: []byte(validArchivedScenario), Hash: hash,
	}
}

func TestScenarioArchiveGetsExactVersion(t *testing.T) {
	prepareSchema(t)

	pool := requirePool(t)
	archive := postgres.NewScenarioArchiveRepository(pool, queryTimeout)
	ctx := context.Background()
	version := validArchiveVersion(t)

	if err := archive.Sync(ctx, []scenarioarchive.Version{version}); err != nil {
		t.Fatalf("синхронизация вернула ошибку: %v", err)
	}

	found, err := archive.GetVersion(ctx, version.ScenarioID, version.Version)
	if err != nil {
		t.Fatalf("точная версия не прочитана: %v", err)
	}

	if found.ID != version.ScenarioID || found.Version != version.Version || !found.IsActive {
		t.Errorf("получена версия %+v", found)
	}

	node, exists := found.Node("decision")
	if !exists || len(node.Choices) != 2 || node.Choices[1].SafetyScore != -20 {
		t.Errorf("граф восстановлен неточно: %+v", node)
	}

	if _, err := archive.GetVersion(ctx, version.ScenarioID, 2); !errors.Is(err, scenario.ErrNotFound) {
		t.Errorf("отсутствующая версия: ошибка = %v, ожидалась ErrNotFound", err)
	}
}

func TestScenarioArchiveDetectsCorruptStoredContent(t *testing.T) {
	prepareSchema(t)

	pool := requirePool(t)
	archive := postgres.NewScenarioArchiveRepository(pool, queryTimeout)
	ctx := context.Background()
	version := validArchiveVersion(t)

	if err := archive.Sync(ctx, []scenarioarchive.Version{version}); err != nil {
		t.Fatalf("синхронизация вернула ошибку: %v", err)
	}

	if _, err := pool.Exec(ctx, `
UPDATE scenario_versions SET content = '{"id":"corrupt"}'::jsonb
 WHERE scenario_id = $1 AND version = $2`, version.ScenarioID, version.Version); err != nil {
		t.Fatalf("не удалось подготовить повреждённую строку: %v", err)
	}

	_, err := archive.GetVersion(ctx, version.ScenarioID, version.Version)
	if !errors.Is(err, scenario.ErrVersionCorrupt) {
		t.Fatalf("ошибка = %v, ожидалась ErrVersionCorrupt", err)
	}
}

func TestScenarioArchiveUpgradesLegacyActivationHash(t *testing.T) {
	prepareSchema(t)

	pool := requirePool(t)
	archive := postgres.NewScenarioArchiveRepository(pool, queryTimeout)
	ctx := context.Background()
	version := validArchiveVersion(t)

	legacyHash, err := scenarioarchive.LegacyHash(version.Content)
	if err != nil {
		t.Fatalf("не удалось вычислить прежний отпечаток: %v", err)
	}

	_, err = pool.Exec(ctx, `
INSERT INTO scenario_versions (
    scenario_id, version, slug, role, title, description, difficulty,
    estimated_minutes, is_active, content, content_hash, created_at
) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,now())`,
		version.ScenarioID, version.Version, version.Slug, version.Role, version.Title,
		version.Description, version.Difficulty, version.EstimatedMinutes,
		version.IsActive, version.Content, legacyHash)
	if err != nil {
		t.Fatalf("не удалось вставить строку прежнего формата: %v", err)
	}

	if err := archive.Sync(ctx, []scenarioarchive.Version{version}); err != nil {
		t.Fatalf("безопасное обновление отпечатка вернуло ошибку: %v", err)
	}

	var storedHash string
	if err := pool.QueryRow(ctx, `
SELECT content_hash FROM scenario_versions WHERE scenario_id=$1 AND version=$2`,
		version.ScenarioID, version.Version).Scan(&storedHash); err != nil {
		t.Fatalf("не удалось прочитать обновлённый отпечаток: %v", err)
	}

	if storedHash != version.Hash || strings.EqualFold(storedHash, legacyHash) {
		t.Errorf("отпечаток = %q, ожидался новый %q", storedHash, version.Hash)
	}
}

func TestScenarioArchiveRejectsAmbiguousActiveSet(t *testing.T) {
	archive := postgres.NewScenarioArchiveRepository(nil, queryTimeout)
	first := archiveVersion("hash-1")
	second := first
	second.Version = 2

	err := archive.Sync(context.Background(), []scenarioarchive.Version{first, second})
	if !errors.Is(err, scenarioarchive.ErrMultipleActiveVersions) {
		t.Fatalf("ошибка = %v, ожидалась ErrMultipleActiveVersions", err)
	}
}

func transitionArchiveVersion(
	t *testing.T,
	version int,
	active bool,
	penalty int,
) (scenario.Scenario, scenarioarchive.Version) {
	t.Helper()

	raw := fmt.Sprintf(`{
  "id":"postgres-version-transition","version":%d,"slug":"postgres-version-transition",
  "role":"buyer","title":"Версия %d","description":"Проверка PostgreSQL",
  "difficulty":"medium","estimatedMinutes":2,"startNodeId":"first","isActive":%t,
  "nodes":[
    {"id":"first","type":"decision","decisionPrompt":"Первый выбор","choices":[
      {"id":"continue","label":"Версия %d","nextNodeId":"second","safetyScore":%d,
       "criticality":"medium","consequence":{"severity":"warning","title":"Версия %d","explanation":"Первый шаг"}},
      {"id":"stop","label":"Остановиться","nextNodeId":"unsafe","safetyScore":%d,
       "criticality":"high","consequence":{"severity":"dangerous","title":"Стоп","explanation":"Опасно"}}
    ]},
    {"id":"second","type":"decision","decisionPrompt":"Второй выбор","choices":[
      {"id":"finish","label":"Завершить","nextNodeId":"safe","safetyScore":%d,
       "criticality":"medium","consequence":{"severity":"warning","title":"Версия %d","explanation":"Второй шаг"}},
      {"id":"fail","label":"Ошибка","nextNodeId":"unsafe","safetyScore":%d,
       "criticality":"high","consequence":{"severity":"dangerous","title":"Ошибка","explanation":"Опасно"}}
    ]},
    {"id":"safe","type":"terminal","outcome":{"type":"safe","title":"Версия %d","explanation":"Безопасно"}},
    {"id":"unsafe","type":"terminal","outcome":{"type":"unsafe","title":"Версия %d","explanation":"Опасно"}}
  ]
}`, version, version, active, version, penalty, version, penalty,
		penalty, version, penalty, version, version)

	definition, err := scenario.DecodeJSON([]byte(raw))
	if err != nil {
		t.Fatalf("не удалось разобрать сценарий версии %d: %v", version, err)
	}

	hash, err := scenarioarchive.Hash([]byte(raw))
	if err != nil {
		t.Fatalf("не удалось вычислить отпечаток версии %d: %v", version, err)
	}

	archived := scenarioarchive.Version{
		ScenarioID: definition.ID, Version: definition.Version, Slug: definition.Slug,
		Role: definition.Role, Title: definition.Title, Description: definition.Description,
		Difficulty: definition.Difficulty, EstimatedMinutes: definition.EstimatedMinutes,
		IsActive: definition.IsActive, Content: []byte(raw), Hash: hash,
	}

	return definition, archived
}

// Сквозной контракт Phase 6 на настоящем PostgreSQL: старая попытка читает
// и применяет граф версии 1 после активации версии 2, прогресс видит её
// сохранённый результат, а новая попытка стартует уже на версии 2.
func TestPostgresAttemptContinuesAfterScenarioVersionTransition(t *testing.T) {
	prepareSchema(t)

	pool := requirePool(t)
	ctx := context.Background()
	archive := postgres.NewScenarioArchiveRepository(pool, queryTimeout)
	attempts := postgres.NewAttemptRepository(pool, queryTimeout)
	profiles := postgres.NewProfileRepository(pool, queryTimeout)
	owner := profile.ID("version-transition-profile")
	moment := time.Date(2026, 8, 5, 12, 0, 0, 0, time.UTC)

	if err := profiles.Ensure(ctx, profile.Profile{ID: owner, CreatedAt: moment, UpdatedAt: moment}); err != nil {
		t.Fatalf("не удалось создать профиль: %v", err)
	}

	version1Current, version1Archive := transitionArchiveVersion(t, 1, true, -10)
	version1Archive.IsActive = true
	if err := archive.Sync(ctx, []scenarioarchive.Version{version1Archive}); err != nil {
		t.Fatalf("не удалось синхронизировать версию 1: %v", err)
	}

	currentV1, err := memory.NewScenarioRepository([]scenario.Scenario{version1Current})
	if err != nil {
		t.Fatalf("не удалось собрать каталог версии 1: %v", err)
	}

	testClock := &clock.Fixed{Moment: moment, Step: time.Minute}
	testIDs := &identifier.Sequential{Prefix: "postgres-version-attempt"}
	serviceV1 := attempt.NewService(currentV1, archive, attempts, testClock, testIDs)

	started, err := serviceV1.Start(ctx, owner, version1Current.ID)
	if err != nil {
		t.Fatalf("не удалось начать попытку версии 1: %v", err)
	}

	firstCommand := attempt.SubmitChoiceCommand{
		AttemptID: started.ID, ProfileID: owner, NodeID: "first",
		ChoiceID: "continue", IdempotencyKey: "postgres-version-first",
	}
	first, err := serviceV1.SubmitChoice(ctx, firstCommand)
	if err != nil {
		t.Fatalf("не удалось применить первый выбор версии 1: %v", err)
	}

	if first.Score != 90 {
		t.Fatalf("score версии 1 после первого выбора = %d, ожидалось 90", first.Score)
	}

	version2Current, version2Archive := transitionArchiveVersion(t, 2, true, -30)
	if err := archive.Sync(ctx, []scenarioarchive.Version{version2Archive}); err != nil {
		t.Fatalf("не удалось синхронизировать версию 2: %v", err)
	}

	currentV2, err := memory.NewScenarioRepository([]scenario.Scenario{version2Current})
	if err != nil {
		t.Fatalf("не удалось собрать каталог версии 2: %v", err)
	}

	serviceV2 := attempt.NewService(currentV2, archive, attempts, testClock, testIDs)

	replay, err := serviceV2.SubmitChoice(ctx, firstCommand)
	if err != nil {
		t.Fatalf("повтор версии 1 после активации версии 2 вернул ошибку: %v", err)
	}

	if replay.Score != 90 || replay.Consequence.Title != "Версия 1" {
		t.Errorf("повтор = score %d / %q, ожидалась версия 1", replay.Score, replay.Consequence.Title)
	}

	completed, err := serviceV2.SubmitChoice(ctx, attempt.SubmitChoiceCommand{
		AttemptID: started.ID, ProfileID: owner, NodeID: "second",
		ChoiceID: "finish", IdempotencyKey: "postgres-version-second",
	})
	if err != nil {
		t.Fatalf("не удалось завершить попытку версии 1: %v", err)
	}

	if completed.Status != attempt.StatusCompleted || completed.Score != 80 ||
		completed.Consequence.Title != "Версия 1" {
		t.Errorf("финал старой попытки = %+v", completed)
	}

	newAttempt, err := serviceV2.Start(ctx, owner, version2Current.ID)
	if err != nil {
		t.Fatalf("не удалось начать новую попытку: %v", err)
	}

	if newAttempt.Scenario.Version != 2 {
		t.Errorf("новая попытка использует версию %d, ожидалась 2", newAttempt.Scenario.Version)
	}

	progressService := progress.NewService(
		postgres.NewProgressRepository(pool, queryTimeout),
		scenario.NewService(currentV2),
	)
	foundProgress, err := progressService.Of(ctx, owner)
	if err != nil {
		t.Fatalf("не удалось прочитать прогресс: %v", err)
	}

	if foundProgress.Summary.CompletedAttempts != 1 || foundProgress.Summary.LatestScore != 80 {
		t.Errorf("прогресс старой версии = %+v, ожидалась одна попытка со score 80",
			foundProgress.Summary)
	}

	var active []struct {
		version  int32
		isActive bool
	}
	rows, err := pool.Query(ctx, `
SELECT version, is_active FROM scenario_versions
 WHERE scenario_id=$1 ORDER BY version`, version1Current.ID)
	if err != nil {
		t.Fatalf("не удалось прочитать политику версий: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var item struct {
			version  int32
			isActive bool
		}
		if err := rows.Scan(&item.version, &item.isActive); err != nil {
			t.Fatalf("не удалось прочитать версию: %v", err)
		}
		active = append(active, item)
	}

	if len(active) != 2 || active[0].isActive || !active[1].isActive {
		t.Errorf("активность версий = %+v, ожидалась false/true", active)
	}
}
