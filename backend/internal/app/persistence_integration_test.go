//go:build integration

// Сквозная проверка долговечности: прохождение обязано переживать
// перезапуск процесса. Ради этого и вводится PostgreSQL.
package app_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/app"
	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/config"
	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/storage/postgres"
)

const testDatabaseURLEnv = "TEST_DATABASE_URL"

// postgresConfig собирает рабочую конфигурацию с настоящей базой.
func postgresConfig(t *testing.T) config.Config {
	t.Helper()

	url, ok := os.LookupEnv(testDatabaseURLEnv)
	if !ok || url == "" {
		t.Skipf("%s не задан: интеграционные тесты пропущены", testDatabaseURLEnv)
	}

	cfg := config.Default()
	cfg.Database.URL = url

	return cfg
}

// prepareDatabase приводит схему в рабочее состояние и убирает прежние попытки.
func prepareDatabase(t *testing.T, cfg config.Config) {
	t.Helper()

	ctx := context.Background()

	db, err := postgres.OpenSQL(ctx, cfg.Database.URL)
	if err != nil {
		t.Fatalf("не удалось подключиться к тестовой базе: %v", err)
	}

	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("не удалось закрыть соединение: %v", err)
		}
	}()

	if err := postgres.RunMigration(ctx, db, postgres.CommandUp, io.Discard); err != nil {
		t.Fatalf("не удалось применить миграции: %v", err)
	}

	resetData(t, db)
}

// resetData убирает попытки и архив сценариев: приложение при старте наполнит
// архив заново, а тесты не должны зависеть от следов других пакетов.
func resetData(t *testing.T, db *sql.DB) {
	t.Helper()

	_, err := db.ExecContext(context.Background(),
		"TRUNCATE attempts, scenario_versions, profiles CASCADE")
	if err != nil {
		t.Fatalf("не удалось очистить данные: %v", err)
	}
}

// newApplication собирает приложение так же, как это делает cmd/api.
func newApplication(t *testing.T, cfg config.Config) *app.Application {
	t.Helper()

	application, err := app.New(context.Background(), cfg, discardLogger())
	if err != nil {
		t.Fatalf("не удалось собрать приложение: %v", err)
	}

	t.Cleanup(application.Close)

	return application
}

type attemptBody struct {
	AttemptID     string `json:"attemptId"`
	Status        string `json:"status"`
	Score         int    `json:"score"`
	CurrentNodeID string `json:"currentNodeId"`
	RevealedNodes []struct {
		ID      string `json:"id"`
		Type    string `json:"type"`
		Choices []struct {
			ID string `json:"id"`
		} `json:"choices"`
	} `json:"revealedNodes"`
	Decisions []struct {
		NodeID      string `json:"nodeId"`
		ChoiceID    string `json:"choiceId"`
		Label       string `json:"label"`
		Consequence struct {
			Severity    string `json:"severity"`
			Title       string `json:"title"`
			Explanation string `json:"explanation"`
		} `json:"consequence"`
	} `json:"decisions"`
}

// client имитирует браузер: хранит cookie профиля между запросами.
//
// Без этого каждый запрос приходил бы от нового анонимного пользователя,
// и созданная попытка была бы чужой уже на следующем шаге.
type client struct {
	handler http.Handler
	cookies map[string]*http.Cookie
}

func newClient(handler http.Handler) *client {
	return &client{handler: handler, cookies: make(map[string]*http.Cookie)}
}

func (c *client) do(t *testing.T, method, target, body string) *httptest.ResponseRecorder {
	t.Helper()

	var reader io.Reader
	if body != "" {
		reader = strings.NewReader(body)
	}

	request := httptest.NewRequest(method, target, reader)
	for _, cookie := range c.cookies {
		request.AddCookie(cookie)
	}

	recorder := httptest.NewRecorder()
	c.handler.ServeHTTP(recorder, request)

	for _, cookie := range recorder.Result().Cookies() {
		c.cookies[cookie.Name] = cookie
	}

	return recorder
}

func (c *client) call(t *testing.T, method, target, body string, wantStatus int) []byte {
	t.Helper()

	recorder := c.do(t, method, target, body)

	if recorder.Code != wantStatus {
		t.Fatalf("%s %s: статус = %d, ожидался %d, тело: %s",
			method, target, recorder.Code, wantStatus, recorder.Body.String())
	}

	return recorder.Body.Bytes()
}

