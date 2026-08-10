//go:build integration

package postgres_test

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/attempt"
	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/config"
	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/scenario"
	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/storage/attempttest"
	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/storage/postgres"
)

// queryTimeout взят с запасом: тесты не проверяют производительность,
// а слишком короткий таймаут сделал бы их нестабильными.
const queryTimeout = 10 * time.Second

// requirePool открывает пул к тестовой базе и готовит чистую схему.
func requirePool(t *testing.T) *pgxpool.Pool {
	t.Helper()

	url, ok := os.LookupEnv(testDatabaseURLEnv)
	if !ok || url == "" {
		t.Skipf("%s не задан: интеграционные тесты пропущены", testDatabaseURLEnv)
	}

	settings := config.Default().Database
	settings.URL = url

	pool, err := postgres.NewPool(context.Background(), settings)
	if err != nil {
		t.Fatalf("не удалось открыть пул: %v", err)
	}

	t.Cleanup(pool.Close)

	return pool
}

// prepareSchema пересоздаёт схему и добавляет версию сценария, на которую
// ссылается контракт: внешний ключ не позволит сохранить попытку без неё.
func prepareSchema(t *testing.T) {
	t.Helper()

	db := requireDatabase(t)
	migrateUp(t, db)
	seedContractScenario(t, db)
	seedContractProfile(t, db)
}

// seedContractProfile создаёт владельца попыток контракта: внешний ключ
// не позволит сохранить попытку без существующего профиля.
func seedContractProfile(t *testing.T, db *sql.DB) {
	t.Helper()

	_, err := db.ExecContext(context.Background(), `
INSERT INTO profiles (id, created_at, updated_at)
VALUES ($1, now(), now())
ON CONFLICT (id) DO NOTHING`, string(attempttest.ProfileID))
	if err != nil {
		t.Fatalf("не удалось подготовить профиль: %v", err)
	}
}

func seedContractScenario(t *testing.T, db *sql.DB) {
	t.Helper()

	_, err := db.ExecContext(context.Background(), `
INSERT INTO scenario_versions (
    scenario_id, version, slug, role, title, description, difficulty,
    estimated_minutes, is_active, content, content_hash, created_at
) VALUES ($1, $2, 'contract-scenario', 'buyer', 'Сценарий контракта', '', 'medium',
    4, true, '{}'::jsonb, 'contract-hash', now())
ON CONFLICT (scenario_id, version) DO NOTHING`,
		string(attempttest.ScenarioID), int(attempttest.ScenarioVersion))
	if err != nil {
		t.Fatalf("не удалось подготовить версию сценария: %v", err)
	}
}

func truncateAttempts(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()

	// CASCADE убирает и решения: они связаны внешним ключом.
	_, err := pool.Exec(context.Background(), "TRUNCATE attempts CASCADE")
	if err != nil {
		t.Fatalf("не удалось очистить попытки: %v", err)
	}
}

// Общий контракт хранилища: PostgreSQL обязан вести себя так же,
// как in-memory адаптер, иначе переключение хранилища изменит поведение API.
func TestAttemptRepositoryContract(t *testing.T) {
	prepareSchema(t)

	pool := requirePool(t)

	attempttest.RunRepositoryContract(t, func(t *testing.T) attempt.Repository {
		truncateAttempts(t, pool)

		return postgres.NewAttemptRepository(pool, queryTimeout)
	})
}

// Попытка не может ссылаться на несуществующую версию сценария:
// иначе её нельзя было бы доиграть после выпуска нового контента.
func TestCreateRejectsUnknownScenarioVersion(t *testing.T) {
	prepareSchema(t)

	pool := requirePool(t)
	truncateAttempts(t, pool)

	repository := postgres.NewAttemptRepository(pool, queryTimeout)

	created := attempttest.NewAttempt(t, "attempt-unknown-version")
	created.ScenarioVersion = 999

	err := repository.Create(context.Background(), created)
	if !postgres.IsIntegrityViolation(err) {
		t.Fatalf("ошибка = %v, ожидалось нарушение целостности", err)
	}
}

