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
