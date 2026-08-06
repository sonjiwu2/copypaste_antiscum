package memory_test

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/scenario"
	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/storage/memory"
	"github.com/sonjiwu2/copypaste_antiscum/backend/scenarios"
)

func loadCatalog(t *testing.T) []scenario.Scenario {
	t.Helper()

	catalog, err := scenario.LoadFS(scenarios.Files())
	if err != nil {
		t.Fatalf("не удалось загрузить сценарии: %v", err)
	}

	return catalog
}

func newRepository(t *testing.T) *memory.ScenarioRepository {
	t.Helper()

	repository, err := memory.NewScenarioRepository(loadCatalog(t))
	if err != nil {
		t.Fatalf("не удалось собрать каталог: %v", err)
	}

	return repository
}

func TestNewScenarioRepositoryRejectsDuplicateIDs(t *testing.T) {
	catalog := loadCatalog(t)

	_, err := memory.NewScenarioRepository(append(catalog, catalog[0]))
	if err == nil {
		t.Fatal("повторяющийся идентификатор сценария должен отклоняться")
	}
}

func TestScenarioRepositoryList(t *testing.T) {
	testCases := []struct {
		name   string
		filter scenario.Filter
		want   int
	}{
		{name: "без фильтра", filter: scenario.Filter{}, want: 6},
		{name: "только активные", filter: scenario.Filter{OnlyActive: true}, want: 6},
		{name: "роль покупателя", filter: scenario.Filter{Role: scenario.RoleBuyer}, want: 3},
		{name: "роль продавца", filter: scenario.Filter{Role: scenario.RoleSeller}, want: 3},
		{name: "неизвестная роль", filter: scenario.Filter{Role: "courier"}, want: 0},
	}

	repository := newRepository(t)

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			found, err := repository.List(context.Background(), testCase.filter)
			if err != nil {
				t.Fatalf("неожиданная ошибка: %v", err)
			}

			if len(found) != testCase.want {
				t.Errorf("найдено %d сценариев, ожидалось %d", len(found), testCase.want)
			}
		})
	}
}

func TestScenarioRepositoryGet(t *testing.T) {
	repository := newRepository(t)

	found, err := repository.Get(context.Background(), "buyer-fake-delivery")
	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}

	if found.Role != scenario.RoleBuyer {
		t.Errorf("роль = %q, ожидалась buyer", found.Role)
	}

	if _, err := repository.Get(context.Background(), "unknown"); !errors.Is(err, scenario.ErrNotFound) {
		t.Errorf("ошибка = %v, ожидалась ErrNotFound", err)
	}
}

func TestScenarioRepositoryKeepsExactHistoricalVersions(t *testing.T) {
	version1 := repositoryScenario(t, "versioned", 1, false)
	version2 := repositoryScenario(t, "versioned", 2, true)

	repository, err := memory.NewScenarioRepository([]scenario.Scenario{version2, version1})
	if err != nil {
		t.Fatalf("не удалось собрать каталог: %v", err)
	}

	current, err := repository.Get(context.Background(), version1.ID)
	if err != nil {
		t.Fatalf("не удалось получить текущую версию: %v", err)
	}

	if current.Version != 2 {
		t.Errorf("текущая версия = %d, ожидалась 2", current.Version)
	}

	historical, err := repository.GetVersion(context.Background(), version1.ID, 1)
	if err != nil {
		t.Fatalf("не удалось получить историческую версию: %v", err)
	}

	if historical.Version != 1 || historical.IsActive {
		t.Errorf("историческая версия = %d/active=%v, ожидалась 1/false",
			historical.Version, historical.IsActive)
	}

	if _, err := repository.GetVersion(context.Background(), version1.ID, 3); !errors.Is(err, scenario.ErrNotFound) {
		t.Errorf("ошибка = %v, ожидалась ErrNotFound", err)
	}
}

func TestScenarioRepositoryRejectsMultipleActiveVersions(t *testing.T) {
	_, err := memory.NewScenarioRepository([]scenario.Scenario{
		repositoryScenario(t, "versioned", 1, true),
		repositoryScenario(t, "versioned", 2, true),
	})
	if err == nil {
		t.Fatal("несколько активных версий должны отклоняться")
	}
}

