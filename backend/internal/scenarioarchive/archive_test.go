package scenarioarchive_test

import (
	"errors"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/scenario"
	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/scenarioarchive"
	"github.com/sonjiwu2/copypaste_antiscum/backend/scenarios"
)

func embeddedCatalog(t *testing.T) []scenario.Scenario {
	t.Helper()

	catalog, err := scenario.LoadFS(scenarios.Files())
	if err != nil {
		t.Fatalf("не удалось загрузить сценарии: %v", err)
	}

	return catalog
}

// Архив обязан покрывать весь каталог: пропущенная версия сделает
// соответствующие попытки невозможными из-за внешнего ключа.
func TestLoadCoversEmbeddedCatalog(t *testing.T) {
	catalog := embeddedCatalog(t)

	versions, err := scenarioarchive.Load(scenarios.Files(), catalog)
	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}

	if len(versions) != len(catalog) {
		t.Fatalf("версий = %d, ожидалось %d", len(versions), len(catalog))
	}

	byID := make(map[scenario.ID]scenarioarchive.Version, len(versions))
	for _, version := range versions {
		byID[version.ScenarioID] = version
	}

	for _, definition := range catalog {
		version, found := byID[definition.ID]
		if !found {
			t.Fatalf("сценарий %q отсутствует в архиве", definition.ID)
		}

		if version.Version != definition.Version {
			t.Errorf("%q: версия = %d, ожидалась %d", definition.ID, version.Version, definition.Version)
		}

		if version.Role != definition.Role || version.Title != definition.Title {
			t.Errorf("%q: метаданные = %q/%q, ожидались %q/%q",
				definition.ID, version.Role, version.Title, definition.Role, definition.Title)
		}

		if version.IsActive != definition.IsActive {
			t.Errorf("%q: доступность = %v, ожидалась %v",
				definition.ID, version.IsActive, definition.IsActive)
		}

		if len(version.Content) == 0 || version.Hash == "" {
			t.Errorf("%q: содержимое или отпечаток пусты", definition.ID)
		}
	}
}

// Отпечаток обязан быть устойчивым: иначе каждый перезапуск выглядел бы
// как изменение содержимого и останавливал приложение.
func TestLoadIsDeterministic(t *testing.T) {
	catalog := embeddedCatalog(t)

	first, err := scenarioarchive.Load(scenarios.Files(), catalog)
	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}

	second, err := scenarioarchive.Load(scenarios.Files(), catalog)
	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}

	for i := range first {
		if first[i].ScenarioID != second[i].ScenarioID {
			t.Fatalf("порядок версий нестабилен: %q и %q", first[i].ScenarioID, second[i].ScenarioID)
		}

		if first[i].Hash != second[i].Hash {
			t.Errorf("%q: отпечаток нестабилен", first[i].ScenarioID)
		}
	}
}

const minimalScenario = `{
  "id": "archive-demo",
  "version": 1,
  "slug": "archive-demo",
  "role": "buyer",
  "title": "Демонстрация архива",
  "description": "",
  "difficulty": "easy",
  "estimatedMinutes": 1,
  "startNodeId": "start",
  "isActive": true,
  "nodes": [
    {
      "id": "start",
      "type": "decision",
      "decisionPrompt": "Что вы сделаете?",
      "choices": [
        {
          "id": "safe",
          "label": "Отказаться",
          "nextNodeId": "safe-ending",
          "safetyScore": 0,
          "criticality": "low",
          "riskTags": ["demo-risk"],
          "skillEffects": [{"skill": "demo-skill", "delta": 1}],
          "consequence": {
            "severity": "safe",
            "title": "Верно",
            "explanation": "Объяснение",
            "realWorldRule": "Правило"
          }
        },
        {
          "id": "unsafe",
          "label": "Согласиться",
          "nextNodeId": "unsafe-ending",
          "safetyScore": -20,
          "criticality": "high",
          "consequence": {
            "severity": "dangerous",
            "title": "Опасно",
            "explanation": "Объяснение",
            "realWorldRule": "Правило"
          }
        }
      ]
    },
    {
      "id": "safe-ending",
      "type": "terminal",
      "outcome": { "type": "safe", "title": "Безопасно", "explanation": "Объяснение" }
    },
    {
      "id": "unsafe-ending",
      "type": "terminal",
      "outcome": { "type": "unsafe", "title": "Опасно", "explanation": "Объяснение" }
    }
  ]
}`

