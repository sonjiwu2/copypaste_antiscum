package attempt_test

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/attempt"
	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/scenario"
)

// startBuyerAttempt создаёт попытку по встроенному сценарию покупателя.
func startBuyerAttempt(t *testing.T) (*attempt.Service, attempt.View) {
	t.Helper()

	service := newService(t, embeddedCatalog(t))

	view, err := service.Start(context.Background(), "buyer-fake-delivery")
	if err != nil {
		t.Fatalf("не удалось начать попытку: %v", err)
	}

	return service, view
}

func submit(
	t *testing.T,
	service *attempt.Service,
	view attempt.View,
	choiceID scenario.ChoiceID,
	key attempt.IdempotencyKey,
) (attempt.Transition, error) {
	t.Helper()

	return service.SubmitChoice(context.Background(), attempt.SubmitChoiceCommand{
		AttemptID:      view.ID,
		NodeID:         view.CurrentNodeID,
		ChoiceID:       choiceID,
		IdempotencyKey: key,
	})
}

func TestSubmitChoiceAppliesSafeOption(t *testing.T) {
	service, view := startBuyerAttempt(t)

	transition, err := submit(t, service, view, "stay-on-platform", "key-1")
	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}

	if transition.Status != attempt.StatusInProgress {
		t.Errorf("статус = %q, ожидался in_progress", transition.Status)
	}

	if transition.Score != attempt.InitialScore {
		t.Errorf("score = %d, ожидался %d", transition.Score, attempt.InitialScore)
	}

	if transition.Accepted.ChoiceID != "stay-on-platform" || transition.Accepted.Label == "" {
		t.Errorf("принятый выбор = %+v", transition.Accepted)
	}

	if transition.Consequence.Severity != scenario.SeveritySafe {
		t.Errorf("опасность = %q, ожидалась safe", transition.Consequence.Severity)
	}

	if transition.Consequence.Explanation == "" {
		t.Error("объяснение последствия должно раскрываться после выбора")
	}

	if len(transition.RevealedNodes) == 0 {
		t.Fatal("переход должен раскрывать новые узлы")
	}

	last := transition.RevealedNodes[len(transition.RevealedNodes)-1]
	if last.Type != scenario.NodeTypeDecision {
		t.Errorf("остановка на узле типа %q, ожидался decision", last.Type)
	}

	if transition.CurrentNodeID != last.ID {
		t.Errorf("текущий узел = %q, ожидался %q", transition.CurrentNodeID, last.ID)
	}
}

func TestSubmitChoiceAppliesUnsafeOption(t *testing.T) {
	service, view := startBuyerAttempt(t)

	transition, err := submit(t, service, view, "move-to-messenger", "key-1")
	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}

	if transition.Score >= attempt.InitialScore {
		t.Errorf("score = %d, опасный выбор должен уменьшать результат", transition.Score)
	}

	if transition.Consequence.Severity != scenario.SeverityDangerous {
		t.Errorf("опасность = %q, ожидалась dangerous", transition.Consequence.Severity)
	}
}

// Клиент не присылает эффекты: и результат, и навыки берутся из сценария.
func TestSubmitChoiceTakesEffectsFromScenarioOnly(t *testing.T) {
	service, view := startBuyerAttempt(t)

	if _, err := submit(t, service, view, "move-to-messenger", "key-1"); err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}

	restored, err := service.Get(context.Background(), view.ID)
	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}

	if restored.Score != 80 {
		t.Errorf("score = %d, ожидался 80 по данным сценария", restored.Score)
	}

	if len(restored.Decisions) != 1 {
		t.Fatalf("решений = %d, ожидалось 1", len(restored.Decisions))
	}

	if restored.Decisions[0].ChoiceID != "move-to-messenger" {
		t.Errorf("выбор = %q, ожидался move-to-messenger", restored.Decisions[0].ChoiceID)
	}
}