// use переносит cookie профиля на новый экземпляр приложения:
// перезапуск сервера не должен выглядеть для пользователя сменой личности.
func (c *client) use(handler http.Handler) {
	c.handler = handler
}

// Полный срез Phase 3: старт попытки, выбор, перезапуск приложения,
// продолжение прохождения на том же состоянии.
func TestAttemptSurvivesApplicationRestart(t *testing.T) {
	cfg := postgresConfig(t)
	prepareDatabase(t, cfg)

	user := newClient(newApplication(t, cfg).Handler())

	var started attemptBody
	if err := json.Unmarshal(
		user.call(t, http.MethodPost, "/api/v1/attempts",
			`{"scenarioId":"buyer-fake-delivery"}`, http.StatusCreated),
		&started,
	); err != nil {
		t.Fatalf("не удалось разобрать созданную попытку: %v", err)
	}

	if started.Score != 100 || started.Status != "in_progress" {
		t.Fatalf("начальное состояние = %d/%q, ожидалось 100/in_progress", started.Score, started.Status)
	}

	user.call(t, http.MethodPost,
		"/api/v1/attempts/"+started.AttemptID+"/choices",
		`{"nodeId":"`+started.CurrentNodeID+`","choiceId":"move-to-messenger","idempotencyKey":"key-1"}`,
		http.StatusOK)

	// Приложение пересоздаётся целиком: новый пул, новый репозиторий.
	// Состояние может прийти только из базы.
	user.use(newApplication(t, cfg).Handler())

	var restored attemptBody
	if err := json.Unmarshal(
		user.call(t, http.MethodGet, "/api/v1/attempts/"+started.AttemptID, "",
			http.StatusOK),
		&restored,
	); err != nil {
		t.Fatalf("не удалось разобрать восстановленную попытку: %v", err)
	}

	if restored.Score != 80 {
		t.Errorf("score после перезапуска = %d, ожидался 80", restored.Score)
	}

	if len(restored.Decisions) != 1 {
		t.Fatalf("решений = %d, ожидалось 1", len(restored.Decisions))
	}

	decision := restored.Decisions[0]
	if decision.ChoiceID != "move-to-messenger" || decision.Label == "" {
		t.Errorf("решение = %+v, ожидался move-to-messenger с подписью", decision)
	}

	if decision.Consequence.Severity != "dangerous" || decision.Consequence.Explanation == "" {
		t.Errorf("последствие = %+v, ожидалось опасное с объяснением", decision.Consequence)
	}

	if len(restored.RevealedNodes) < 4 {
		t.Errorf("раскрытых узлов = %d, переписка восстановлена не полностью",
			len(restored.RevealedNodes))
	}

	// Прохождение продолжается после перезапуска: попытка не «застыла».
	user.call(t, http.MethodPost,
		"/api/v1/attempts/"+started.AttemptID+"/choices",
		`{"nodeId":"`+restored.CurrentNodeID+`","choiceId":"check-in-app","idempotencyKey":"key-2"}`,
		http.StatusOK)

	var continued attemptBody
	if err := json.Unmarshal(
		user.call(t, http.MethodGet, "/api/v1/attempts/"+started.AttemptID, "",
			http.StatusOK),
		&continued,
	); err != nil {
		t.Fatalf("не удалось разобрать попытку: %v", err)
	}

	if len(continued.Decisions) != 2 {
		t.Errorf("решений = %d, ожидалось 2", len(continued.Decisions))
	}
}

