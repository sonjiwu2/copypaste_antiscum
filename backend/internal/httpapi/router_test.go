package httpapi

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/attempt"
	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/platform/clock"
	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/platform/identifier"
	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/scenario"
	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/storage/memory"
	"github.com/sonjiwu2/copypaste_antiscum/backend/scenarios"
)

// newTestRouter собирает роутер поверх встроенных сценариев,
// чтобы HTTP-тесты работали с тем же каталогом, что и приложение.
func newTestRouter(t *testing.T) http.Handler {
	t.Helper()

	catalog, err := scenario.LoadFS(scenarios.Files())
	if err != nil {
		t.Fatalf("не удалось загрузить сценарии: %v", err)
	}

	repository, err := memory.NewScenarioRepository(catalog)
	if err != nil {
		t.Fatalf("не удалось собрать каталог: %v", err)
	}

	return NewRouter(RouterDeps{
		Logger:          slog.New(slog.NewJSONHandler(io.Discard, nil)),
		RequestIDs:      &identifier.Sequential{Prefix: "request"},
		MaxRequestBytes: 4096,
		Scenarios:       scenario.NewService(repository),
		Attempts: attempt.NewService(
			repository,
			memory.NewAttemptRepository(),
			&clock.Fixed{Moment: time.Date(2026, time.August, 3, 12, 0, 0, 0, time.UTC), Step: time.Minute},
			&identifier.Sequential{Prefix: "attempt"},
		),
	})
}

func TestRoutes(t *testing.T) {
	testCases := []struct {
		name       string
		method     string
		target     string
		wantStatus int
		wantCode   string
	}{
		{
			name:       "проверка живости отвечает успехом",
			method:     http.MethodGet,
			target:     "/healthz",
			wantStatus: http.StatusOK,
		},
		{
			name:       "неизвестный маршрут отдаёт ошибку в едином формате",
			method:     http.MethodGet,
			target:     "/api/v1/unknown",
			wantStatus: http.StatusNotFound,
			wantCode:   CodeNotFound,
		},
		{
			name:       "корневой путь отдаёт ошибку в едином формате",
			method:     http.MethodGet,
			target:     "/",
			wantStatus: http.StatusNotFound,
			wantCode:   CodeNotFound,
		},
	}

	router := newTestRouter(t)

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, httptest.NewRequest(testCase.method, testCase.target, nil))

			if recorder.Code != testCase.wantStatus {
				t.Fatalf("статус = %d, ожидался %d", recorder.Code, testCase.wantStatus)
			}

			if contentType := recorder.Header().Get("Content-Type"); !strings.HasPrefix(contentType, "application/json") {
				t.Errorf("Content-Type = %q, ожидался JSON", contentType)
			}

			if testCase.wantCode == "" {
				return
			}

			assertErrorCode(t, recorder.Body.Bytes(), testCase.wantCode)
		})
	}
}

func TestHealthResponseBody(t *testing.T) {
	recorder := httptest.NewRecorder()
	newTestRouter(t).ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/healthz", nil))

	var body healthResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("не удалось разобрать тело ответа: %v", err)
	}

	if body.Status != "ok" {
		t.Errorf("status = %q, ожидалось \"ok\"", body.Status)
	}
}

func TestRequestIDIsGeneratedAndReturned(t *testing.T) {
	recorder := httptest.NewRecorder()
	newTestRouter(t).ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/unknown", nil))

	header := recorder.Header().Get(requestHeaderName)
	if header == "" {
		t.Fatal("заголовок X-Request-ID должен присутствовать")
	}

	var envelope errorEnvelope
	if err := json.Unmarshal(recorder.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("не удалось разобрать тело ответа: %v", err)
	}

	if envelope.Error.RequestID != header {
		t.Errorf("requestId в теле = %q, в заголовке = %q", envelope.Error.RequestID, header)
	}
}

func TestRequestIDFromClientIsPreserved(t *testing.T) {
	const clientRequestID = "client-supplied-id"

	request := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	request.Header.Set(requestHeaderName, clientRequestID)

	recorder := httptest.NewRecorder()
	newTestRouter(t).ServeHTTP(recorder, request)

	if got := recorder.Header().Get(requestHeaderName); got != clientRequestID {
		t.Errorf("X-Request-ID = %q, ожидался %q", got, clientRequestID)
	}
}

func TestPanicRecoveryHidesInternalDetails(t *testing.T) {
	const secret = "секрет из stack trace"

	panicking := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic(secret)
	})

	handler := withRequestContext(
		slog.New(slog.NewJSONHandler(io.Discard, nil)),
		&identifier.Sequential{Prefix: "request"},
		withPanicRecovery(panicking),
	)

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/boom", nil))

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("статус = %d, ожидался %d", recorder.Code, http.StatusInternalServerError)
	}

	if strings.Contains(recorder.Body.String(), secret) {
		t.Error("тело ответа не должно содержать внутренние детали паники")
	}

	assertErrorCode(t, recorder.Body.Bytes(), CodeInternalError)
}

func TestBodyLimitRejectsOversizedRequest(t *testing.T) {
	const limit = 16

	var readErr error

	reading := http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		_, readErr = io.ReadAll(r.Body)
	})

	handler := withBodyLimit(limit, reading)

	request := httptest.NewRequest(http.MethodPost, "/api/v1/attempts", strings.NewReader(strings.Repeat("a", limit+1)))
	handler.ServeHTTP(httptest.NewRecorder(), request)

	if readErr == nil {
		t.Fatal("чтение тела сверх лимита должно возвращать ошибку")
	}
}

func assertErrorCode(t *testing.T, body []byte, wantCode string) {
	t.Helper()

	var envelope errorEnvelope
	if err := json.Unmarshal(body, &envelope); err != nil {
		t.Fatalf("не удалось разобрать тело ошибки: %v", err)
	}

	if envelope.Error.Code != wantCode {
		t.Errorf("code = %q, ожидался %q", envelope.Error.Code, wantCode)
	}

	if envelope.Error.Message == "" {
		t.Error("сообщение об ошибке не должно быть пустым")
	}
}