func TestScenarioRepositoryListIsDeterministic(t *testing.T) {
	repository, err := memory.NewScenarioRepository([]scenario.Scenario{
		repositoryScenario(t, "z-scenario", 1, true),
		repositoryScenario(t, "a-scenario", 1, true),
	})
	if err != nil {
		t.Fatalf("не удалось собрать каталог: %v", err)
	}

	found, err := repository.List(context.Background(), scenario.Filter{})
	if err != nil {
		t.Fatalf("не удалось получить список: %v", err)
	}

	if len(found) != 2 || found[0].ID != "a-scenario" || found[1].ID != "z-scenario" {
		t.Errorf("порядок = %v, ожидался a-scenario, z-scenario", []scenario.ID{found[0].ID, found[1].ID})
	}
}

func TestScenarioRepositoryRespectsCanceledContext(t *testing.T) {
	repository := newRepository(t)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if _, err := repository.List(ctx, scenario.Filter{}); !errors.Is(err, context.Canceled) {
		t.Errorf("List: ошибка = %v, ожидалась context.Canceled", err)
	}

	if _, err := repository.Get(ctx, "buyer-fake-delivery"); !errors.Is(err, context.Canceled) {
		t.Errorf("Get: ошибка = %v, ожидалась context.Canceled", err)
	}

	if _, err := repository.GetVersion(ctx, "buyer-fake-delivery", 1); !errors.Is(err, context.Canceled) {
		t.Errorf("GetVersion: ошибка = %v, ожидалась context.Canceled", err)
	}
}

func repositoryScenario(t *testing.T, id scenario.ID, version scenario.Version, active bool) scenario.Scenario {
	t.Helper()

	built, err := scenario.New(scenario.Draft{
		ID: id, Version: version, Slug: string(id), Role: scenario.RoleBuyer,
		Title: "Версионный сценарий", Difficulty: scenario.DifficultyEasy,
		EstimatedMinutes: 1, StartNodeID: "decision", IsActive: active,
		Nodes: []scenario.Node{
			{
				ID: "decision", Type: scenario.NodeTypeDecision,
				Choices: []scenario.Choice{
					{ID: "safe", Label: "Безопасно", NextNodeID: "safe", Criticality: scenario.CriticalityLow,
						Consequence: scenario.Consequence{Severity: scenario.SeveritySafe, Title: "Верно", Explanation: "Безопасно"}},
					{ID: "unsafe", Label: "Опасно", NextNodeID: "unsafe", Criticality: scenario.CriticalityHigh,
						Consequence: scenario.Consequence{Severity: scenario.SeverityDangerous, Title: "Опасно", Explanation: "Риск"}},
				},
			},
			{ID: "safe", Type: scenario.NodeTypeTerminal,
				TerminalOutcome: &scenario.Outcome{Type: scenario.OutcomeSafe, Title: "Безопасно", Explanation: "Конец"}},
			{ID: "unsafe", Type: scenario.NodeTypeTerminal,
				TerminalOutcome: &scenario.Outcome{Type: scenario.OutcomeUnsafe, Title: "Опасно", Explanation: "Конец"}},
		},
	})
	if err != nil {
		t.Fatalf("не удалось собрать сценарий: %v", err)
	}

	return built
}

// Изменение полученного сценария не должно доходить до хранилища.
func TestScenarioRepositoryDoesNotLeakMutableState(t *testing.T) {
	repository := newRepository(t)

	first, err := repository.Get(context.Background(), "buyer-fake-delivery")
	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}

	node, found := first.Node("channel-decision")
	if !found {
		t.Fatal("узел решения должен существовать")
	}

	originalLabel := node.Choices[0].Label
	node.Choices[0].Label = "подменённая подпись"
	node.Choices[0].RiskTags = append(node.Choices[0].RiskTags, "подменённый тег")

	second, err := repository.Get(context.Background(), "buyer-fake-delivery")
	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}

	fresh, _ := second.Node("channel-decision")

	if fresh.Choices[0].Label != originalLabel {
		t.Errorf("подпись в хранилище = %q, ожидалась %q", fresh.Choices[0].Label, originalLabel)
	}
}

func TestScenarioRepositoryHandlesConcurrentReads(t *testing.T) {
	const readers = 50

	repository := newRepository(t)

	var waitGroup sync.WaitGroup

	for range readers {
		waitGroup.Add(1)

		go func() {
			defer waitGroup.Done()

			if _, err := repository.List(context.Background(), scenario.Filter{OnlyActive: true}); err != nil {
				t.Errorf("List вернул ошибку: %v", err)

				return
			}

			found, err := repository.Get(context.Background(), "seller-payment-already-sent")
			if err != nil {
				t.Errorf("Get вернул ошибку: %v", err)

				return
			}

			if _, exists := found.Node(found.StartNodeID); !exists {
				t.Error("стартовый узел должен быть доступен")
			}
		}()
	}

	waitGroup.Wait()
}
