package attempt_test

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/attempt"
	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/platform/clock"
	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/platform/identifier"
	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/scenario"
	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/storage/memory"
	"github.com/sonjiwu2/copypaste_antiscum/backend/scenarios"
)

// stubCatalog отдаёт заранее подготовленные сценарии.
type stubCatalog struct {
	scenarios map[scenario.ID]scenario.Scenario
	failure   error
}

func (s stubCatalog) Get(_ context.Context, id scenario.ID) (scenario.Scenario, error) {
	if s.failure != nil {
		return scenario.Scenario{}, s.failure
	}

	found, exists := s.scenarios[id]
	if !exists {
		return scenario.Scenario{}, scenario.ErrNotFound
	}

	return found, nil
}

func embeddedCatalog(t *testing.T) stubCatalog {
	t.Helper()

	loaded, err := scenario.LoadFS(scenarios.Files())
	if err != nil {
		t.Fatalf("не удалось загрузить сценарии: %v", err)
	}

	catalog := stubCatalog{scenarios: make(map[scenario.ID]scenario.Scenario, len(loaded))}
	for _, found := range loaded {
		catalog.scenarios[found.ID] = found
	}

	return catalog
}

func newService(t *testing.T, catalog attempt.ScenarioCatalog) *attempt.Service {
	t.Helper()

	return attempt.NewService(
		catalog,
		memory.NewAttemptRepository(),
		&clock.Fixed{Moment: startMoment, Step: time.Minute},
		&identifier.Sequential{Prefix: "attempt"},
	)
}

func TestStartAttemptRevealsFirstNodes(t *testing.T) {
	service := newService(t, embeddedCatalog(t))

	view, err := service.Start(context.Background(), "buyer-fake-delivery")
	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}

	if view.Status != attempt.StatusInProgress {
		t.Errorf("статус = %q, ожидался in_progress", view.Status)
	}

	if view.Score != attempt.InitialScore {
		t.Errorf("score = %d, ожидался %d", view.Score, attempt.InitialScore)
	}

	if view.Scenario.Version != 1 {
		t.Errorf("версия сценария = %d, ожидалась 1", view.Scenario.Version)
	}

	if len(view.RevealedNodes) == 0 {
		t.Fatal("должен быть раскрыт хотя бы один узел")
	}

	// Раскрытие обязано останавливаться ровно на решении.
	last := view.RevealedNodes[len(view.RevealedNodes)-1]
	if last.Type != scenario.NodeTypeDecision {
		t.Errorf("последний раскрытый узел = %q, ожидался decision", last.Type)
	}

	if view.CurrentNodeID != last.ID {
		t.Errorf("текущий узел = %q, ожидался %q", view.CurrentNodeID, last.ID)
	}

	if len(last.Choices) < 2 {
		t.Errorf("вариантов выбора = %d, ожидалось минимум 2", len(last.Choices))
	}

	if len(view.Decisions) != 0 {
		t.Errorf("решений = %d, ожидалось 0", len(view.Decisions))
	}

	if view.CompletedAt != nil {
		t.Error("новая попытка не должна быть завершённой")
	}
}

// Все сообщения до первого решения выдаются одной пачкой,
// иначе фронтенду пришлось бы запрашивать каждую реплику отдельно.
func TestStartAttemptRevealsAllMessagesBeforeDecision(t *testing.T) {
	service := newService(t, embeddedCatalog(t))

	view, err := service.Start(context.Background(), "seller-payment-already-sent")
	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}

	messages := 0

	for _, node := range view.RevealedNodes[:len(view.RevealedNodes)-1] {
		if node.Type != scenario.NodeTypeMessage {
			t.Fatalf("до решения ожидались только сообщения, встречен %q", node.Type)
		}

		messages++
	}

	if messages == 0 {
		t.Error("перед решением должно быть хотя бы одно сообщение")
	}
}

func TestStartAttemptRejectsUnknownScenario(t *testing.T) {
	service := newService(t, embeddedCatalog(t))

	if _, err := service.Start(context.Background(), "unknown-scenario"); !errors.Is(err, scenario.ErrNotFound) {
		t.Errorf("ошибка = %v, ожидалась scenario.ErrNotFound", err)
	}
}

func TestStartAttemptPropagatesCatalogFailure(t *testing.T) {
	catalogFailure := errors.New("каталог недоступен")
	service := newService(t, stubCatalog{failure: catalogFailure})

	if _, err := service.Start(context.Background(), "buyer-fake-delivery"); !errors.Is(err, catalogFailure) {
		t.Errorf("ошибка = %v, ожидалась обёртка над %v", err, catalogFailure)
	}
}

func TestStartAttemptRejectsInactiveScenario(t *testing.T) {
	catalog := embeddedCatalog(t)

	inactive := buildScenario(t, func(draft *scenario.Draft) {
		draft.ID = "inactive-scenario"
		draft.IsActive = false
	})
	catalog.scenarios[inactive.ID] = inactive

	service := newService(t, catalog)

	if _, err := service.Start(context.Background(), inactive.ID); !errors.Is(err, scenario.ErrNotFound) {
		t.Errorf("ошибка = %v, ожидалась scenario.ErrNotFound", err)
	}
}

