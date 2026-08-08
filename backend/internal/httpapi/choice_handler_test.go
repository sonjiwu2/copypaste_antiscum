package httpapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func submitChoiceRaw(t *testing.T, router http.Handler, attemptID, body string) *httptest.ResponseRecorder {
	t.Helper()

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost,
		"/api/v1/attempts/"+attemptID+"/choices", strings.NewReader(body)))

	return recorder
}

func submitChoiceOK(t *testing.T, router http.Handler, attemptID, nodeID, choiceID, key string) transitionResponse {
	t.Helper()

	body := fmt.Sprintf(`{"nodeId":%q,"choiceId":%q,"idempotencyKey":%q}`, nodeID, choiceID, key)
	recorder := submitChoiceRaw(t, router, attemptID, body)

	if recorder.Code != http.StatusOK {
		t.Fatalf("статус = %d, ожидался 200, тело: %s", recorder.Code, recorder.Body.String())
	}

	var transition transitionResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &transition); err != nil {
		t.Fatalf("не удалось разобрать ответ: %v", err)
	}

	return transition
}

func TestSubmitChoiceEndpoint(t *testing.T) {
	router := newTestRouter(t)
	started := startAttempt(t, router, "buyer-iphone-deposit")

	transition := submitChoiceOK(t, router, started.AttemptID,
		started.CurrentNodeID, "agree-deposit", "key-1")

	if transition.AttemptID != started.AttemptID {
		t.Errorf("идентификатор = %q, ожидался %q", transition.AttemptID, started.AttemptID)
	}

	if transition.Status != "in_progress" {
		t.Errorf("статус = %q, ожидался in_progress", transition.Status)
	}

	if transition.Score != 70 {
		t.Errorf("score = %d, ожидался 70", transition.Score)
	}

	if transition.AcceptedChoice.ChoiceID != "agree-deposit" || transition.AcceptedChoice.Label == "" {
		t.Errorf("принятый выбор = %+v", transition.AcceptedChoice)
	}

	if transition.Consequence.Severity != "dangerous" || transition.Consequence.Explanation == "" {
		t.Errorf("последствие = %+v", transition.Consequence)
	}

	if len(transition.RevealedNodes) == 0 {
		t.Fatal("переход должен раскрывать новые узлы")
	}

	if transition.CurrentNodeID == started.CurrentNodeID {
		t.Error("текущий узел должен измениться после выбора")
	}
}

func TestSubmitChoiceEndpointErrors(t *testing.T) {
	testCases := []struct {
		name       string
		body       func(started attemptResponse) string
		attemptID  func(started attemptResponse) string
		wantStatus int
		wantCode   string
	}{
		{
			name:       "битый JSON",
			body:       func(attemptResponse) string { return `{"nodeId":` },
			wantStatus: http.StatusBadRequest,
			wantCode:   CodeInvalidRequest,
		},
		{
			name:       "нет nodeId",
			body:       func(attemptResponse) string { return `{"choiceId":"x","idempotencyKey":"k"}` },
			wantStatus: http.StatusBadRequest,
			wantCode:   CodeInvalidRequest,
		},
		{
			name: "нет idempotencyKey",
			body: func(s attemptResponse) string {
				return fmt.Sprintf(`{"nodeId":%q,"choiceId":"request-live-video"}`, s.CurrentNodeID)
			},
			wantStatus: http.StatusBadRequest,
			wantCode:   CodeInvalidRequest,
		},
		{
			name: "неизвестная попытка",
			body: func(s attemptResponse) string {
				return fmt.Sprintf(`{"nodeId":%q,"choiceId":"request-live-video","idempotencyKey":"k"}`, s.CurrentNodeID)
			},
			attemptID:  func(attemptResponse) string { return "no-such-attempt" },
			wantStatus: http.StatusNotFound,
			wantCode:   CodeAttemptNotFound,
		},
		{
			name: "устаревший узел",
			body: func(attemptResponse) string {
				return `{"nodeId":"seller-greeting","choiceId":"request-live-video","idempotencyKey":"k"}`
			},
			wantStatus: http.StatusConflict,
			wantCode:   CodeStaleNode,
		},
		{
			name: "неизвестный вариант выбора",
			body: func(s attemptResponse) string {
				return fmt.Sprintf(`{"nodeId":%q,"choiceId":"ghost-choice","idempotencyKey":"k"}`, s.CurrentNodeID)
			},
			wantStatus: http.StatusUnprocessableEntity,
			wantCode:   CodeChoiceNotFound,
		},
		{
			name: "вариант выбора из другого сценария",
			body: func(s attemptResponse) string {
				return fmt.Sprintf(`{"nodeId":%q,"choiceId":"refund-difference-to-buyer","idempotencyKey":"k"}`, s.CurrentNodeID)
			},
			wantStatus: http.StatusUnprocessableEntity,
			wantCode:   CodeChoiceNotFound,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			router := newTestRouter(t)
			started := startAttempt(t, router, "buyer-iphone-deposit")

			attemptID := started.AttemptID
			if testCase.attemptID != nil {
				attemptID = testCase.attemptID(started)
			}

			recorder := submitChoiceRaw(t, router, attemptID, testCase.body(started))

			if recorder.Code != testCase.wantStatus {
				t.Fatalf("статус = %d, ожидался %d, тело: %s",
					recorder.Code, testCase.wantStatus, recorder.Body.String())
			}

			assertErrorCode(t, recorder.Body.Bytes(), testCase.wantCode)
		})
	}
}