// Главный риск транзакции: попытка обновлена, а решение не сохранено.
// Решение с недопустимой критичностью отвергается ограничением схемы,
// и вся транзакция обязана откатиться.
func TestUpdateRollsBackWhenDecisionIsRejected(t *testing.T) {
	prepareSchema(t)

	pool := requirePool(t)
	truncateAttempts(t, pool)

	repository := postgres.NewAttemptRepository(pool, queryTimeout)
	ctx := context.Background()

	created := attempttest.NewAttempt(t, "attempt-rollback")
	if err := repository.Create(ctx, created); err != nil {
		t.Fatalf("не удалось сохранить попытку: %v", err)
	}

	broken := attempttest.FullDecision("key-broken")
	broken.NodeID = created.CurrentNodeID
	broken.Criticality = scenario.Criticality("extreme")

	source := created.Clone()
	if _, err := source.Record(broken, attempttest.StartMoment.Add(time.Minute)); err != nil {
		t.Fatalf("не удалось записать решение: %v", err)
	}

	if _, err := repository.Update(ctx, source); !postgres.IsIntegrityViolation(err) {
		t.Fatalf("ошибка = %v, ожидалось нарушение целостности", err)
	}

	found, err := repository.Get(ctx, created.ID)
	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}

	if found.Version != created.Version {
		t.Errorf("версия = %d, ожидалась %d: обновление должно было откатиться",
			found.Version, created.Version)
	}

	if len(found.Decisions) != 0 {
		t.Errorf("решений = %d, ожидалось 0", len(found.Decisions))
	}

	if found.Score != created.Score {
		t.Errorf("score = %d, ожидался %d", found.Score, created.Score)
	}
}

// Уникальность (attempt_id, idempotency_key) — второй рубеж идемпотентности.
// Она обязана срабатывать даже если сервис пропустил повтор.
func TestUpdateRejectsDuplicateIdempotencyKey(t *testing.T) {
	prepareSchema(t)

	pool := requirePool(t)
	truncateAttempts(t, pool)

	repository := postgres.NewAttemptRepository(pool, queryTimeout)
	ctx := context.Background()

	created := attempttest.NewAttempt(t, "attempt-duplicate-key")
	if err := repository.Create(ctx, created); err != nil {
		t.Fatalf("не удалось сохранить попытку: %v", err)
	}

	source := created.Clone()

	// Домен такую историю не построит: два решения с одним ключом
	// собираются вручную именно для проверки ограничения базы.
	duplicate := attempttest.FullDecision("key-same")
	duplicate.NodeID = created.CurrentNodeID
	duplicate.CreatedAt = attempttest.StartMoment
	duplicate.ScoreAfter = 80
	source.Decisions = []attempt.Decision{duplicate, duplicate}

	_, err := repository.Update(ctx, source)
	if !errors.Is(err, attempt.ErrConcurrentUpdate) {
		t.Fatalf("ошибка = %v, ожидалась ErrConcurrentUpdate", err)
	}

	found, err := repository.Get(ctx, created.ID)
	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}

	if len(found.Decisions) != 0 {
		t.Errorf("решений = %d, ожидалось 0: транзакция должна была откатиться", len(found.Decisions))
	}
}

