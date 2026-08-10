package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// progressHiddenFields — данные, которых не должно быть в ответе прогресса.
//
// Аналитика показывает результат уже сделанных выборов, но не имеет права
// раскрывать устройство сценариев: иначе экран прогресса подсказал бы
// правильные ответы в непройденных прохождениях.
var progressHiddenFields = []string{
	"nextNodeId", "startNodeId", "safetyScore", "criticality",
	"nodes", "choices", "consequence", "isActive",
}

func fetchProgress(t *testing.T, client http.Handler) progressResponse {
	t.Helper()

	recorder := httptest.NewRecorder()
	client.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/progress", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("статус = %d, ожидался 200, тело: %s", recorder.Code, recorder.Body.String())
	}

	var body progressResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("не удалось разобрать ответ: %v", err)
	}

	return body
}

// Новый пользователь получает 200 с пустыми списками, а не отказ.
func TestProgressEndpointForNewProfile(t *testing.T) {
	router := newTestRouter(t)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/progress", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("статус = %d, ожидался 200", recorder.Code)
	}

	// Пустые коллекции сериализуются массивами, а не null:
	// клиенту не нужно проверять каждое поле перед перебором.
	body := recorder.Body.String()
	for _, field := range []string{
		`"activeAttempts":[]`, `"scenarioProgress":[]`, `"skills":[]`,
		`"weakRiskTags":[]`, `"recentAttempts":[]`,
	} {
		if !strings.Contains(body, field) {
			t.Errorf("в ответе нет %s: %s", field, body)
		}
	}

	var parsed progressResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &parsed); err != nil {
		t.Fatalf("не удалось разобрать ответ: %v", err)
	}

	if parsed.Summary.CompletedAttempts != 0 || parsed.Summary.BestScore != 0 {
		t.Errorf("сводка = %+v, ожидалась нулевая", parsed.Summary)
	}

	if len(parsed.Recommendations) == 0 {
		t.Error("новому пользователю нужна рекомендация с чего начать")
	}
}

// Незавершённое прохождение показывается отдельно и не влияет на статистику.
func TestProgressSeparatesActiveAttempt(t *testing.T) {
	router := newTestRouter(t)
	started := startAttempt(t, router, "buyer-fake-delivery")

	found := fetchProgress(t, router)

	if found.Summary.CompletedAttempts != 0 {
		t.Errorf("завершённых попыток = %d, активная попала в статистику",
			found.Summary.CompletedAttempts)
	}

	if len(found.ActiveAttempts) != 1 || found.ActiveAttempts[0].AttemptID != started.AttemptID {
		t.Fatalf("активные попытки = %+v", found.ActiveAttempts)
	}

	if len(found.ScenarioProgress) != 1 || found.ScenarioProgress[0].ActiveAttemptID != started.AttemptID {
		t.Errorf("прогресс по сценарию = %+v", found.ScenarioProgress)
	}
}

// Полный путь: пройденный сценарий попадает в сводку, историю и аналитику.
func TestProgressAfterCompletedAttempt(t *testing.T) {
	router := newTestRouter(t)
	started := startAttempt(t, router, "buyer-fake-delivery")

	node := started.CurrentNodeID
	for i, choiceID := range []string{"move-to-messenger", "open-payment-link", "send-prepay"} {
		transition := submitChoiceOK(t, router, started.AttemptID, node, choiceID,
			"key-"+string(rune('a'+i)))
		node = transition.CurrentNodeID
	}

	found := fetchProgress(t, router)

	if found.Summary.CompletedAttempts != 1 || found.Summary.CompletedScenarios != 1 {
		t.Errorf("сводка = %+v, ожидалась одна завершённая попытка", found.Summary)
	}

	if found.Summary.BestScore != 20 || found.Summary.LatestScore != 20 {
		t.Errorf("лучший/последний = %d/%d, ожидалось 20/20",
			found.Summary.BestScore, found.Summary.LatestScore)
	}

	if len(found.ActiveAttempts) != 0 {
		t.Errorf("активных попыток = %d, ожидалось 0", len(found.ActiveAttempts))
	}

	if len(found.RecentAttempts) != 1 || found.RecentAttempts[0].Decisions != 3 {
		t.Errorf("история = %+v, ожидалась одна запись с тремя решениями", found.RecentAttempts)
	}

	if found.RecentAttempts[0].Outcome != "unsafe" {
		t.Errorf("итог = %q, ожидался unsafe", found.RecentAttempts[0].Outcome)
	}

	// Три опасных выбора дали отрицательные навыки и слабые метки риска.
	if len(found.Skills) == 0 {
		t.Fatal("навыки должны накопиться после прохождения")
	}

	if found.Skills[0].Value >= 0 {
		t.Errorf("самый слабый навык = %+v, ожидалось отрицательное значение", found.Skills[0])
	}

	if len(found.WeakRiskTags) == 0 {
		t.Fatal("слабые метки риска должны появиться после ошибок")
	}

	if found.WeakRiskTags[0].Mistakes == 0 || found.WeakRiskTags[0].WeightedSeverity == 0 {
		t.Errorf("слабая метка = %+v", found.WeakRiskTags[0])
	}
}