func TestSubmitChoiceEndpointIdempotency(t *testing.T) {
	router := newTestRouter(t)
	started := startAttempt(t, router, "buyer-iphone-deposit")

	first := submitChoiceOK(t, router, started.AttemptID, started.CurrentNodeID, "agree-deposit", "key-1")
	second := submitChoiceOK(t, router, started.AttemptID, started.CurrentNodeID, "agree-deposit", "key-1")

	if second.Score != first.Score || second.CurrentNodeID != first.CurrentNodeID {
		t.Errorf("повтор вернул другой результат: %+v против %+v", second, first)
	}

	// Тот же ключ с другим выбором — конфликт.
	body := fmt.Sprintf(`{"nodeId":%q,"choiceId":"request-live-video","idempotencyKey":"key-1"}`,
		started.CurrentNodeID)

	recorder := submitChoiceRaw(t, router, started.AttemptID, body)
	if recorder.Code != http.StatusConflict {
		t.Fatalf("статус = %d, ожидался 409", recorder.Code)
	}

	assertErrorCode(t, recorder.Body.Bytes(), CodeIdempotencyKeyConflict)
}

func TestSubmitChoiceEndpointRejectsCompletedAttempt(t *testing.T) {
	router := newTestRouter(t)
	started := startAttempt(t, router, "buyer-iphone-deposit")

	current := started.CurrentNodeID
	for i, choiceID := range []string{"request-live-video", "request-random-action", "refuse-without-profile-check"} {
		transition := submitChoiceOK(t, router, started.AttemptID, current, choiceID,
			fmt.Sprintf("key-%d", i))
		current = transition.CurrentNodeID
	}

	body := fmt.Sprintf(`{"nodeId":%q,"choiceId":"request-live-video","idempotencyKey":"key-new"}`, current)

	recorder := submitChoiceRaw(t, router, started.AttemptID, body)
	if recorder.Code != http.StatusConflict {
		t.Fatalf("статус = %d, ожидался 409, тело: %s", recorder.Code, recorder.Body.String())
	}

	assertErrorCode(t, recorder.Body.Bytes(), CodeAttemptAlreadyCompleted)
}