// Полное прохождение до финала: три решения, завершение и итог.
func TestSubmitChoiceReachesTerminal(t *testing.T) {
	testCases := []struct {
		name        string
		choices     []scenario.ChoiceID
		wantOutcome scenario.OutcomeType
		wantScore   int
	}{
		{
			name:        "безопасный финал",
			choices:     []scenario.ChoiceID{"stay-on-platform", "check-in-app", "refuse-prepay"},
			wantOutcome: scenario.OutcomeSafe,
			wantScore:   100,
		},
		{
			name:        "небезопасный финал",
			choices:     []scenario.ChoiceID{"move-to-messenger", "open-payment-link", "send-prepay"},
			wantOutcome: scenario.OutcomeUnsafe,
			wantScore:   20,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			service, view := startBuyerAttempt(t)

			var transition attempt.Transition

			for i, choiceID := range testCase.choices {
				var err error

				transition, err = submit(t, service, view, choiceID,
					attempt.IdempotencyKey("key-"+string(rune('a'+i))))
				if err != nil {
					t.Fatalf("шаг %d (%s) вернул ошибку: %v", i, choiceID, err)
				}

				view.CurrentNodeID = transition.CurrentNodeID
			}

			if transition.Status != attempt.StatusCompleted {
				t.Fatalf("статус = %q, ожидался completed", transition.Status)
			}

			if transition.Outcome != testCase.wantOutcome {
				t.Errorf("итог = %q, ожидался %q", transition.Outcome, testCase.wantOutcome)
			}

			if transition.Score != testCase.wantScore {
				t.Errorf("score = %d, ожидался %d", transition.Score, testCase.wantScore)
			}

			if transition.CompletedAt == nil {
				t.Error("завершённая попытка должна иметь время завершения")
			}

			last := transition.RevealedNodes[len(transition.RevealedNodes)-1]
			if last.Type != scenario.NodeTypeTerminal || last.Outcome == nil {
				t.Fatalf("последний узел = %q, итог = %v", last.Type, last.Outcome)
			}

			if last.Outcome.Title == "" || last.Outcome.Explanation == "" {
				t.Error("финальный узел должен объяснять результат")
			}
		})
	}
}

func TestSubmitChoiceRejectsInvalidRequests(t *testing.T) {
	testCases := []struct {
		name    string
		command func(view attempt.View) attempt.SubmitChoiceCommand
		wantErr error
	}{
		{
			name: "неизвестная попытка",
			command: func(view attempt.View) attempt.SubmitChoiceCommand {
				return attempt.SubmitChoiceCommand{
					AttemptID: "no-such-attempt", NodeID: view.CurrentNodeID,
					ChoiceID: "stay-on-platform", IdempotencyKey: "key-1",
				}
			},
			wantErr: attempt.ErrNotFound,
		},
		{
			name: "устаревший узел",
			command: func(view attempt.View) attempt.SubmitChoiceCommand {
				return attempt.SubmitChoiceCommand{
					AttemptID: view.ID, NodeID: "greeting",
					ChoiceID: "stay-on-platform", IdempotencyKey: "key-1",
				}
			},
			wantErr: attempt.ErrStaleNode,
		},
		{
			name: "неизвестный вариант выбора",
			command: func(view attempt.View) attempt.SubmitChoiceCommand {
				return attempt.SubmitChoiceCommand{
					AttemptID: view.ID, NodeID: view.CurrentNodeID,
					ChoiceID: "no-such-choice", IdempotencyKey: "key-1",
				}
			},
			wantErr: attempt.ErrChoiceNotFound,
		},
		{
			name: "вариант выбора другого узла",
			command: func(view attempt.View) attempt.SubmitChoiceCommand {
				return attempt.SubmitChoiceCommand{
					AttemptID: view.ID, NodeID: view.CurrentNodeID,
					ChoiceID: "refuse-prepay", IdempotencyKey: "key-1",
				}
			},
			wantErr: attempt.ErrChoiceNotFound,
		},
		{
			name: "вариант выбора другого сценария",
			command: func(view attempt.View) attempt.SubmitChoiceCommand {
				return attempt.SubmitChoiceCommand{
					AttemptID: view.ID, NodeID: view.CurrentNodeID,
					ChoiceID: "send-code", IdempotencyKey: "key-1",
				}
			},
			wantErr: attempt.ErrChoiceNotFound,
		},
		{
			name: "пустой ключ повтора",
			command: func(view attempt.View) attempt.SubmitChoiceCommand {
				return attempt.SubmitChoiceCommand{
					AttemptID: view.ID, NodeID: view.CurrentNodeID,
					ChoiceID: "stay-on-platform",
				}
			},
			wantErr: attempt.ErrEmptyIdempotencyKey,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			service, view := startBuyerAttempt(t)

			_, err := service.SubmitChoice(context.Background(), testCase.command(view))
			if !errors.Is(err, testCase.wantErr) {
				t.Fatalf("ошибка = %v, ожидалась %v", err, testCase.wantErr)
			}

			// Отклонённый запрос не должен менять состояние попытки.
			after, getErr := service.Get(context.Background(), view.ID)
			if getErr != nil {
				t.Fatalf("не удалось перечитать попытку: %v", getErr)
			}

			if after.Score != attempt.InitialScore || len(after.Decisions) != 0 {
				t.Errorf("попытка изменилась: score %d, решений %d", after.Score, len(after.Decisions))
			}
		})
	}
}

