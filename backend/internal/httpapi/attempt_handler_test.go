package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// attemptHiddenFields — данные, которых не должно быть в ответе о начатой попытке.
// Пользователь ещё не сделал выбор, поэтому ни переходы, ни последствия,
// ни внутренние коэффициенты ему недоступны.
var attemptHiddenFields = []string{
	"nextNodeId", "startNodeId", "safetyScore", "criticality",
	"riskTags", "skillEffects", "consequence", "outcome", "isActive",
}

func startAttempt(t *testing.T, router http.Handler, scenarioID string) attemptResponse {
	t.Helper()

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/attempts",
		strings.NewReader(`{"scenarioId":"`+scenarioID+`"}`))

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("статус = %d, ожидался 201, тело: %s", recorder.Code, recorder.Body.String())
	}

	var body attemptResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("не удалось разобрать ответ: %v", err)
	}

	return body
}

func TestStartAttemptEndpoint(t *testing.T) {
	router := newTestRouter(t)
	started := startAttempt(t, router, "buyer-fake-delivery")

	if started.AttemptID == "" {
		t.Error("идентификатор попытки не должен быть пустым")
	}

	if started.Status != "in_progress" {
		t.Errorf("статус = %q, ожидался in_progress", started.Status)
	}

	if started.Score != 100 {
		t.Errorf("score = %d, ожидался 100", started.Score)
	}

	if started.Scenario.ID != "buyer-fake-delivery" || started.Scenario.Version != 1 {
		t.Errorf("сценарий = %+v, ожидался buyer-fake-delivery версии 1", started.Scenario)
	}

	if len(started.RevealedNodes) < 2 {
		t.Fatalf("раскрытых узлов = %d, ожидалось минимум 2", len(started.RevealedNodes))
	}

	last := started.RevealedNodes[len(started.RevealedNodes)-1]
	if last.Type != "decision" || started.CurrentNodeID != last.ID {
		t.Errorf("остановка не на решении: тип %q, текущий узел %q", last.Type, started.CurrentNodeID)
	}

	if len(last.Choices) < 2 {
		t.Errorf("вариантов выбора = %d, ожидалось минимум 2", len(last.Choices))
	}

	if len(started.Decisions) != 0 {
		t.Errorf("решений = %d, ожидалось 0", len(started.Decisions))
	}

	if started.StartedAt == "" || started.UpdatedAt == "" {
		t.Error("метки времени должны быть заполнены")
	}

	if started.CompletedAt != nil {
		t.Error("новая попытка не должна иметь время завершения")
	}
}

func TestStartAttemptEndpointRejectsBadRequests(t *testing.T) {
	testCases := []struct {
		name       string
		body       string
		wantStatus int
		wantCode   string
	}{
		{
			name:       "неизвестный сценарий",
			body:       `{"scenarioId":"unknown-scenario"}`,
			wantStatus: http.StatusNotFound,
			wantCode:   CodeScenarioNotFound,
		},
		{
			name:       "битый JSON",
			body:       `{"scenarioId":`,
			wantStatus: http.StatusBadRequest,
			wantCode:   CodeInvalidRequest,
		},
		{
			name:       "пустое тело",
			body:       ``,
			wantStatus: http.StatusBadRequest,
			wantCode:   CodeInvalidRequest,
		},
		{
			name:       "отсутствует scenarioId",
			body:       `{}`,
			wantStatus: http.StatusBadRequest,
			wantCode:   CodeInvalidRequest,
		},
		{
			name:       "пустой scenarioId",
			body:       `{"scenarioId":""}`,
			wantStatus: http.StatusBadRequest,
			wantCode:   CodeInvalidRequest,
		},
	}

	router := newTestRouter(t)

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/api/v1/attempts",
				strings.NewReader(testCase.body)))

			if recorder.Code != testCase.wantStatus {
				t.Fatalf("статус = %d, ожидался %d, тело: %s",
					recorder.Code, testCase.wantStatus, recorder.Body.String())
			}

			assertErrorCode(t, recorder.Body.Bytes(), testCase.wantCode)
		})
	}
}