// Ответ на выбор раскрывает последствие сделанного шага, но не подсказывает,
// куда ведут варианты следующего решения.
func TestSubmitChoiceResponseHidesFutureBranches(t *testing.T) {
	router := newTestRouter(t)
	started := startAttempt(t, router, "buyer-iphone-deposit")

	body := fmt.Sprintf(`{"nodeId":%q,"choiceId":"request-live-video","idempotencyKey":"key-1"}`,
		started.CurrentNodeID)

	recorder := submitChoiceRaw(t, router, started.AttemptID, body)
	if recorder.Code != http.StatusOK {
		t.Fatalf("статус = %d, ожидался 200", recorder.Code)
	}

	var payload any
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("не удалось разобрать ответ: %v", err)
	}

	forbidden := []string{"nextNodeId", "safetyScore", "criticality", "riskTags", "skillEffects", "isActive"}

	for _, field := range collectJSONKeys(payload) {
		for _, hidden := range forbidden {
			if field == hidden {
				t.Errorf("в ответе присутствует скрытое поле %q", field)
			}
		}
	}

	// Финальные узлы ещё не достигнуты и не должны упоминаться.
	for _, hiddenNode := range []string{"safe-full-ending", "unsafe-double-deposit-ending"} {
		if strings.Contains(recorder.Body.String(), hiddenNode) {
			t.Errorf("в ответе виден будущий узел %q", hiddenNode)
		}
	}
}

// Перезагрузка страницы в середине прохождения не должна терять прогресс:
// клиент обязан получить весь диалог, все прошлые выборы с объяснениями
// и текущее решение — но не будущие ветки.
func TestResumeInTheMiddleOfJourney(t *testing.T) {
	router := newTestRouter(t)
	started := startAttempt(t, router, "buyer-iphone-deposit")

	first := submitChoiceOK(t, router, started.AttemptID, started.CurrentNodeID,
		"request-live-video", "step-0")
	second := submitChoiceOK(t, router, started.AttemptID, first.CurrentNodeID,
		"request-random-action", "step-1")

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet,
		"/api/v1/attempts/"+started.AttemptID, nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("статус = %d, ожидался 200", recorder.Code)
	}

	var resumed attemptResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &resumed); err != nil {
		t.Fatalf("не удалось разобрать ответ: %v", err)
	}

	if resumed.Status != "in_progress" {
		t.Errorf("статус = %q, ожидался in_progress", resumed.Status)
	}

	if resumed.CurrentNodeID != second.CurrentNodeID {
		t.Errorf("текущий узел = %q, ожидался %q", resumed.CurrentNodeID, second.CurrentNodeID)
	}

	if resumed.Score != second.Score {
		t.Errorf("score = %d, ожидался %d", resumed.Score, second.Score)
	}

	// Весь диалог с начала: раскрытые узлы старта плюс оба перехода.
	wantNodes := len(started.RevealedNodes) + len(first.RevealedNodes) + len(second.RevealedNodes)
	if len(resumed.RevealedNodes) != wantNodes {
		t.Errorf("раскрытых узлов = %d, ожидалось %d", len(resumed.RevealedNodes), wantNodes)
	}

	if resumed.RevealedNodes[0].ID != started.RevealedNodes[0].ID {
		t.Errorf("диалог начинается с %q, ожидался %q",
			resumed.RevealedNodes[0].ID, started.RevealedNodes[0].ID)
	}

	// Прошлые выборы вместе с уже показанными объяснениями.
	if len(resumed.Decisions) != 2 {
		t.Fatalf("решений = %d, ожидалось 2", len(resumed.Decisions))
	}

	for i, want := range []string{"request-live-video", "request-random-action"} {
		if resumed.Decisions[i].ChoiceID != want {
			t.Errorf("решение %d = %q, ожидалось %q", i, resumed.Decisions[i].ChoiceID, want)
		}

		if resumed.Decisions[i].Label == "" || resumed.Decisions[i].Consequence.Explanation == "" {
			t.Errorf("решение %d потеряло подпись или объяснение", i)
		}
	}

	// Текущее решение доступно для ответа.
	currentNode := resumed.RevealedNodes[len(resumed.RevealedNodes)-1]
	if currentNode.ID != resumed.CurrentNodeID || currentNode.Type != "decision" {
		t.Fatalf("текущий узел = %+v", currentNode)
	}

	if len(currentNode.Choices) < 2 {
		t.Errorf("вариантов выбора = %d, ожидалось минимум 2", len(currentNode.Choices))
	}

	// Финалы ещё не достигнуты и не должны быть видны.
	for _, hiddenNode := range []string{"safe-full-ending", "unsafe-double-deposit-ending"} {
		if strings.Contains(recorder.Body.String(), hiddenNode) {
			t.Errorf("в ответе виден будущий узел %q", hiddenNode)
		}
	}

	// После перезагрузки прохождение продолжается с того же места.
	final := submitChoiceOK(t, router, started.AttemptID, resumed.CurrentNodeID,
		"refuse-without-profile-check", "step-2")
	if final.Status != "completed" {
		t.Errorf("статус = %q, ожидался completed", final.Status)
	}
}

