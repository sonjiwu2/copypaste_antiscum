//go:build integration

package app_test

import (
	"encoding/json"
	"net/http"
	"testing"
)

type progressBody struct {
	Summary struct {
		CompletedAttempts  int `json:"completedAttempts"`
		CompletedScenarios int `json:"completedScenarios"`
		AverageScore       int `json:"averageScore"`
		BestScore          int `json:"bestScore"`
		LatestScore        int `json:"latestScore"`
	} `json:"summary"`
	ActiveAttempts []struct {
		AttemptID string `json:"attemptId"`
	} `json:"activeAttempts"`
	ScenarioProgress []struct {
		ScenarioID    string `json:"scenarioId"`
		Attempts      int    `json:"attempts"`
		LastScore     int    `json:"lastScore"`
		BestScore     int    `json:"bestScore"`
		PreviousScore *int   `json:"previousScore"`
		Improvement   *int   `json:"improvement"`
	} `json:"scenarioProgress"`
	Skills []struct {
		Code  string `json:"code"`
		Value int    `json:"value"`
	} `json:"skills"`
	WeakRiskTags []struct {
		Code             string `json:"code"`
		Mistakes         int    `json:"mistakes"`
		WeightedSeverity int    `json:"weightedSeverity"`
	} `json:"weakRiskTags"`
	RecentAttempts []struct {
		AttemptID string `json:"attemptId"`
		Score     int    `json:"score"`
		Decisions int    `json:"decisions"`
	} `json:"recentAttempts"`
	Recommendations []struct {
		ScenarioID string `json:"scenarioId"`
		Reason     string `json:"reason"`
	} `json:"recommendedScenarios"`
}

// play проходит сценарий покупателя до конца указанными выборами.
func play(t *testing.T, user *client, keyPrefix string, choices []string) attemptBody {
	t.Helper()

	const scenarioID = "buyer-fake-delivery"

	var started attemptBody
	if err := json.Unmarshal(
		user.call(t, http.MethodPost, "/api/v1/attempts",
			`{"scenarioId":"`+scenarioID+`"}`, http.StatusCreated),
		&started,
	); err != nil {
		t.Fatalf("не удалось разобрать созданную попытку: %v", err)
	}

	node := started.CurrentNodeID

	for i, choiceID := range choices {
		var transition struct {
			CurrentNodeID string `json:"currentNodeId"`
		}

		body := `{"nodeId":"` + node + `","choiceId":"` + choiceID +
			`","idempotencyKey":"` + keyPrefix + "-" + string(rune('a'+i)) + `"}`

		if err := json.Unmarshal(
			user.call(t, http.MethodPost, "/api/v1/attempts/"+started.AttemptID+"/choices",
				body, http.StatusOK),
			&transition,
		); err != nil {
			t.Fatalf("не удалось разобрать переход: %v", err)
		}

		node = transition.CurrentNodeID
	}

	return started
}

func fetchProgress(t *testing.T, user *client) progressBody {
	t.Helper()

	var found progressBody
	if err := json.Unmarshal(
		user.call(t, http.MethodGet, "/api/v1/progress", "", http.StatusOK),
		&found,
	); err != nil {
		t.Fatalf("не удалось разобрать прогресс: %v", err)
	}

	return found
}

// Прогресс считается по сохранённым фактам, поэтому переживает перезапуск.
func TestProgressSurvivesRestart(t *testing.T) {
	cfg := postgresConfig(t)
	prepareDatabase(t, cfg)

	user := newClient(newApplication(t, cfg).Handler())

	// Первое прохождение — опасное, второе — безопасное.
	play(t, user, "first",
		[]string{"move-to-messenger", "open-payment-link", "send-prepay"})
	play(t, user, "second",
		[]string{"stay-on-platform", "check-in-app", "refuse-prepay"})

	before := fetchProgress(t, user)

	user.use(newApplication(t, cfg).Handler())

	after := fetchProgress(t, user)

	if before.Summary != after.Summary {
		t.Errorf("сводка изменилась после перезапуска: %+v и %+v", before.Summary, after.Summary)
	}

	if after.Summary.CompletedAttempts != 2 || after.Summary.CompletedScenarios != 1 {
		t.Errorf("сводка = %+v, ожидались две попытки одного сценария", after.Summary)
	}

	if after.Summary.BestScore != 100 || after.Summary.LatestScore != 100 {
		t.Errorf("лучший/последний = %d/%d, ожидалось 100/100",
			after.Summary.BestScore, after.Summary.LatestScore)
	}

	if after.Summary.AverageScore != 60 {
		t.Errorf("среднее = %d, ожидалось 60", after.Summary.AverageScore)
	}

	if len(after.ScenarioProgress) != 1 {
		t.Fatalf("сценариев = %d, ожидался 1", len(after.ScenarioProgress))
	}

	item := after.ScenarioProgress[0]
	if item.Attempts != 2 || item.LastScore != 100 || item.BestScore != 100 {
		t.Errorf("прогресс сценария = %+v", item)
	}

	if item.PreviousScore == nil || *item.PreviousScore != 20 {
		t.Errorf("предыдущий результат = %v, ожидалось 20", item.PreviousScore)
	}

	if item.Improvement == nil || *item.Improvement != 80 {
		t.Errorf("улучшение = %v, ожидалось 80", item.Improvement)
	}

	if len(after.RecentAttempts) != 2 {
		t.Errorf("история = %d записей, ожидалось 2", len(after.RecentAttempts))
	}

	if len(after.WeakRiskTags) == 0 || len(after.Skills) == 0 {
		t.Error("аналитика ошибок должна восстанавливаться из базы")
	}
}

