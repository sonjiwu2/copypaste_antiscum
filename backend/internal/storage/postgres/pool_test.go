package postgres_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/config"
	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/storage/postgres"
)

func TestNewPoolRequiresDatabaseURL(t *testing.T) {
	settings := config.Default().Database

	_, err := postgres.NewPool(context.Background(), settings)
	if !errors.Is(err, postgres.ErrEmptyDatabaseURL) {
		t.Fatalf("ошибка = %v, ожидалась ErrEmptyDatabaseURL", err)
	}
}

// Некорректный адрес обязан отвергаться до попытки соединения,
// иначе ошибка конфигурации выглядела бы как недоступность базы.
func TestNewPoolRejectsMalformedURL(t *testing.T) {
	settings := config.Default().Database
	settings.URL = "://not-a-database-url"

	_, err := postgres.NewPool(context.Background(), settings)
	if err == nil {
		t.Fatal("ожидалась ошибка разбора адреса")
	}

	if errors.Is(err, postgres.ErrEmptyDatabaseURL) {
		t.Fatalf("ошибка = %v, ожидалась ошибка разбора", err)
	}
}

// stubPinger заменяет базу в проверках readiness.
type stubPinger struct {
	err     error
	delay   time.Duration
	calls   int
	lastCtx context.Context //nolint:containedctx // тест сохраняет контекст для проверки дедлайна
}

func (p *stubPinger) Ping(ctx context.Context) error {
	p.calls++
	p.lastCtx = ctx

	if p.delay > 0 {
		select {
		case <-time.After(p.delay):
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	return p.err
}

func TestPingSucceeds(t *testing.T) {
	pinger := &stubPinger{}

	if err := postgres.Ping(context.Background(), pinger, time.Second); err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}

	if pinger.calls != 1 {
		t.Errorf("вызовов Ping = %d, ожидался 1", pinger.calls)
	}

	if _, ok := pinger.lastCtx.Deadline(); !ok {
		t.Error("проверка доступности должна выполняться с таймаутом")
	}
}

func TestPingReportsFailure(t *testing.T) {
	pinger := &stubPinger{err: errors.New("база недоступна")}

	err := postgres.Ping(context.Background(), pinger, time.Second)
	if err == nil {
		t.Fatal("ожидалась ошибка недоступной базы")
	}

	if !errors.Is(err, pinger.err) {
		t.Errorf("ошибка = %v, ожидалась исходная причина", err)
	}
}

// Отменённый вызывающим контекст обязан прерывать проверку немедленно.
func TestPingRespectsCanceledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := postgres.Ping(ctx, &stubPinger{delay: time.Second}, time.Minute)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("ошибка = %v, ожидалась context.Canceled", err)
	}
}

// Ключевой инвариант: таймаут базы не имеет права продлевать дедлайн,
// уже назначенный вызывающим. Иначе отменённый клиентом запрос продолжал бы
// занимать соединение.
func TestWithTimeoutNeverExtendsCallerDeadline(t *testing.T) {
	const callerTimeout = 20 * time.Millisecond

	callerCtx, cancelCaller := context.WithTimeout(context.Background(), callerTimeout)
	defer cancelCaller()

	callerDeadline, ok := callerCtx.Deadline()
	if !ok {
		t.Fatal("контекст вызывающего должен иметь дедлайн")
	}

	ctx, cancel := postgres.WithTimeout(callerCtx, time.Hour)
	defer cancel()

	deadline, ok := ctx.Deadline()
	if !ok {
		t.Fatal("дедлайн вызывающего потерян")
	}

	if deadline.After(callerDeadline) {
		t.Errorf("дедлайн = %v, он не должен быть позже %v", deadline, callerDeadline)
	}
}

func TestWithTimeoutAppliesShorterTimeout(t *testing.T) {
	callerCtx, cancelCaller := context.WithTimeout(context.Background(), time.Hour)
	defer cancelCaller()

	callerDeadline, _ := callerCtx.Deadline()

	ctx, cancel := postgres.WithTimeout(callerCtx, time.Second)
	defer cancel()

	deadline, ok := ctx.Deadline()
	if !ok {
		t.Fatal("таймаут операции не применён")
	}

	if !deadline.Before(callerDeadline) {
		t.Errorf("дедлайн = %v, ожидался раньше %v", deadline, callerDeadline)
	}
}

func TestWithTimeoutWithoutTimeoutKeepsContext(t *testing.T) {
	ctx, cancel := postgres.WithTimeout(context.Background(), 0)
	defer cancel()

	if _, ok := ctx.Deadline(); ok {
		t.Error("нулевой таймаут не должен создавать дедлайн")
	}
}