// Повтор запроса с тем же ключом после перезапуска обязан вернуть тот же
// результат и не применить эффекты второй раз.
func TestIdempotentReplaySurvivesRestart(t *testing.T) {
	cfg := postgresConfig(t)
	prepareDatabase(t, cfg)

	user := newClient(newApplication(t, cfg).Handler())

	var started attemptBody
	if err := json.Unmarshal(
		user.call(t, http.MethodPost, "/api/v1/attempts",
			`{"scenarioId":"buyer-fake-delivery"}`, http.StatusCreated),
		&started,
	); err != nil {
		t.Fatalf("не удалось разобрать созданную попытку: %v", err)
	}

	choice := `{"nodeId":"` + started.CurrentNodeID +
		`","choiceId":"move-to-messenger","idempotencyKey":"key-replay"}`

	original := user.call(t, http.MethodPost,
		"/api/v1/attempts/"+started.AttemptID+"/choices", choice, http.StatusOK)

	user.use(newApplication(t, cfg).Handler())

	replay := user.call(t, http.MethodPost,
		"/api/v1/attempts/"+started.AttemptID+"/choices", choice, http.StatusOK)

	if string(original) != string(replay) {
		t.Errorf("повтор вернул другой ответ:\n%s\n%s", original, replay)
	}

	var restored attemptBody
	if err := json.Unmarshal(
		user.call(t, http.MethodGet, "/api/v1/attempts/"+started.AttemptID, "",
			http.StatusOK),
		&restored,
	); err != nil {
		t.Fatalf("не удалось разобрать попытку: %v", err)
	}

	if len(restored.Decisions) != 1 {
		t.Errorf("решений = %d, эффекты применены больше одного раза", len(restored.Decisions))
	}

	if restored.Score != 80 {
		t.Errorf("score = %d, ожидался 80", restored.Score)
	}
}

// Тот же ключ с другими данными остаётся конфликтом и после перезапуска.
func TestIdempotencyConflictSurvivesRestart(t *testing.T) {
	cfg := postgresConfig(t)
	prepareDatabase(t, cfg)

	user := newClient(newApplication(t, cfg).Handler())

	var started attemptBody
	if err := json.Unmarshal(
		user.call(t, http.MethodPost, "/api/v1/attempts",
			`{"scenarioId":"buyer-fake-delivery"}`, http.StatusCreated),
		&started,
	); err != nil {
		t.Fatalf("не удалось разобрать созданную попытку: %v", err)
	}

	user.call(t, http.MethodPost,
		"/api/v1/attempts/"+started.AttemptID+"/choices",
		`{"nodeId":"`+started.CurrentNodeID+`","choiceId":"move-to-messenger","idempotencyKey":"key-1"}`,
		http.StatusOK)

	user.use(newApplication(t, cfg).Handler())

	body := user.call(t, http.MethodPost,
		"/api/v1/attempts/"+started.AttemptID+"/choices",
		`{"nodeId":"`+started.CurrentNodeID+`","choiceId":"stay-on-platform","idempotencyKey":"key-1"}`,
		http.StatusConflict)

	if !strings.Contains(string(body), "IDEMPOTENCY_KEY_CONFLICT") {
		t.Errorf("тело = %s, ожидался код IDEMPOTENCY_KEY_CONFLICT", body)
	}
}

// Готовность отражает доступность базы, а не только живость процесса.
func TestReadinessReportsDatabase(t *testing.T) {
	cfg := postgresConfig(t)
	prepareDatabase(t, cfg)

	user := newClient(newApplication(t, cfg).Handler())

	user.call(t, http.MethodGet, "/readyz", "", http.StatusOK)
	user.call(t, http.MethodGet, "/healthz", "", http.StatusOK)
}

// Каталог сценариев обязан пережить синхронизацию версий без изменений:
// архив читает файлы, но не подменяет содержимое каталога.
func TestScenarioCatalogIsUnchangedAfterSync(t *testing.T) {
	cfg := postgresConfig(t)
	prepareDatabase(t, cfg)

	user := newClient(newApplication(t, cfg).Handler())

	var catalog struct {
		Scenarios []struct {
			ID      string `json:"id"`
			Version int    `json:"version"`
			Role    string `json:"role"`
		} `json:"scenarios"`
	}

	if err := json.Unmarshal(
		user.call(t, http.MethodGet, "/api/v1/scenarios", "", http.StatusOK),
		&catalog,
	); err != nil {
		t.Fatalf("не удалось разобрать каталог: %v", err)
	}

	if len(catalog.Scenarios) != 6 {
		t.Fatalf("сценариев = %d, ожидалось 6", len(catalog.Scenarios))
	}

	for _, item := range catalog.Scenarios {
		if item.Version < 1 || item.Role == "" {
			t.Errorf("сценарий %q отдан неполным: %+v", item.ID, item)
		}
	}
}

