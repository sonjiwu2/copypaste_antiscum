package app_test

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/app"
	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/config"
)

// memoryConfig собирает приложение без базы данных.
//
// Хранилище в памяти выбирается явно: рабочий режим требует PostgreSQL,
// а модульные тесты не должны зависеть от внешней системы.
func memoryConfig() config.Config {
	cfg := config.Default()
	cfg.StorageDriver = config.StorageMemory

	return cfg
}

func discardLogger() *slog.Logger {
	return slog.New(slog.NewJSONHandler(io.Discard, nil))
}

func TestRunShutsDownOnContextCancel(t *testing.T) {
	cfg := memoryConfig()
	cfg.HTTPAddr = freeAddr(t)
	cfg.ShutdownTimeout = 2 * time.Second

	application, err := app.New(context.Background(), cfg, discardLogger())
	if err != nil {
		t.Fatalf("не удалось собрать приложение: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	runErr := make(chan error, 1)

	go func() {
		runErr <- application.Run(ctx)
	}()

	waitForHealth(t, cfg.HTTPAddr)
	cancel()

	select {
	case err := <-runErr:
		if err != nil {
			t.Fatalf("остановка вернула ошибку: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("сервер не остановился за отведённое время")
	}
}

// Приложение обязано подняться с рабочим каталогом сценариев:
// проверка фикстур на старте — часть контракта запуска.
func TestNewServesScenarioCatalog(t *testing.T) {
	application, err := app.New(context.Background(), memoryConfig(), discardLogger())
	if err != nil {
		t.Fatalf("не удалось собрать приложение: %v", err)
	}

	defer application.Close()

	first := httptest.NewRecorder()
	application.Handler().ServeHTTP(first, httptest.NewRequest(http.MethodGet, "/api/v1/scenarios", nil))
	if first.Code != http.StatusUnauthorized {
		t.Fatalf("каталог без входа: статус = %d, ожидался 401", first.Code)
	}
	register := httptest.NewRecorder()
	registerRequest := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", strings.NewReader(
		`{"email":"tester@example.com","password":"password1!","displayName":"Тестер","avatar":"profile"}`))
	for _, cookie := range first.Result().Cookies() {
		registerRequest.AddCookie(cookie)
	}
	application.Handler().ServeHTTP(register, registerRequest)
	if register.Code != http.StatusCreated {
		t.Fatalf("регистрация: статус = %d, тело: %s", register.Code, register.Body.String())
	}

	recorder := httptest.NewRecorder()
	catalogRequest := httptest.NewRequest(http.MethodGet, "/api/v1/scenarios", nil)
	for _, cookie := range append(first.Result().Cookies(), register.Result().Cookies()...) {
		catalogRequest.AddCookie(cookie)
	}
	application.Handler().ServeHTTP(recorder, catalogRequest)

	if recorder.Code != http.StatusOK {
		t.Fatalf("статус = %d, ожидался 200", recorder.Code)
	}

	var body struct {
		Scenarios []struct {
			ID string `json:"id"`
		} `json:"scenarios"`
	}

	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("не удалось разобрать ответ: %v", err)
	}

	if len(body.Scenarios) < 2 {
		t.Fatalf("в каталоге %d сценариев, ожидалось минимум 2", len(body.Scenarios))
	}
}

func TestRunFailsOnBusyAddress(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("не удалось занять порт: %v", err)
	}

	defer func() { _ = listener.Close() }()

	cfg := memoryConfig()
	cfg.HTTPAddr = listener.Addr().String()

	application, err := app.New(context.Background(), cfg, discardLogger())
	if err != nil {
		t.Fatalf("не удалось собрать приложение: %v", err)
	}

	if err := application.Run(context.Background()); err == nil {
		t.Fatal("ожидалась ошибка запуска на занятом адресе")
	}
}

// freeAddr резервирует свободный порт и сразу освобождает его,
// чтобы тест не зависел от фиксированного номера.
func freeAddr(t *testing.T) string {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("не удалось получить свободный порт: %v", err)
	}

	addr := listener.Addr().String()

	if err := listener.Close(); err != nil {
		t.Fatalf("не удалось освободить порт: %v", err)
	}

	return addr
}

func waitForHealth(t *testing.T, addr string) {
	t.Helper()

	deadline := time.Now().Add(3 * time.Second)

	for time.Now().Before(deadline) {
		response, err := http.Get("http://" + addr + "/healthz") //nolint:noctx // короткий локальный запрос в тесте
		if err == nil {
			_ = response.Body.Close()

			if response.StatusCode == http.StatusOK {
				return
			}
		}

		time.Sleep(20 * time.Millisecond)
	}

	t.Fatal("сервер не ответил на /healthz")
}