// Слишком большое тело — отдельная ошибка, а не «некорректный JSON»:
// иначе клиент чинил бы не ту проблему.
func TestStartAttemptEndpointRejectsOversizedBody(t *testing.T) {
	oversized := `{"scenarioId":"` + strings.Repeat("x", 8192) + `"}`

	recorder := httptest.NewRecorder()
	newTestRouter(t).ServeHTTP(recorder, httptest.NewRequest(http.MethodPost,
		"/api/v1/attempts", strings.NewReader(oversized)))

	if recorder.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("статус = %d, ожидался 413, тело: %s", recorder.Code, recorder.Body.String())
	}

	assertErrorCode(t, recorder.Body.Bytes(), CodePayloadTooLarge)
}

// Ответ после перезагрузки страницы должен полностью восстанавливать экран.
func TestGetAttemptEndpointRestoresState(t *testing.T) {
	router := newTestRouter(t)
	started := startAttempt(t, router, "seller-payment-already-sent")

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet,
		"/api/v1/attempts/"+started.AttemptID, nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("статус = %d, ожидался 200", recorder.Code)
	}

	var restored attemptResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &restored); err != nil {
		t.Fatalf("не удалось разобрать ответ: %v", err)
	}

	if restored.AttemptID != started.AttemptID {
		t.Errorf("идентификатор = %q, ожидался %q", restored.AttemptID, started.AttemptID)
	}

	if restored.CurrentNodeID != started.CurrentNodeID {
		t.Errorf("текущий узел = %q, ожидался %q", restored.CurrentNodeID, started.CurrentNodeID)
	}

	if len(restored.RevealedNodes) != len(started.RevealedNodes) {
		t.Fatalf("раскрытых узлов = %d, ожидалось %d",
			len(restored.RevealedNodes), len(started.RevealedNodes))
	}

	for i, node := range restored.RevealedNodes {
		if node.ID != started.RevealedNodes[i].ID {
			t.Errorf("узел %d = %q, ожидался %q", i, node.ID, started.RevealedNodes[i].ID)
		}

		if node.Type == "message" && node.Text == "" {
			t.Errorf("узел %q: текст сообщения потерян", node.ID)
		}
	}

	if restored.Score != started.Score {
		t.Errorf("score = %d, ожидался %d", restored.Score, started.Score)
	}
}

func TestGetAttemptEndpointUnknown(t *testing.T) {
	recorder := httptest.NewRecorder()
	newTestRouter(t).ServeHTTP(recorder, httptest.NewRequest(http.MethodGet,
		"/api/v1/attempts/no-such-attempt", nil))

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("статус = %d, ожидался 404", recorder.Code)
	}

	assertErrorCode(t, recorder.Body.Bytes(), CodeAttemptNotFound)
}

// До первого выбора наружу не должно уходить ничего, что подсказывает ответ.
func TestAttemptResponsesHideFutureBranches(t *testing.T) {
	router := newTestRouter(t)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/api/v1/attempts",
		strings.NewReader(`{"scenarioId":"buyer-fake-delivery"}`)))

	if recorder.Code != http.StatusCreated {
		t.Fatalf("статус = %d, ожидался 201", recorder.Code)
	}

	var payload any
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("не удалось разобрать ответ: %v", err)
	}

	for _, field := range collectJSONKeys(payload) {
		for _, hidden := range attemptHiddenFields {
			if field == hidden {
				t.Errorf("в ответе присутствует скрытое поле %q", field)
			}
		}
	}

	// Дополнительная проверка по сырому тексту: подписи вариантов видны,
	// а идентификаторы следующих узлов встречаться не должны.
	body := recorder.Body.String()
	for _, hiddenNode := range []string{"delivery-message", "safe-ending", "unsafe-ending"} {
		if strings.Contains(body, hiddenNode) {
			t.Errorf("в ответе виден будущий узел %q", hiddenNode)
		}
	}
}

// Каждая попытка независима: два запуска не делят состояние.
func TestAttemptsAreIndependent(t *testing.T) {
	router := newTestRouter(t)

	first := startAttempt(t, router, "buyer-fake-delivery")
	second := startAttempt(t, router, "seller-payment-already-sent")

	if first.AttemptID == second.AttemptID {
		t.Fatal("идентификаторы попыток должны различаться")
	}

	if first.Scenario.Role == second.Scenario.Role {
		t.Errorf("роли совпали: %q", first.Scenario.Role)
	}

	if first.CurrentNodeID == second.CurrentNodeID {
		t.Errorf("текущие узлы совпали: %q", first.CurrentNodeID)
	}
}