// Повторный запуск не должен менять архив: версии уже на месте.
func TestSecondStartupIsIdempotent(t *testing.T) {
	cfg := postgresConfig(t)
	prepareDatabase(t, cfg)

	newApplication(t, cfg).Close()
	newApplication(t, cfg).Close()

	db, err := postgres.OpenSQL(context.Background(), cfg.Database.URL)
	if err != nil {
		t.Fatalf("не удалось подключиться к базе: %v", err)
	}

	defer func() { _ = db.Close() }()

	var count int
	if err := db.QueryRowContext(context.Background(),
		"SELECT count(*) FROM scenario_versions").Scan(&count); err != nil {
		t.Fatalf("не удалось прочитать архив: %v", err)
	}

	if count != 6 {
		t.Errorf("версий в архиве = %d, ожидалось 6", count)
	}
}

// Профили изолированы на всём пути: чужая попытка недоступна даже при
// известном идентификаторе, и разделение переживает перезапуск.
func TestProfilesAreIsolated(t *testing.T) {
	cfg := postgresConfig(t)
	prepareDatabase(t, cfg)

	handler := newApplication(t, cfg).Handler()

	owner := newClient(handler)
	stranger := newClient(handler)

	var started attemptBody
	if err := json.Unmarshal(
		owner.call(t, http.MethodPost, "/api/v1/attempts",
			`{"scenarioId":"buyer-fake-delivery"}`, http.StatusCreated),
		&started,
	); err != nil {
		t.Fatalf("не удалось разобрать созданную попытку: %v", err)
	}

	owner.call(t, http.MethodPost, "/api/v1/attempts/"+started.AttemptID+"/choices",
		`{"nodeId":"`+started.CurrentNodeID+`","choiceId":"move-to-messenger","idempotencyKey":"key-1"}`,
		http.StatusOK)

	// Чужой профиль знает идентификатор, но доступа не получает.
	forbidden := stranger.call(t, http.MethodGet, "/api/v1/attempts/"+started.AttemptID, "",
		http.StatusForbidden)

	if !strings.Contains(string(forbidden), "ATTEMPT_FORBIDDEN") {
		t.Errorf("тело = %s, ожидался код ATTEMPT_FORBIDDEN", forbidden)
	}

	stranger.call(t, http.MethodPost, "/api/v1/attempts/"+started.AttemptID+"/choices",
		`{"nodeId":"`+started.CurrentNodeID+`","choiceId":"stay-on-platform","idempotencyKey":"key-2"}`,
		http.StatusForbidden)

	// После перезапуска разделение сохраняется: владелец хранится в базе.
	restarted := newApplication(t, cfg).Handler()
	owner.use(restarted)
	stranger.use(restarted)

	stranger.call(t, http.MethodGet, "/api/v1/attempts/"+started.AttemptID, "", http.StatusForbidden)

	var restored attemptBody
	if err := json.Unmarshal(
		owner.call(t, http.MethodGet, "/api/v1/attempts/"+started.AttemptID, "", http.StatusOK),
		&restored,
	); err != nil {
		t.Fatalf("не удалось разобрать попытку владельца: %v", err)
	}

	if len(restored.Decisions) != 1 || restored.Score != 80 {
		t.Errorf("попытка владельца = %d решений / score %d, ожидалось 1 / 80",
			len(restored.Decisions), restored.Score)
	}

	// Отказ чужому профилю не изменил попытку.
	if restored.CurrentNodeID != "link-decision" {
		t.Errorf("текущий узел = %q, ожидался link-decision", restored.CurrentNodeID)
	}
}

// Профиль создаётся один раз: повторные запросы не плодят записи.
func TestProfileIsCreatedOnce(t *testing.T) {
	cfg := postgresConfig(t)
	prepareDatabase(t, cfg)

	user := newClient(newApplication(t, cfg).Handler())

	for range 3 {
		user.call(t, http.MethodGet, "/api/v1/scenarios", "", http.StatusOK)
	}

	db, err := postgres.OpenSQL(context.Background(), cfg.Database.URL)
	if err != nil {
		t.Fatalf("не удалось подключиться к базе: %v", err)
	}

	defer func() { _ = db.Close() }()

	var count int
	if err := db.QueryRowContext(context.Background(),
		"SELECT count(*) FROM profiles").Scan(&count); err != nil {
		t.Fatalf("не удалось прочитать профили: %v", err)
	}

	if count != 1 {
		t.Errorf("профилей = %d, ожидался 1", count)
	}
}