func loadSingle(t *testing.T, raw string) scenarioarchive.Version {
	t.Helper()

	fsys := fstest.MapFS{"demo.json": &fstest.MapFile{Data: []byte(raw)}}

	catalog, err := scenario.LoadFS(fsys)
	if err != nil {
		t.Fatalf("не удалось загрузить сценарий: %v", err)
	}

	versions, err := scenarioarchive.Load(fsys, catalog)
	if err != nil {
		t.Fatalf("не удалось собрать архив: %v", err)
	}

	if len(versions) != 1 {
		t.Fatalf("версий = %d, ожидалась 1", len(versions))
	}

	return versions[0]
}

// Автор сценария вправе переформатировать файл: отступы и порядок ключей
// не являются содержанием и не должны требовать поднятия версии.
func TestHashIgnoresFormatting(t *testing.T) {
	original := loadSingle(t, minimalScenario)

	// Тот же сценарий, но однострочный и с другим порядком верхних ключей.
	reformatted := loadSingle(t, `{"version":1,"id":"archive-demo","slug":"archive-demo",`+
		`"role":"buyer","title":"Демонстрация архива","description":"","difficulty":"easy",`+
		`"estimatedMinutes":1,"startNodeId":"start","isActive":true,"nodes":`+
		mustNodes(t, minimalScenario)+`}`)

	if original.Hash != reformatted.Hash {
		t.Errorf("отпечаток изменился от переформатирования: %q и %q",
			original.Hash, reformatted.Hash)
	}
}

// Смысловая правка обязана менять отпечаток: именно так ловится
// изменение содержимого без поднятия версии.
func TestHashChangesWithContent(t *testing.T) {
	original := loadSingle(t, minimalScenario)

	changed := loadSingle(t, strings.Replace(minimalScenario,
		`"label": "Отказаться"`, `"label": "Вежливо отказаться"`, 1))

	if original.Hash == changed.Hash {
		t.Error("отпечаток не изменился при правке содержимого")
	}
}

func TestHashCoversImmutableScenarioFields(t *testing.T) {
	original := loadSingle(t, minimalScenario)

	testCases := []struct {
		name string
		old  string
		new  string
	}{
		{name: "версия", old: `"version": 1`, new: `"version": 2`},
		{name: "узел", old: `"decisionPrompt": "Что вы сделаете?"`, new: `"decisionPrompt": "Ваш выбор?"`},
		{name: "вариант", old: `"label": "Отказаться"`, new: `"label": "Отклонить"`},
		{name: "score", old: `"safetyScore": 0`, new: `"safetyScore": -1`},
		{name: "последствие", old: `"title": "Верно"`, new: `"title": "Точно"`},
		{name: "risk tag", old: `"demo-risk"`, new: `"changed-risk"`},
		{name: "skill effect", old: `"delta": 1`, new: `"delta": 2`},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			changedRaw := strings.Replace(minimalScenario, testCase.old, testCase.new, 1)
			if changedRaw == minimalScenario {
				t.Fatalf("тестовая замена %q не сработала", testCase.old)
			}

			changed := loadSingle(t, changedRaw)
			if original.Hash == changed.Hash {
				t.Errorf("отпечаток не изменился после правки поля %s", testCase.name)
			}
		})
	}
}

// Доступность — изменяемая политика каталога. Её можно переключать без
// выпуска новой версии, поэтому в immutable-отпечаток она не входит.
func TestHashIgnoresActivationMetadata(t *testing.T) {
	active := loadSingle(t, minimalScenario)
	inactive := loadSingle(t, strings.Replace(minimalScenario,
		`"isActive": true`, `"isActive": false`, 1))

	if active.Hash != inactive.Hash {
		t.Errorf("isActive изменил отпечаток: %q и %q", active.Hash, inactive.Hash)
	}

	if string(active.Content) == string(inactive.Content) {
		t.Error("сохранённый JSON должен сохранять фактическую политику каталога")
	}
}

// Файл, не прошедший проверку графа, не должен попасть в архив.
func TestLoadRejectsUnknownFile(t *testing.T) {
	fsys := fstest.MapFS{"demo.json": &fstest.MapFile{Data: []byte(minimalScenario)}}

	_, err := scenarioarchive.Load(fsys, nil)
	if !errors.Is(err, scenarioarchive.ErrUnknownScenarioFile) {
		t.Fatalf("ошибка = %v, ожидалась ErrUnknownScenarioFile", err)
	}
}

// mustNodes вырезает массив узлов из исходного файла, чтобы переформатированный
// вариант отличался только оформлением верхнего уровня.
func mustNodes(t *testing.T, raw string) string {
	t.Helper()

	const marker = `"nodes":`

	index := strings.Index(raw, marker)
	if index < 0 {
		t.Fatalf("массив узлов не найден")
	}

	return raw[index+len(marker) : len(raw)-1]
}