// Сохранённая история неизменяема: подмена уже записанного решения означает
// дефект приложения, а не конкурентный запрос.
func TestUpdateRejectsHistoryRewrite(t *testing.T) {
	prepareSchema(t)

	pool := requirePool(t)
	truncateAttempts(t, pool)

	repository := postgres.NewAttemptRepository(pool, queryTimeout)
	ctx := context.Background()

	created := attempttest.NewAttempt(t, "attempt-history")
	if err := repository.Create(ctx, created); err != nil {
		t.Fatalf("не удалось сохранить попытку: %v", err)
	}

	source := created.Clone()

	first := attempttest.FullDecision("key-1")
	first.NodeID = created.CurrentNodeID
	first.ResultingNodeID = created.CurrentNodeID

	if _, err := source.Record(first, attempttest.StartMoment.Add(time.Minute)); err != nil {
		t.Fatalf("не удалось записать решение: %v", err)
	}

	saved, err := repository.Update(ctx, source)
	if err != nil {
		t.Fatalf("не удалось сохранить переход: %v", err)
	}

	// Тот же номер решения, но другой ключ повтора.
	rewritten := saved.Clone()
	rewritten.Decisions[0].IdempotencyKey = "key-подменённый"

	if _, err := repository.Update(ctx, rewritten); !errors.Is(err, postgres.ErrHistoryRewrite) {
		t.Fatalf("ошибка = %v, ожидалась ErrHistoryRewrite", err)
	}

	found, err := repository.Get(ctx, created.ID)
	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}

	if found.Decisions[0].IdempotencyKey != "key-1" {
		t.Errorf("ключ решения = %q, ожидался key-1", found.Decisions[0].IdempotencyKey)
	}

	if found.Version != saved.Version {
		t.Errorf("версия = %d, ожидалась %d: обновление должно было откатиться",
			found.Version, saved.Version)
	}
}

// Попытка обязана переживать пересоздание пула: это и есть durability,
// ради которой вводится PostgreSQL.
func TestAttemptSurvivesPoolRecreation(t *testing.T) {
	prepareSchema(t)

	pool := requirePool(t)
	truncateAttempts(t, pool)

	ctx := context.Background()

	created := attempttest.NewAttempt(t, "attempt-durable")
	source := created.Clone()

	decision := attempttest.FullDecision("key-durable")
	decision.NodeID = created.CurrentNodeID
	decision.ResultingNodeID = "link-decision"

	if _, err := source.Record(decision, attempttest.StartMoment.Add(time.Minute)); err != nil {
		t.Fatalf("не удалось записать решение: %v", err)
	}

	writer := postgres.NewAttemptRepository(pool, queryTimeout)

	if err := writer.Create(ctx, created); err != nil {
		t.Fatalf("не удалось сохранить попытку: %v", err)
	}

	if _, err := writer.Update(ctx, source); err != nil {
		t.Fatalf("не удалось сохранить переход: %v", err)
	}

	// Новый пул имитирует перезапуск процесса: состояние живёт в базе,
	// а не в памяти адаптера.
	settings := config.Default().Database
	settings.URL = os.Getenv(testDatabaseURLEnv)

	restarted, err := postgres.NewPool(ctx, settings)
	if err != nil {
		t.Fatalf("не удалось открыть новый пул: %v", err)
	}

	defer restarted.Close()

	found, err := postgres.NewAttemptRepository(restarted, queryTimeout).Get(ctx, created.ID)
	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}

	if found.CurrentNodeID != "link-decision" {
		t.Errorf("текущий узел = %q, ожидался link-decision", found.CurrentNodeID)
	}

	if found.Score != source.Score {
		t.Errorf("score = %d, ожидался %d", found.Score, source.Score)
	}

	if len(found.Decisions) != 1 {
		t.Fatalf("решений = %d, ожидалось 1", len(found.Decisions))
	}

	if found.Decisions[0].Consequence != decision.Consequence {
		t.Errorf("последствие = %+v, ожидалось %+v", found.Decisions[0].Consequence, decision.Consequence)
	}

	if found.AppliedSkillEffects["verification_discipline"] != -2 {
		t.Errorf("накопленный навык = %d, ожидался -2",
			found.AppliedSkillEffects["verification_discipline"])
	}
}

// Истёкший таймаут запроса обязан возвращать ошибку контекста,
// а не блокировать вызывающего.
func TestRepositoryRespectsQueryTimeout(t *testing.T) {
	prepareSchema(t)

	pool := requirePool(t)
	truncateAttempts(t, pool)

	repository := postgres.NewAttemptRepository(pool, time.Nanosecond)

	err := repository.Create(context.Background(), attempttest.NewAttempt(t, "attempt-timeout"))
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("ошибка = %v, ожидалась context.DeadlineExceeded", err)
	}

	if _, err := repository.Get(context.Background(), "attempt-timeout"); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("Get: ошибка = %v, ожидалась context.DeadlineExceeded", err)
	}
}