// Завершённая попытка неизменяема, но повтор последнего запроса — не изменение.
func TestSubmitChoiceOnCompletedAttempt(t *testing.T) {
	service, view := startBuyerAttempt(t)

	choices := []scenario.ChoiceID{"stay-on-platform", "check-in-app", "refuse-prepay"}
	keys := []attempt.IdempotencyKey{"key-a", "key-b", "key-c"}

	var last attempt.Transition

	for i, choiceID := range choices {
		transition, err := submit(t, service, view, choiceID, keys[i])
		if err != nil {
			t.Fatalf("шаг %d вернул ошибку: %v", i, err)
		}

		view.CurrentNodeID = transition.CurrentNodeID
		last = transition
	}

	// Новый ключ на завершённой попытке — отказ.
	_, err := service.SubmitChoice(context.Background(), attempt.SubmitChoiceCommand{
		AttemptID: view.ID, NodeID: last.CurrentNodeID,
		ChoiceID: "stay-on-platform", IdempotencyKey: "key-new",
	})
	if !errors.Is(err, attempt.ErrAlreadyCompleted) {
		t.Errorf("ошибка = %v, ожидалась ErrAlreadyCompleted", err)
	}

	// Повтор последнего запроса возвращает тот же результат, а не отказ.
	replay, err := service.SubmitChoice(context.Background(), attempt.SubmitChoiceCommand{
		AttemptID: view.ID, NodeID: last.Accepted.NodeID,
		ChoiceID: last.Accepted.ChoiceID, IdempotencyKey: "key-c",
	})
	if err != nil {
		t.Fatalf("повтор завершающего выбора вернул ошибку: %v", err)
	}

	if replay.Status != attempt.StatusCompleted || replay.Outcome != last.Outcome {
		t.Errorf("повтор = %+v, ожидался тот же финал", replay)
	}
}

func TestSubmitChoiceIsIdempotent(t *testing.T) {
	service, view := startBuyerAttempt(t)

	first, err := submit(t, service, view, "move-to-messenger", "key-1")
	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}

	second, err := submit(t, service, view, "move-to-messenger", "key-1")
	if err != nil {
		t.Fatalf("повтор вернул ошибку: %v", err)
	}

	if second.Score != first.Score {
		t.Errorf("score повтора = %d, ожидался %d", second.Score, first.Score)
	}

	if second.CurrentNodeID != first.CurrentNodeID {
		t.Errorf("узел повтора = %q, ожидался %q", second.CurrentNodeID, first.CurrentNodeID)
	}

	if len(second.RevealedNodes) != len(first.RevealedNodes) {
		t.Errorf("раскрытых узлов = %d, ожидалось %d",
			len(second.RevealedNodes), len(first.RevealedNodes))
	}

	// Главное: эффекты применены ровно один раз.
	after, err := service.Get(context.Background(), view.ID)
	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}

	if len(after.Decisions) != 1 {
		t.Errorf("решений = %d, ожидалось 1", len(after.Decisions))
	}

	if after.Score != first.Score {
		t.Errorf("score попытки = %d, ожидался %d", after.Score, first.Score)
	}
}