// Повтор завершающего запроса не должен удваивать статистику.
func TestProgressDoesNotDoubleCountReplay(t *testing.T) {
	cfg := postgresConfig(t)
	prepareDatabase(t, cfg)

	user := newClient(newApplication(t, cfg).Handler())

	started := play(t, user, "key",
		[]string{"move-to-messenger", "open-payment-link", "send-prepay"})

	before := fetchProgress(t, user)

	// Повторная доставка последнего запроса: попытка уже завершена.
	user.call(t, http.MethodPost, "/api/v1/attempts/"+started.AttemptID+"/choices",
		`{"nodeId":"prepay-decision","choiceId":"send-prepay","idempotencyKey":"key-c"}`,
		http.StatusOK)

	after := fetchProgress(t, user)

	if before.Summary != after.Summary {
		t.Errorf("повтор изменил сводку: %+v и %+v", before.Summary, after.Summary)
	}

	if len(after.RecentAttempts) != 1 || after.RecentAttempts[0].Decisions != 3 {
		t.Errorf("история = %+v, ожидалась одна запись с тремя решениями", after.RecentAttempts)
	}
}

// Прогресс изолирован между профилями на всём пути до базы.
func TestProgressIsIsolatedBetweenProfiles(t *testing.T) {
	cfg := postgresConfig(t)
	prepareDatabase(t, cfg)

	handler := newApplication(t, cfg).Handler()

	owner := newClient(handler)
	stranger := newClient(handler)

	play(t, owner, "key",
		[]string{"stay-on-platform", "check-in-app", "refuse-prepay"})

	ownerProgress := fetchProgress(t, owner)
	if ownerProgress.Summary.CompletedAttempts != 1 {
		t.Fatalf("прогресс владельца = %+v", ownerProgress.Summary)
	}

	strangerProgress := fetchProgress(t, stranger)
	if strangerProgress.Summary.CompletedAttempts != 0 || len(strangerProgress.RecentAttempts) != 0 {
		t.Errorf("прогресс чужого профиля = %+v, ожидался пустым", strangerProgress)
	}

	// Новому профилю всё равно есть что предложить.
	if len(strangerProgress.Recommendations) == 0 {
		t.Error("новому профилю нужна рекомендация")
	}
}

// Незавершённое прохождение видно отдельно и не влияет на статистику.
func TestProgressSeparatesActiveAttempts(t *testing.T) {
	cfg := postgresConfig(t)
	prepareDatabase(t, cfg)

	user := newClient(newApplication(t, cfg).Handler())

	play(t, user, "done",
		[]string{"stay-on-platform", "check-in-app", "refuse-prepay"})

	var active attemptBody
	if err := json.Unmarshal(
		user.call(t, http.MethodPost, "/api/v1/attempts",
			`{"scenarioId":"seller-payment-already-sent"}`, http.StatusCreated),
		&active,
	); err != nil {
		t.Fatalf("не удалось разобрать попытку: %v", err)
	}

	found := fetchProgress(t, user)

	if found.Summary.CompletedAttempts != 1 || found.Summary.CompletedScenarios != 1 {
		t.Errorf("сводка = %+v, активная попытка попала в статистику", found.Summary)
	}

	if len(found.ActiveAttempts) != 1 || found.ActiveAttempts[0].AttemptID != active.AttemptID {
		t.Errorf("активные попытки = %+v", found.ActiveAttempts)
	}

	if len(found.ScenarioProgress) != 2 {
		t.Errorf("сценариев в прогрессе = %d, ожидалось 2", len(found.ScenarioProgress))
	}
}
