package attempt_test

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/attempt"
	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/platform/clock"
	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/platform/identifier"
	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/profile"
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

func (s stubCatalog) GetVersion(
	ctx context.Context,
	id scenario.ID,
	version scenario.Version,
) (scenario.Scenario, error) {
	found, err := s.Get(ctx, id)
	if err != nil {
		return scenario.Scenario{}, err
	}

	if found.Version != version {
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

// testProfile — владелец попыток во всех тестах прохождения.
const testProfile = profile.ID("test-profile-owner")

func newService(t *testing.T, catalog attempt.ScenarioCatalog) *attempt.Service {
	t.Helper()

	return attempt.NewService(
		catalog,
		catalog.(attempt.ScenarioVersions),
		memory.NewAttemptRepository(),
		&clock.Fixed{Moment: startMoment, Step: time.Minute},
		&identifier.Sequential{Prefix: "attempt"},
	)
}

func TestStartAttemptRevealsFirstNodes(t *testing.T) {
	service := newService(t, embeddedCatalog(t))

	view, err := service.Start(context.Background(), testProfile, "buyer-fake-delivery")
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

	view, err := service.Start(context.Background(), testProfile, "seller-payment-already-sent")
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

	if _, err := service.Start(context.Background(), testProfile, "unknown-scenario"); !errors.Is(err, scenario.ErrNotFound) {
		t.Errorf("ошибка = %v, ожидалась scenario.ErrNotFound", err)
	}
}

func TestStartAttemptPropagatesCatalogFailure(t *testing.T) {
	catalogFailure := errors.New("каталог недоступен")
	service := newService(t, stubCatalog{failure: catalogFailure})

	if _, err := service.Start(context.Background(), testProfile, "buyer-fake-delivery"); !errors.Is(err, catalogFailure) {
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

	if _, err := service.Start(context.Background(), testProfile, inactive.ID); !errors.Is(err, scenario.ErrNotFound) {
		t.Errorf("ошибка = %v, ожидалась scenario.ErrNotFound", err)
	}
}

func TestGetAttemptRestoresState(t *testing.T) {
	service := newService(t, embeddedCatalog(t))

	started, err := service.Start(context.Background(), testProfile, "buyer-fake-delivery")
	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}

	restored, err := service.Get(context.Background(), testProfile, started.ID)
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

	if _, err := service.Get(context.Background(), testProfile, "no-such-attempt"); !errors.Is(err, attempt.ErrNotFound) {
		t.Errorf("ошибка = %v, ожидалась ErrNotFound", err)
	}
}

// Начатое прохождение доигрывается на зафиксированной версии: подмена
// содержимого сценария не должна молча менять уже идущую попытку.
func TestGetAttemptDetectsScenarioVersionChange(t *testing.T) {
	catalog := embeddedCatalog(t)
	service := newService(t, catalog)

	started, err := service.Start(context.Background(), testProfile, "buyer-fake-delivery")
	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}

	updated := buildScenario(t, func(draft *scenario.Draft) {
		draft.ID = "buyer-fake-delivery"
		draft.Version = 2
	})
	catalog.scenarios["buyer-fake-delivery"] = updated

	if _, err := service.Get(context.Background(), testProfile, started.ID); !errors.Is(err, attempt.ErrScenarioVersionChanged) {
		t.Errorf("ошибка = %v, ожидалась ErrScenarioVersionChanged", err)
	}
}

// Выпуск новой активной версии не меняет уже начатое прохождение: чтение,
// повтор и следующие решения используют точный архивный граф версии 1,
// тогда как новая попытка сразу стартует на версии 2.
func TestAttemptContinuesOnExactVersionAfterCatalogUpdate(t *testing.T) {
	version1Current := transitionScenario(t, 1, true, -10)
	version1Archived := transitionScenario(t, 1, false, -10)
	version2Current := transitionScenario(t, 2, true, -30)

	currentV1, err := memory.NewScenarioRepository([]scenario.Scenario{version1Current})
	if err != nil {
		t.Fatalf("не удалось собрать каталог версии 1: %v", err)
	}

	currentV2, err := memory.NewScenarioRepository([]scenario.Scenario{version2Current})
	if err != nil {
		t.Fatalf("не удалось собрать каталог версии 2: %v", err)
	}

	versions, err := memory.NewScenarioRepository([]scenario.Scenario{version1Archived, version2Current})
	if err != nil {
		t.Fatalf("не удалось собрать архив версий: %v", err)
	}

	attempts := memory.NewAttemptRepository()
	testClock := &clock.Fixed{Moment: startMoment, Step: time.Minute}
	testIDs := &identifier.Sequential{Prefix: "versioned-attempt"}

	serviceV1 := attempt.NewService(currentV1, versions, attempts, testClock, testIDs)
	started, err := serviceV1.Start(context.Background(), testProfile, version1Current.ID)
	if err != nil {
		t.Fatalf("не удалось начать попытку версии 1: %v", err)
	}

	firstCommand := attempt.SubmitChoiceCommand{
		AttemptID: started.ID, ProfileID: testProfile, NodeID: "first",
		ChoiceID: "continue", IdempotencyKey: "version-1-first",
	}
	first, err := serviceV1.SubmitChoice(context.Background(), firstCommand)
	if err != nil {
		t.Fatalf("не удалось применить первый выбор версии 1: %v", err)
	}

	if first.Score != 90 || first.Consequence.Title != "Версия 1" {
		t.Fatalf("первый переход = score %d / %q, ожидалось 90 / Версия 1",
			first.Score, first.Consequence.Title)
	}

	serviceV2 := attempt.NewService(currentV2, versions, attempts, testClock, testIDs)

	// Повтор после смены активной версии должен вернуть снимок версии 1.
	replay, err := serviceV2.SubmitChoice(context.Background(), firstCommand)
	if err != nil {
		t.Fatalf("повтор после смены версии вернул ошибку: %v", err)
	}

	if replay.Score != first.Score || replay.Consequence != first.Consequence {
		t.Errorf("повтор изменился после выпуска версии 2: %+v и %+v", first, replay)
	}

	restored, err := serviceV2.Get(context.Background(), testProfile, started.ID)
	if err != nil {
		t.Fatalf("версия 1 не восстановилась после выпуска версии 2: %v", err)
	}

	if restored.Scenario.Version != 1 || restored.Score != 90 {
		t.Fatalf("восстановлена версия %d со score %d, ожидалась версия 1 / 90",
			restored.Scenario.Version, restored.Score)
	}

	completed, err := serviceV2.SubmitChoice(context.Background(), attempt.SubmitChoiceCommand{
		AttemptID: started.ID, ProfileID: testProfile, NodeID: "second",
		ChoiceID: "finish", IdempotencyKey: "version-1-second",
	})
	if err != nil {
		t.Fatalf("не удалось завершить попытку версии 1: %v", err)
	}

	if completed.Status != attempt.StatusCompleted || completed.Score != 80 ||
		completed.Consequence.Title != "Версия 1" {
		t.Errorf("финал версии 1 = status %q / score %d / %q",
			completed.Status, completed.Score, completed.Consequence.Title)
	}

	newAttempt, err := serviceV2.Start(context.Background(), testProfile, version2Current.ID)
	if err != nil {
		t.Fatalf("не удалось начать новую попытку: %v", err)
	}

	if newAttempt.Scenario.Version != 2 {
		t.Fatalf("новая попытка использует версию %d, ожидалась 2", newAttempt.Scenario.Version)
	}

	newTransition, err := serviceV2.SubmitChoice(context.Background(), attempt.SubmitChoiceCommand{
		AttemptID: newAttempt.ID, ProfileID: testProfile, NodeID: "first",
		ChoiceID: "continue", IdempotencyKey: "version-2-first",
	})
	if err != nil {
		t.Fatalf("не удалось применить выбор версии 2: %v", err)
	}

	if newTransition.Score != 70 || newTransition.Consequence.Title != "Версия 2" {
		t.Errorf("переход версии 2 = score %d / %q, ожидалось 70 / Версия 2",
			newTransition.Score, newTransition.Consequence.Title)
	}
}

func TestStartAttemptRespectsCanceledContext(t *testing.T) {
	service := newService(t, embeddedCatalog(t))

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if _, err := service.Start(ctx, testProfile, "buyer-fake-delivery"); !errors.Is(err, context.Canceled) {
		t.Errorf("ошибка = %v, ожидалась context.Canceled", err)
	}
}

// Публичное состояние попытки не должно содержать скрытых данных сценария.
func TestViewHidesInternalChoiceData(t *testing.T) {
	service := newService(t, embeddedCatalog(t))

	view, err := service.Start(context.Background(), testProfile, "buyer-fake-delivery")
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

func transitionScenario(
	t *testing.T,
	version scenario.Version,
	active bool,
	penalty int,
) scenario.Scenario {
	t.Helper()

	title := fmt.Sprintf("Версия %d", version)
	consequence := func() scenario.Consequence {
		return scenario.Consequence{
			Severity: scenario.SeverityWarning, Title: title,
			Explanation: "Проверка точной версии.",
		}
	}

	built, err := scenario.New(scenario.Draft{
		ID: "version-transition", Version: version, Slug: "version-transition",
		Role: scenario.RoleBuyer, Title: title, Difficulty: scenario.DifficultyMedium,
		EstimatedMinutes: 2, StartNodeID: "first", IsActive: active,
		Nodes: []scenario.Node{
			{
				ID: "first", Type: scenario.NodeTypeDecision, DecisionPrompt: "Первый выбор",
				Choices: []scenario.Choice{
					{ID: "continue", Label: title, NextNodeID: "second", SafetyScore: penalty,
						Criticality: scenario.CriticalityMedium, Consequence: consequence()},
					{ID: "stop", Label: "Остановиться", NextNodeID: "unsafe", SafetyScore: penalty,
						Criticality: scenario.CriticalityHigh, Consequence: consequence()},
				},
			},
			{
				ID: "second", Type: scenario.NodeTypeDecision, DecisionPrompt: "Второй выбор",
				Choices: []scenario.Choice{
					{ID: "finish", Label: title, NextNodeID: "safe", SafetyScore: penalty,
						Criticality: scenario.CriticalityMedium, Consequence: consequence()},
					{ID: "fail", Label: "Ошибка", NextNodeID: "unsafe", SafetyScore: penalty,
						Criticality: scenario.CriticalityHigh, Consequence: consequence()},
				},
			},
			{ID: "safe", Type: scenario.NodeTypeTerminal,
				TerminalOutcome: &scenario.Outcome{Type: scenario.OutcomeSafe, Title: title, Explanation: "Безопасный финал"}},
			{ID: "unsafe", Type: scenario.NodeTypeTerminal,
				TerminalOutcome: &scenario.Outcome{Type: scenario.OutcomeUnsafe, Title: title, Explanation: "Опасный финал"}},
		},
	})
	if err != nil {
		t.Fatalf("не удалось собрать сценарий перехода версий: %v", err)
	}

	return built
}

// Владение проверяется на всех операциях с попыткой: идентификатор попытки
// остаётся секретом, но после появления прогресса чужая история не должна
// быть доступна даже тому, кто угадал идентификатор.
func TestAttemptOwnershipIsEnforced(t *testing.T) {
	const stranger = profile.ID("another-profile")

	service := newService(t, embeddedCatalog(t))

	view, err := service.Start(context.Background(), testProfile, "buyer-fake-delivery")
	if err != nil {
		t.Fatalf("не удалось начать попытку: %v", err)
	}

	if _, err := service.Get(context.Background(), stranger, view.ID); !errors.Is(err, attempt.ErrForbidden) {
		t.Errorf("Get: ошибка = %v, ожидалась ErrForbidden", err)
	}

	_, err = service.SubmitChoice(context.Background(), attempt.SubmitChoiceCommand{
		AttemptID:      view.ID,
		ProfileID:      stranger,
		NodeID:         view.CurrentNodeID,
		ChoiceID:       "stay-on-platform",
		IdempotencyKey: "key-1",
	})
	if !errors.Is(err, attempt.ErrForbidden) {
		t.Errorf("SubmitChoice: ошибка = %v, ожидалась ErrForbidden", err)
	}

	// Отказ не должен менять попытку.
	after, err := service.Get(context.Background(), testProfile, view.ID)
	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}

	if len(after.Decisions) != 0 {
		t.Errorf("решений = %d, ожидалось 0", len(after.Decisions))
	}
}

// Попытка без владельца не создаётся: прогресс должен кому-то принадлежать.
func TestStartRequiresProfile(t *testing.T) {
	service := newService(t, embeddedCatalog(t))

	if _, err := service.Start(context.Background(), "", "buyer-fake-delivery"); !errors.Is(err, attempt.ErrEmptyProfileID) {
		t.Errorf("ошибка = %v, ожидалась ErrEmptyProfileID", err)
	}
}