func TestSubmitChoiceRejectsReusedKeyWithDifferentPayload(t *testing.T) {
	service, view := startBuyerAttempt(t)

	if _, err := submit(t, service, view, "move-to-messenger", "key-1"); err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}

	_, err := submit(t, service, view, "stay-on-platform", "key-1")
	if !errors.Is(err, attempt.ErrIdempotencyConflict) {
		t.Fatalf("ошибка = %v, ожидалась ErrIdempotencyConflict", err)
	}
}

func TestSubmitChoiceRespectsCanceledContext(t *testing.T) {
	service, view := startBuyerAttempt(t)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := service.SubmitChoice(ctx, attempt.SubmitChoiceCommand{
		AttemptID: view.ID, NodeID: view.CurrentNodeID,
		ChoiceID: "stay-on-platform", IdempotencyKey: "key-1",
	})
	if !errors.Is(err, context.Canceled) {
		t.Errorf("ошибка = %v, ожидалась context.Canceled", err)
	}
}

// Параллельные отправки разных выборов из одного узла: применяется ровно один,
// остальные получают конфликт. Эффекты не могут удвоиться.
func TestSubmitChoiceUnderConcurrency(t *testing.T) {
	const writers = 12

	service, view := startBuyerAttempt(t)

	options := []scenario.ChoiceID{"stay-on-platform", "move-to-messenger"}

	var (
		waitGroup sync.WaitGroup
		mu        sync.Mutex
		accepted  int
		conflicts int
	)

	for i := range writers {
		waitGroup.Add(1)

		go func() {
			defer waitGroup.Done()

			_, err := service.SubmitChoice(context.Background(), attempt.SubmitChoiceCommand{
				AttemptID:      view.ID,
				NodeID:         view.CurrentNodeID,
				ChoiceID:       options[i%len(options)],
				IdempotencyKey: attempt.IdempotencyKey("key-" + string(rune('a'+i))),
			})

			mu.Lock()
			defer mu.Unlock()

			switch {
			case err == nil:
				accepted++
			case errors.Is(err, attempt.ErrConcurrentUpdate),
				errors.Is(err, attempt.ErrStaleNode),
				errors.Is(err, attempt.ErrAlreadyCompleted):
				conflicts++
			default:
				t.Errorf("неожиданная ошибка: %v", err)
			}
		}()
	}

	waitGroup.Wait()

	if accepted != 1 {
		t.Fatalf("принято переходов = %d, ожидался ровно 1 (конфликтов %d)", accepted, conflicts)
	}

	after, err := service.Get(context.Background(), view.ID)
	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}

	if len(after.Decisions) != 1 {
		t.Errorf("решений в попытке = %d, ожидалось 1", len(after.Decisions))
	}
}

// Повторная доставка одного и того же запроса при гонке должна возвращать
// применённый результат, а не конфликт.
func TestSubmitChoiceConcurrentDuplicateReturnsSameResult(t *testing.T) {
	const writers = 10

	service, view := startBuyerAttempt(t)

	results := make([]attempt.Transition, writers)
	failures := make([]error, writers)

	var waitGroup sync.WaitGroup

	for i := range writers {
		waitGroup.Add(1)

		go func() {
			defer waitGroup.Done()

			results[i], failures[i] = submit(t, service, view, "move-to-messenger", "same-key")
		}()
	}

	waitGroup.Wait()

	for i, err := range failures {
		if err != nil {
			t.Fatalf("запрос %d вернул ошибку: %v", i, err)
		}

		if results[i].Score != results[0].Score || results[i].CurrentNodeID != results[0].CurrentNodeID {
			t.Errorf("запрос %d вернул другой результат: %+v", i, results[i])
		}
	}

	after, err := service.Get(context.Background(), view.ID)
	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}

	if len(after.Decisions) != 1 {
		t.Errorf("решений = %d, эффекты применены больше одного раза", len(after.Decisions))
	}
}
