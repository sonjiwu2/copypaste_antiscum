package app_test

import (
	"context"
	"io"
	"log/slog"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/app"
	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/config"
)

func TestRunShutsDownOnContextCancel(t *testing.T) {
	cfg := config.Default()
	cfg.HTTPAddr = freeAddr(t)
	cfg.ShutdownTimeout = 2 * time.Second

	application := app.New(cfg, slog.New(slog.NewJSONHandler(io.Discard, nil)))

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

func TestRunFailsOnBusyAddress(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("не удалось занять порт: %v", err)
	}

	defer func() { _ = listener.Close() }()

	cfg := config.Default()
	cfg.HTTPAddr = listener.Addr().String()

	application := app.New(cfg, slog.New(slog.NewJSONHandler(io.Discard, nil)))

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
