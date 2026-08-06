package memory_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/attempt"
	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/scenario"
	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/storage/memory"
)

var attemptStartMoment = time.Date(2026, time.August, 3, 12, 0, 0, 0, time.UTC)

func newAttempt(t *testing.T, id attempt.ID) attempt.Attempt {
	t.Helper()

	created, err := attempt.Start(attempt.StartParams{
		ID:              id,
		ProfileID:       "memory-test-profile",
		ScenarioID:      "buyer-fake-delivery",
		ScenarioVersion: 1,
		StartNodeID:     "greeting",
		RevealedNodeIDs: []scenario.NodeID{"greeting", "channel-decision"},
		CurrentNodeID:   "channel-decision",
		StartedAt:       attemptStartMoment,
	})
	if err != nil {
		t.Fatalf("не удалось создать попытку: %v", err)
	}

	return created
}

func storedAttempt(t *testing.T) (*memory.AttemptRepository, attempt.Attempt) {
	t.Helper()

	repository := memory.NewAttemptRepository()
	created := newAttempt(t, "attempt-1")

	if err := repository.Create(context.Background(), created); err != nil {
		t.Fatalf("не удалось сохранить попытку: %v", err)
	}

	return repository, created
}

func TestAttemptRepositoryCreateAndGet(t *testing.T) {
	repository, created := storedAttempt(t)

	found, err := repository.Get(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}

	if found.CurrentNodeID != created.CurrentNodeID {
		t.Errorf("текущий узел = %q, ожидался %q", found.CurrentNodeID, created.CurrentNodeID)
	}

	if found.Score != attempt.InitialScore {
		t.Errorf("score = %d, ожидался %d", found.Score, attempt.InitialScore)
	}
}

func TestAttemptRepositoryRejectsDuplicateCreate(t *testing.T) {
	repository, created := storedAttempt(t)

	if err := repository.Create(context.Background(), created); !errors.Is(err, attempt.ErrDuplicateAttempt) {
		t.Errorf("ошибка = %v, ожидалась ErrDuplicateAttempt", err)
	}
}

func TestAttemptRepositoryGetUnknown(t *testing.T) {
	repository := memory.NewAttemptRepository()

	if _, err := repository.Get(context.Background(), "unknown"); !errors.Is(err, attempt.ErrNotFound) {
		t.Errorf("ошибка = %v, ожидалась ErrNotFound", err)
	}
}

func TestAttemptRepositoryUpdateRaisesVersion(t *testing.T) {
	repository, created := storedAttempt(t)

	created.CurrentNodeID = "link-decision"

	saved, err := repository.Update(context.Background(), created)
	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}

	if saved.Version != created.Version+1 {
		t.Errorf("версия = %d, ожидалась %d", saved.Version, created.Version+1)
	}

	found, err := repository.Get(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}

	if found.CurrentNodeID != "link-decision" {
		t.Errorf("текущий узел = %q, ожидался link-decision", found.CurrentNodeID)
	}
}

// Обновление с устаревшей версией означает, что попытку изменил другой запрос.
func TestAttemptRepositoryRejectsStaleVersion(t *testing.T) {
	repository, created := storedAttempt(t)

	if _, err := repository.Update(context.Background(), created); err != nil {
		t.Fatalf("первое обновление вернуло ошибку: %v", err)
	}

	if _, err := repository.Update(context.Background(), created); !errors.Is(err, attempt.ErrConcurrentUpdate) {
		t.Errorf("ошибка = %v, ожидалась ErrConcurrentUpdate", err)
	}
}

func TestAttemptRepositoryUpdateUnknown(t *testing.T) {
	repository := memory.NewAttemptRepository()

	if _, err := repository.Update(context.Background(), newAttempt(t, "ghost")); !errors.Is(err, attempt.ErrNotFound) {
		t.Errorf("ошибка = %v, ожидалась ErrNotFound", err)
	}
}

func TestAttemptRepositoryRespectsCanceledContext(t *testing.T) {
	repository, created := storedAttempt(t)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if err := repository.Create(ctx, newAttempt(t, "attempt-2")); !errors.Is(err, context.Canceled) {
		t.Errorf("Create: ошибка = %v, ожидалась context.Canceled", err)
	}

	if _, err := repository.Get(ctx, created.ID); !errors.Is(err, context.Canceled) {
		t.Errorf("Get: ошибка = %v, ожидалась context.Canceled", err)
	}

	if _, err := repository.Update(ctx, created); !errors.Is(err, context.Canceled) {
		t.Errorf("Update: ошибка = %v, ожидалась context.Canceled", err)
	}
}

// Изменение полученной попытки не должно доходить до хранилища.
func TestAttemptRepositoryDoesNotLeakMutableState(t *testing.T) {
	repository, created := storedAttempt(t)

	found, err := repository.Get(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}

	found.RevealedNodeIDs[0] = "подменённый-узел"
	found.AppliedSkillEffects["channel_safety"] = 999
	found.Score = 0

	fresh, err := repository.Get(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}

	if fresh.RevealedNodeIDs[0] != "greeting" {
		t.Errorf("узел в хранилище = %q, ожидался greeting", fresh.RevealedNodeIDs[0])
	}

	if fresh.AppliedSkillEffects["channel_safety"] != 0 {
		t.Errorf("эффект навыка в хранилище = %d, ожидался 0", fresh.AppliedSkillEffects["channel_safety"])
	}

	if fresh.Score != attempt.InitialScore {
		t.Errorf("score в хранилище = %d, ожидался %d", fresh.Score, attempt.InitialScore)
	}
}

// Ключевой инвариант конкурентности: несколько запросов, стартовавших от
// одного и того же состояния попытки, не могут примениться все. Проходит
// ровно один, остальные получают конфликт вместо тихой перезаписи чужого
// перехода.
func TestAttemptRepositoryAppliesSingleTransitionFromSameState(t *testing.T) {
	const writers = 20

	repository, created := storedAttempt(t)

	// Снимок делается один раз: так воспроизводится ситуация, когда параллельные
	// запросы прочитали попытку до того, как любой из них успел её изменить.
	loaded, err := repository.Get(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}

	var (
		waitGroup sync.WaitGroup
		mu        sync.Mutex
		accepted  int
		conflicts int
	)

	for i := range writers {
		waitGroup.Add(1)

		go func() {
			defer waitGroup.Done()

			candidate := loaded.Clone()
			candidate.CurrentNodeID = scenario.NodeID("node-" + string(rune('a'+i)))

			_, err := repository.Update(context.Background(), candidate)

			mu.Lock()
			defer mu.Unlock()

			switch {
			case err == nil:
				accepted++
			case errors.Is(err, attempt.ErrConcurrentUpdate):
				conflicts++
			default:
				t.Errorf("неожиданная ошибка: %v", err)
			}
		}()
	}

	waitGroup.Wait()

	if accepted != 1 {
		t.Fatalf("принято переходов = %d, ожидался ровно 1 (конфликтов %d)", accepted, conflicts)
	}

	if conflicts != writers-1 {
		t.Fatalf("конфликтов = %d, ожидалось %d", conflicts, writers-1)
	}

	final, err := repository.Get(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}

	if final.Version != loaded.Version+1 {
		t.Errorf("версия = %d, ожидалась %d", final.Version, loaded.Version+1)
	}
}