func TestGetAttemptRestoresState(t *testing.T) {
	service := newService(t, embeddedCatalog(t))

	started, err := service.Start(context.Background(), "buyer-fake-delivery")
	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}

	restored, err := service.Get(context.Background(), started.ID)
	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}

	if restored.ID != started.ID {
		t.Errorf("идентификатор = %q, ожидался %q", restored.ID, started.ID)
	}

	if restored.CurrentNodeID != started.CurrentNodeID {
		t.Errorf("текущий узел = %q, ожидался %q", restored.CurrentNodeID, started.CurrentNodeID)
	}

	if len(restored.RevealedNodes) != len(started.RevealedNodes) {
		t.Errorf("раскрытых узлов = %d, ожидалось %d",
			len(restored.RevealedNodes), len(started.RevealedNodes))
	}

	if restored.Score != started.Score {
		t.Errorf("score = %d, ожидался %d", restored.Score, started.Score)
	}
}

func TestGetAttemptUnknown(t *testing.T) {
	service := newService(t, embeddedCatalog(t))

	if _, err := service.Get(context.Background(), "no-such-attempt"); !errors.Is(err, attempt.ErrNotFound) {
		t.Errorf("ошибка = %v, ожидалась ErrNotFound", err)
	}
}

// Начатое прохождение доигрывается на зафиксированной версии: подмена
// содержимого сценария не должна молча менять уже идущую попытку.
func TestGetAttemptDetectsScenarioVersionChange(t *testing.T) {
	catalog := embeddedCatalog(t)
	service := newService(t, catalog)

	started, err := service.Start(context.Background(), "buyer-fake-delivery")
	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}

	updated := buildScenario(t, func(draft *scenario.Draft) {
		draft.ID = "buyer-fake-delivery"
		draft.Version = 2
	})
	catalog.scenarios["buyer-fake-delivery"] = updated

	if _, err := service.Get(context.Background(), started.ID); !errors.Is(err, attempt.ErrScenarioVersionChanged) {
		t.Errorf("ошибка = %v, ожидалась ErrScenarioVersionChanged", err)
	}
}

func TestStartAttemptRespectsCanceledContext(t *testing.T) {
	service := newService(t, embeddedCatalog(t))

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if _, err := service.Start(ctx, "buyer-fake-delivery"); !errors.Is(err, context.Canceled) {
		t.Errorf("ошибка = %v, ожидалась context.Canceled", err)
	}
}

// Публичное состояние попытки не должно содержать скрытых данных сценария.
func TestViewHidesInternalChoiceData(t *testing.T) {
	service := newService(t, embeddedCatalog(t))

	view, err := service.Start(context.Background(), "buyer-fake-delivery")
	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}

	decision := view.RevealedNodes[len(view.RevealedNodes)-1]

	for _, choice := range decision.Choices {
		if choice.ID == "" || choice.Label == "" {
			t.Errorf("вариант выбора неполон: %+v", choice)
		}
	}

	// В публичной проекции выбора есть ровно два поля — идентификатор и подпись.
	// Любое новое поле здесь означало бы утечку серверных данных.
	choiceType := reflect.TypeOf(attempt.PublicChoice{})
	if choiceType.NumField() != 2 {
		t.Errorf("полей в PublicChoice = %d, ожидалось 2", choiceType.NumField())
	}

	allowed := map[string]struct{}{"ID": {}, "Label": {}}
	for i := range choiceType.NumField() {
		if _, ok := allowed[choiceType.Field(i).Name]; !ok {
			t.Errorf("в PublicChoice появилось поле %q", choiceType.Field(i).Name)
		}
	}
}

// buildScenario собирает валидный сценарий и применяет к нему изменение.
func buildScenario(t *testing.T, mutate func(draft *scenario.Draft)) scenario.Scenario {
	t.Helper()

	draft := scenario.Draft{
		ID:               "test-scenario",
		Version:          1,
		Slug:             "test-scenario",
		Role:             scenario.RoleBuyer,
		Title:            "Проверочный сценарий",
		Description:      "Используется в тестах.",
		Difficulty:       scenario.DifficultyEasy,
		EstimatedMinutes: 1,
		StartNodeID:      "greeting",
		IsActive:         true,
		Nodes: []scenario.Node{
			{
				ID:         "greeting",
				Type:       scenario.NodeTypeMessage,
				Sender:     "seller",
				Text:       "Здравствуйте.",
				NextNodeID: "channel-decision",
			},
			{
				ID:             "channel-decision",
				Type:           scenario.NodeTypeDecision,
				DecisionPrompt: "Что вы сделаете?",
				Choices: []scenario.Choice{
					{
						ID: "safe", Label: "Безопасно", NextNodeID: "safe-ending",
						Criticality: scenario.CriticalityLow,
						Consequence: scenario.Consequence{
							Severity: scenario.SeveritySafe, Title: "Верно", Explanation: "Так безопаснее.",
						},
					},
					{
						ID: "risky", Label: "Опасно", NextNodeID: "unsafe-ending",
						SafetyScore: -20, Criticality: scenario.CriticalityHigh,
						Consequence: scenario.Consequence{
							Severity: scenario.SeverityDangerous, Title: "Ошибка", Explanation: "Так делать нельзя.",
						},
					},
				},
			},
			{
				ID:   "safe-ending",
				Type: scenario.NodeTypeTerminal,
				TerminalOutcome: &scenario.Outcome{
					Type: scenario.OutcomeSafe, Title: "Хорошо", Explanation: "Всё в порядке.",
				},
			},
			{
				ID:   "unsafe-ending",
				Type: scenario.NodeTypeTerminal,
				TerminalOutcome: &scenario.Outcome{
					Type: scenario.OutcomeUnsafe, Title: "Плохо", Explanation: "Деньги потеряны.",
				},
			},
		},
	}

	mutate(&draft)

	built, err := scenario.New(draft)
	if err != nil {
		t.Fatalf("не удалось собрать сценарий: %v", err)
	}

	return built
}