// Повторное прохождение показывает динамику и не затирает предыдущее.
func TestProgressShowsImprovement(t *testing.T) {
	router := newTestRouter(t)

	// Первое прохождение — опасное.
	first := startAttempt(t, router, "buyer-fake-delivery")
	node := first.CurrentNodeID

	for i, choiceID := range []string{"move-to-messenger", "open-payment-link", "send-prepay"} {
		transition := submitChoiceOK(t, router, first.AttemptID, node, choiceID,
			"first-"+string(rune('a'+i)))
		node = transition.CurrentNodeID
	}

	// Второе — безопасное.
	second := startAttempt(t, router, "buyer-fake-delivery")
	node = second.CurrentNodeID

	for i, choiceID := range []string{"stay-on-platform", "check-in-app", "refuse-prepay"} {
		transition := submitChoiceOK(t, router, second.AttemptID, node, choiceID,
			"second-"+string(rune('a'+i)))
		node = transition.CurrentNodeID
	}

	found := fetchProgress(t, router)

	if found.Summary.CompletedAttempts != 2 || found.Summary.CompletedScenarios != 1 {
		t.Errorf("сводка = %+v, ожидались две попытки одного сценария", found.Summary)
	}

	if found.Summary.BestScore != 100 || found.Summary.LatestScore != 100 {
		t.Errorf("лучший/последний = %d/%d, ожидалось 100/100",
			found.Summary.BestScore, found.Summary.LatestScore)
	}

	if len(found.ScenarioProgress) != 1 {
		t.Fatalf("сценариев = %d, ожидался 1", len(found.ScenarioProgress))
	}

	item := found.ScenarioProgress[0]

	if item.Attempts != 2 {
		t.Errorf("попыток = %d, ожидалось 2: история не должна затираться", item.Attempts)
	}

	if item.PreviousScore == nil || *item.PreviousScore != 20 {
		t.Errorf("предыдущий результат = %v, ожидалось 20", item.PreviousScore)
	}

	if item.Improvement == nil || *item.Improvement != 80 {
		t.Errorf("улучшение = %v, ожидалось 80", item.Improvement)
	}

	// История содержит оба прохождения, свежее — первым.
	if len(found.RecentAttempts) != 2 || found.RecentAttempts[0].AttemptID != second.AttemptID {
		t.Errorf("история = %+v", found.RecentAttempts)
	}
}

// Прогресс принадлежит профилю: чужие прохождения в него не попадают.
func TestProgressIsIsolatedBetweenProfiles(t *testing.T) {
	router := NewRouter(testRouterDeps(t))

	owner := newBrowser(router)
	stranger := newBrowser(router)

	started := startAttempt(t, owner, "buyer-fake-delivery")

	node := started.CurrentNodeID
	for i, choiceID := range []string{"stay-on-platform", "check-in-app", "refuse-prepay"} {
		transition := submitChoiceOK(t, owner, started.AttemptID, node, choiceID,
			"key-"+string(rune('a'+i)))
		node = transition.CurrentNodeID
	}

	ownerProgress := fetchProgress(t, owner)
	if ownerProgress.Summary.CompletedAttempts != 1 {
		t.Fatalf("прогресс владельца = %+v", ownerProgress.Summary)
	}

	strangerProgress := fetchProgress(t, stranger)
	if strangerProgress.Summary.CompletedAttempts != 0 {
		t.Errorf("прогресс чужого профиля = %+v, ожидался пустым", strangerProgress.Summary)
	}

	if len(strangerProgress.RecentAttempts) != 0 {
		t.Errorf("история чужого профиля = %+v, ожидалась пустой", strangerProgress.RecentAttempts)
	}
}

// Повтор запроса с тем же ключом не должен удваивать статистику.
func TestProgressDoesNotDoubleCountRetries(t *testing.T) {
	router := newTestRouter(t)
	started := startAttempt(t, router, "buyer-fake-delivery")

	body := `{"nodeId":"` + started.CurrentNodeID +
		`","choiceId":"move-to-messenger","idempotencyKey":"key-retry"}`

	for range 3 {
		if recorder := submitChoiceRaw(t, router, started.AttemptID, body); recorder.Code != http.StatusOK {
			t.Fatalf("статус = %d, ожидался 200", recorder.Code)
		}
	}

	found := fetchProgress(t, router)

	if len(found.ActiveAttempts) != 1 {
		t.Fatalf("активных попыток = %d, ожидалась 1", len(found.ActiveAttempts))
	}

	// Попытка ещё не завершена, поэтому в статистику не входит,
	// но и продублироваться не может.
	if found.Summary.CompletedAttempts != 0 {
		t.Errorf("завершённых попыток = %d, ожидалось 0", found.Summary.CompletedAttempts)
	}

	if len(found.ScenarioProgress) != 1 {
		t.Errorf("сценариев в прогрессе = %d, ожидался 1", len(found.ScenarioProgress))
	}
}

// Прогресс не раскрывает устройство сценариев.
func TestProgressHidesScenarioInternals(t *testing.T) {
	router := newTestRouter(t)
	started := startAttempt(t, router, "buyer-fake-delivery")

	submitChoiceOK(t, router, started.AttemptID, started.CurrentNodeID, "move-to-messenger", "key-1")

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/progress", nil))

	body := recorder.Body.String()
	for _, field := range progressHiddenFields {
		if strings.Contains(body, `"`+field+`"`) {
			t.Errorf("ответ прогресса содержит скрытое поле %q: %s", field, body)
		}
	}
}