// Полный путь пользователя: список, старт, три выбора, финал, чтение результата.
func TestFullJourneyThroughAPI(t *testing.T) {
	testCases := []struct {
		name        string
		scenarioID  string
		choices     []string
		wantOutcome string
		wantScore   int
	}{
		{
			name:        "покупатель проходит безопасно",
			scenarioID:  "buyer-iphone-deposit",
			choices:     []string{"request-live-video", "request-random-action", "refuse-without-profile-check"},
			wantOutcome: "safe",
			wantScore:   95,
		},
		{
			name:        "продавец теряет доступ",
			scenarioID:  "seller-third-party-overpayment",
			choices:     []string{"accept-third-party-transfer", "refund-difference-to-buyer", "handover-camera"},
			wantOutcome: "unsafe",
			wantScore:   0,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			router := newTestRouter(t)

			catalog := doRequest(t, "/api/v1/scenarios")
			if catalog.Code != http.StatusOK {
				t.Fatalf("каталог вернул %d", catalog.Code)
			}

			started := startAttempt(t, router, testCase.scenarioID)
			current := started.CurrentNodeID

			var last transitionResponse

			for i, choiceID := range testCase.choices {
				last = submitChoiceOK(t, router, started.AttemptID, current, choiceID,
					fmt.Sprintf("step-%d", i))
				current = last.CurrentNodeID
			}

			if last.Status != "completed" {
				t.Fatalf("статус = %q, ожидался completed", last.Status)
			}

			if last.Outcome != testCase.wantOutcome {
				t.Errorf("итог = %q, ожидался %q", last.Outcome, testCase.wantOutcome)
			}

			if last.Score != testCase.wantScore {
				t.Errorf("score = %d, ожидался %d", last.Score, testCase.wantScore)
			}

			if last.CompletedAt == nil {
				t.Error("завершённая попытка должна иметь время завершения")
			}

			// Итоговое состояние доступно обычным чтением попытки.
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet,
				"/api/v1/attempts/"+started.AttemptID, nil))

			if recorder.Code != http.StatusOK {
				t.Fatalf("чтение попытки вернуло %d", recorder.Code)
			}

			var finished attemptResponse
			if err := json.Unmarshal(recorder.Body.Bytes(), &finished); err != nil {
				t.Fatalf("не удалось разобрать ответ: %v", err)
			}

			if finished.Status != "completed" || finished.Score != testCase.wantScore {
				t.Errorf("итоговая попытка = %+v", finished)
			}

			if len(finished.Decisions) != len(testCase.choices) {
				t.Errorf("решений = %d, ожидалось %d", len(finished.Decisions), len(testCase.choices))
			}

			for i, decision := range finished.Decisions {
				if decision.ChoiceID != testCase.choices[i] {
					t.Errorf("решение %d = %q, ожидалось %q", i, decision.ChoiceID, testCase.choices[i])
				}

				if decision.Consequence.Explanation == "" {
					t.Errorf("решение %d: объяснение потеряно", i)
				}
			}

			terminal := finished.RevealedNodes[len(finished.RevealedNodes)-1]
			if terminal.Type != "terminal" || terminal.Outcome == nil {
				t.Fatalf("последний узел = %+v", terminal)
			}

			if terminal.Outcome.Type != testCase.wantOutcome {
				t.Errorf("итог финального узла = %q, ожидался %q",
					terminal.Outcome.Type, testCase.wantOutcome)
			}
		})
	}
}
