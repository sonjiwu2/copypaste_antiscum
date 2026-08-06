// Package attempttest содержит общий контракт хранилища попыток.
//
// Один и тот же набор проверок запускается против in-memory и PostgreSQL
// адаптеров: подмена хранилища не должна менять поведение сервиса, а
// расхождение семантики — самый дорогой класс ошибок при переезде на базу.
//
// Два различия допущены осознанно и в контракт не входят:
//
//   - уникальность ключа повтора PostgreSQL обеспечивает индексом, а in-memory
//     адаптер — нет. Это второй рубеж защиты, а не замена проверке в сервисе;
//   - пустые коллекции in-memory адаптер возвращает как nil, PostgreSQL — как
//     пустые срезы. Контракт требует нулевой длины, а не конкретного
//     представления: проекции попытки одинаково работают с обоими.
package attempttest

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/attempt"
	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/profile"
	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/scenario"
)

// Сценарий, к которому привязаны попытки контракта.
//
// Значения фиксированы, чтобы PostgreSQL-адаптер мог заранее создать нужную
// версию сценария: внешний ключ не позволит сохранить попытку без неё.
// Идентификатор намеренно не совпадает ни с одним реальным сценарием:
// иначе тестовая версия с фиктивным содержимым столкнулась бы с настоящей.
const (
	ScenarioID      = scenario.ID("contract-scenario")
	ScenarioVersion = scenario.Version(1)
)

// ProfileID — владелец попыток контракта. Внешний ключ требует, чтобы такой
// профиль существовал, поэтому значение фиксировано.
const ProfileID = profile.ID("contract-profile-owner")

// StartMoment — точка отсчёта времени в контракте.
var StartMoment = time.Date(2026, time.August, 3, 12, 0, 0, 0, time.UTC)

// Factory создаёт пустое хранилище для одной проверки.
type Factory func(t *testing.T) attempt.Repository

// RunRepositoryContract прогоняет все обязательные проверки хранилища.
func RunRepositoryContract(t *testing.T, newRepository Factory) {
	t.Helper()

	checks := []struct {
		name string
		run  func(t *testing.T, newRepository Factory)
	}{
		{"создание и чтение", checkCreateAndGet},
		{"повторное создание", checkDuplicateCreate},
		{"чтение отсутствующей попытки", checkGetUnknown},
		{"обновление отсутствующей попытки", checkUpdateUnknown},
		{"обновление повышает версию", checkUpdateRaisesVersion},
		{"устаревшая версия", checkStaleVersion},
		{"отменённый контекст", checkCanceledContext},
		{"хранилище не делится изменяемым состоянием", checkNoMutationLeak},
		{"вход не изменяется", checkInputIsNotMutated},
		{"решение восстанавливается полностью", checkDecisionRoundtrip},
		{"порядок решений сохраняется", checkDecisionOrder},
		{"завершённая попытка", checkCompletedAttempt},
		{"пустые коллекции", checkEmptyCollections},
		{"история решений только дополняется", checkHistoryIsAppendOnly},
		{"параллельные обновления", checkConcurrentUpdates},
		{"владение попыткой", checkOwnership},
	}

	for _, check := range checks {
		t.Run(check.name, func(t *testing.T) {
			check.run(t, newRepository)
		})
	}
}

// NewAttempt создаёт попытку в начальном состоянии.
func NewAttempt(t *testing.T, id attempt.ID) attempt.Attempt {
	t.Helper()

	created, err := attempt.Start(attempt.StartParams{
		ID:              id,
		ProfileID:       ProfileID,
		ScenarioID:      ScenarioID,
		ScenarioVersion: ScenarioVersion,
		StartNodeID:     "greeting",
		RevealedNodeIDs: []scenario.NodeID{"greeting", "channel-decision"},
		CurrentNodeID:   "channel-decision",
		StartedAt:       StartMoment,
	})
	if err != nil {
		t.Fatalf("не удалось создать попытку: %v", err)
	}

	return created
}

// FullDecision возвращает решение со всеми заполненными полями.
// Контракт проверяет именно его: частично заполненное решение не поймало бы
// потерю вложенных значений при сериализации.
func FullDecision(key attempt.IdempotencyKey) attempt.Decision {
	return attempt.Decision{
		NodeID:         "channel-decision",
		ChoiceID:       "move-to-messenger",
		ChoiceLabel:    "Перейти в мессенджер",
		IdempotencyKey: key,
		Consequence: scenario.Consequence{
			Severity:      scenario.SeverityDangerous,
			Title:         "Защита площадки потеряна",
			Explanation:   "Вне сервиса переписка не подтвердит договорённости при споре.",
			RealWorldRule: "Просьба уйти в мессенджер — типичный первый шаг обмана.",
		},
		Criticality: scenario.CriticalityHigh,
		RiskTags:    []scenario.RiskTag{"off_platform", "urgency"},
		SkillEffects: []scenario.SkillEffect{
			{Skill: "channel_safety", Delta: -1},
			{Skill: "verification_discipline", Delta: -2},
		},
		ScoreDelta:      -20,
		RevealedNodeIDs: []scenario.NodeID{"delivery-message", "link-decision"},
		ResultingNodeID: "link-decision",
	}
}

func storedAttempt(t *testing.T, newRepository Factory) (attempt.Repository, attempt.Attempt) {
	t.Helper()

	repository := newRepository(t)
	created := NewAttempt(t, "attempt-1")

	if err := repository.Create(context.Background(), created); err != nil {
		t.Fatalf("не удалось сохранить попытку: %v", err)
	}

	return repository, created
}

func checkCreateAndGet(t *testing.T, newRepository Factory) {
	repository, created := storedAttempt(t, newRepository)

	found, err := repository.Get(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}

	assertSameAttempt(t, found, created)
}

func checkDuplicateCreate(t *testing.T, newRepository Factory) {
	repository, created := storedAttempt(t, newRepository)

	if err := repository.Create(context.Background(), created); !errors.Is(err, attempt.ErrDuplicateAttempt) {
		t.Errorf("ошибка = %v, ожидалась ErrDuplicateAttempt", err)
	}
}

func checkGetUnknown(t *testing.T, newRepository Factory) {
	repository := newRepository(t)

	if _, err := repository.Get(context.Background(), "unknown"); !errors.Is(err, attempt.ErrNotFound) {
		t.Errorf("ошибка = %v, ожидалась ErrNotFound", err)
	}
}

func checkUpdateUnknown(t *testing.T, newRepository Factory) {
	repository := newRepository(t)

	_, err := repository.Update(context.Background(), NewAttempt(t, "ghost"))
	if !errors.Is(err, attempt.ErrNotFound) {
		t.Errorf("ошибка = %v, ожидалась ErrNotFound", err)
	}
}

func checkUpdateRaisesVersion(t *testing.T, newRepository Factory) {
	repository, created := storedAttempt(t, newRepository)

	next := created.Clone()
	next.CurrentNodeID = "link-decision"

	saved, err := repository.Update(context.Background(), next)
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

	if found.Version != saved.Version {
		t.Errorf("сохранённая версия = %d, возвращена %d", found.Version, saved.Version)
	}
}

// Обновление с устаревшей версией означает, что попытку изменил другой запрос.
func checkStaleVersion(t *testing.T, newRepository Factory) {
	repository, created := storedAttempt(t, newRepository)

	if _, err := repository.Update(context.Background(), created.Clone()); err != nil {
		t.Fatalf("первое обновление вернуло ошибку: %v", err)
	}

	_, err := repository.Update(context.Background(), created.Clone())
	if !errors.Is(err, attempt.ErrConcurrentUpdate) {
		t.Errorf("ошибка = %v, ожидалась ErrConcurrentUpdate", err)
	}
}

func checkCanceledContext(t *testing.T, newRepository Factory) {
	repository, created := storedAttempt(t, newRepository)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if err := repository.Create(ctx, NewAttempt(t, "attempt-2")); !errors.Is(err, context.Canceled) {
		t.Errorf("Create: ошибка = %v, ожидалась context.Canceled", err)
	}

	if _, err := repository.Get(ctx, created.ID); !errors.Is(err, context.Canceled) {
		t.Errorf("Get: ошибка = %v, ожидалась context.Canceled", err)
	}

	if _, err := repository.Update(ctx, created.Clone()); !errors.Is(err, context.Canceled) {
		t.Errorf("Update: ошибка = %v, ожидалась context.Canceled", err)
	}
}

// Изменение полученной попытки не должно доходить до хранилища.
func checkNoMutationLeak(t *testing.T, newRepository Factory) {
	repository, created := storedAttempt(t, newRepository)

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

// Хранилище не имеет права менять переданный агрегат: сервис использует его
// дальше и рассчитывает на неизменность своей копии.
func checkInputIsNotMutated(t *testing.T, newRepository Factory) {
	repository, created := storedAttempt(t, newRepository)

	source := created.Clone()
	recordedDecision(t, &source, "key-1")

	before := source.Clone()

	if _, err := repository.Update(context.Background(), source); err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}

	if source.Version != before.Version {
		t.Errorf("версия входного агрегата изменена: %d, было %d", source.Version, before.Version)
	}

	if len(source.Decisions) != len(before.Decisions) {
		t.Errorf("решения входного агрегата изменены: %d, было %d",
			len(source.Decisions), len(before.Decisions))
	}
}

func checkDecisionRoundtrip(t *testing.T, newRepository Factory) {
	repository, created := storedAttempt(t, newRepository)

	source := created.Clone()
	recorded := recordedDecision(t, &source, "key-1")

	if _, err := repository.Update(context.Background(), source); err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}

	found, err := repository.Get(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}

	if len(found.Decisions) != 1 {
		t.Fatalf("решений = %d, ожидалось 1", len(found.Decisions))
	}

	assertSameDecision(t, found.Decisions[0], recorded)

	if found.Score != recorded.ScoreAfter {
		t.Errorf("score попытки = %d, ожидался %d", found.Score, recorded.ScoreAfter)
	}

	if found.AppliedSkillEffects["channel_safety"] != -1 {
		t.Errorf("накопленный навык channel_safety = %d, ожидался -1",
			found.AppliedSkillEffects["channel_safety"])
	}
}

func checkDecisionOrder(t *testing.T, newRepository Factory) {
	repository, created := storedAttempt(t, newRepository)

	source := created.Clone()
	keys := []attempt.IdempotencyKey{"key-1", "key-2", "key-3"}

	for _, key := range keys {
		recordedDecision(t, &source, key)

		saved, err := repository.Update(context.Background(), source)
		if err != nil {
			t.Fatalf("обновление с ключом %q вернуло ошибку: %v", key, err)
		}

		source.Version = saved.Version
	}

	found, err := repository.Get(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}

	if len(found.Decisions) != len(keys) {
		t.Fatalf("решений = %d, ожидалось %d", len(found.Decisions), len(keys))
	}

	for i, key := range keys {
		if found.Decisions[i].IdempotencyKey != key {
			t.Errorf("решение %d имеет ключ %q, ожидался %q", i, found.Decisions[i].IdempotencyKey, key)
		}
	}
}

func checkCompletedAttempt(t *testing.T, newRepository Factory) {
	repository, created := storedAttempt(t, newRepository)

	source := created.Clone()

	final := FullDecision("key-final")
	final.ResultingNodeID = "unsafe-ending"
	final.RevealedNodeIDs = []scenario.NodeID{"unsafe-ending"}
	final.Completed = true
	final.Outcome = scenario.OutcomeUnsafe

	if _, err := source.Record(final, StartMoment.Add(time.Minute)); err != nil {
		t.Fatalf("не удалось записать решение: %v", err)
	}

	if _, err := repository.Update(context.Background(), source); err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}

	found, err := repository.Get(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}

	if found.Status != attempt.StatusCompleted {
		t.Errorf("статус = %q, ожидался completed", found.Status)
	}

	if found.Outcome != scenario.OutcomeUnsafe {
		t.Errorf("итог = %q, ожидался unsafe", found.Outcome)
	}

	if found.CompletedAt == nil {
		t.Fatal("завершённая попытка должна иметь время завершения")
	}

	if !found.CompletedAt.Equal(*source.CompletedAt) {
		t.Errorf("время завершения = %v, ожидалось %v", found.CompletedAt, source.CompletedAt)
	}

	if !found.Decisions[0].Completed || found.Decisions[0].Outcome != scenario.OutcomeUnsafe {
		t.Errorf("решение = %+v, ожидалось завершающее с итогом unsafe", found.Decisions[0])
	}
}

// Незаполненные коллекции обязаны возвращаться пригодными к чтению.
// Хранилища различаются в том, отдают ли они nil или пустой срез, поэтому
// контракт требует именно нулевой длины, а не конкретного представления.
func checkEmptyCollections(t *testing.T, newRepository Factory) {
	repository, created := storedAttempt(t, newRepository)

	source := created.Clone()

	sparse := FullDecision("key-sparse")
	sparse.RiskTags = nil
	sparse.SkillEffects = nil

	if _, err := source.Record(sparse, StartMoment.Add(time.Minute)); err != nil {
		t.Fatalf("не удалось записать решение: %v", err)
	}

	if _, err := repository.Update(context.Background(), source); err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}

	found, err := repository.Get(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}

	if found.AppliedSkillEffects == nil {
		t.Error("накопленные эффекты навыков должны читаться без проверки на nil")
	}

	if found.Decisions == nil {
		t.Fatal("решения должны читаться без проверки на nil")
	}

	if len(found.Decisions[0].RiskTags) != 0 {
		t.Errorf("метки риска = %v, ожидалось пусто", found.Decisions[0].RiskTags)
	}

	if len(found.Decisions[0].SkillEffects) != 0 {
		t.Errorf("эффекты навыков = %v, ожидалось пусто", found.Decisions[0].SkillEffects)
	}
}

// Уже сохранённые решения не переписываются: повторное обновление добавляет
// только новый суффикс истории.
func checkHistoryIsAppendOnly(t *testing.T, newRepository Factory) {
	repository, created := storedAttempt(t, newRepository)

	source := created.Clone()
	recordedDecision(t, &source, "key-1")

	saved, err := repository.Update(context.Background(), source)
	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}

	source.Version = saved.Version
	recordedDecision(t, &source, "key-2")

	if _, err := repository.Update(context.Background(), source); err != nil {
		t.Fatalf("второе обновление вернуло ошибку: %v", err)
	}

	found, err := repository.Get(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}

	if len(found.Decisions) != 2 {
		t.Fatalf("решений = %d, ожидалось 2", len(found.Decisions))
	}

	if found.Decisions[0].IdempotencyKey != "key-1" || found.Decisions[1].IdempotencyKey != "key-2" {
		t.Errorf("история = %q, %q, ожидалась key-1, key-2",
			found.Decisions[0].IdempotencyKey, found.Decisions[1].IdempotencyKey)
	}
}

// Ключевой инвариант конкурентности: несколько запросов, стартовавших от
// одного состояния попытки, не могут примениться все. Проходит ровно один,
// остальные получают конфликт вместо тихой перезаписи чужого перехода.
func checkConcurrentUpdates(t *testing.T, newRepository Factory) {
	const writers = 8

	repository, created := storedAttempt(t, newRepository)

	loaded, err := repository.Get(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}

	var (
		waitGroup sync.WaitGroup
		mu        sync.Mutex
		accepted  int
		conflicts int
		failures  []error
	)

	for i := range writers {
		waitGroup.Add(1)

		go func() {
			defer waitGroup.Done()

			candidate := loaded.Clone()
			recorded := FullDecision(attempt.IdempotencyKey("key-" + string(rune('a'+i))))
			candidate.Decisions = append(candidate.Decisions, recorded)
			candidate.CurrentNodeID = recorded.ResultingNodeID

			_, err := repository.Update(context.Background(), candidate)

			mu.Lock()
			defer mu.Unlock()

			switch {
			case err == nil:
				accepted++
			case errors.Is(err, attempt.ErrConcurrentUpdate):
				conflicts++
			default:
				failures = append(failures, err)
			}
		}()
	}

	waitGroup.Wait()

	for _, err := range failures {
		t.Errorf("неожиданная ошибка: %v", err)
	}

	if accepted != 1 {
		t.Fatalf("принято переходов = %d, ожидался ровно 1 (конфликтов %d)", accepted, conflicts)
	}

	final, err := repository.Get(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}

	if final.Version != loaded.Version+1 {
		t.Errorf("версия = %d, ожидалась %d", final.Version, loaded.Version+1)
	}

	if len(final.Decisions) != 1 {
		t.Errorf("решений = %d, эффекты применены больше одного раза", len(final.Decisions))
	}
}

// Владение проверяется доменом, а не хранилищем: репозиторий обязан лишь
// точно восстановить владельца, чтобы сервис мог принять решение.
func checkOwnership(t *testing.T, newRepository Factory) {
	repository, created := storedAttempt(t, newRepository)

	found, err := repository.Get(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}

	if err := found.EnsureOwnedBy(ProfileID); err != nil {
		t.Errorf("владелец не признан своим: %v", err)
	}

	if err := found.EnsureOwnedBy("someone-else"); !errors.Is(err, attempt.ErrForbidden) {
		t.Errorf("ошибка = %v, ожидалась ErrForbidden", err)
	}
}

// recordedDecision применяет решение к попытке доменным способом,
// чтобы проверялся тот же путь, которым пользуется сервис.
func recordedDecision(t *testing.T, source *attempt.Attempt, key attempt.IdempotencyKey) attempt.Decision {
	t.Helper()

	decision := FullDecision(key)
	decision.NodeID = source.CurrentNodeID
	decision.ResultingNodeID = source.CurrentNodeID

	recorded, err := source.Record(decision, StartMoment.Add(time.Minute))
	if err != nil {
		t.Fatalf("не удалось записать решение: %v", err)
	}

	return recorded
}

func assertSameAttempt(t *testing.T, found, want attempt.Attempt) {
	t.Helper()

	if found.ID != want.ID {
		t.Errorf("идентификатор = %q, ожидался %q", found.ID, want.ID)
	}

	// Владелец обязан восстанавливаться: без него прогресс попадёт не тому.
	if found.ProfileID != want.ProfileID {
		t.Errorf("владелец = %q, ожидался %q", found.ProfileID, want.ProfileID)
	}

	if found.ScenarioID != want.ScenarioID || found.ScenarioVersion != want.ScenarioVersion {
		t.Errorf("сценарий = %q/%d, ожидался %q/%d",
			found.ScenarioID, found.ScenarioVersion, want.ScenarioID, want.ScenarioVersion)
	}

	if found.CurrentNodeID != want.CurrentNodeID {
		t.Errorf("текущий узел = %q, ожидался %q", found.CurrentNodeID, want.CurrentNodeID)
	}

	if found.Status != want.Status || found.Score != want.Score || found.Outcome != want.Outcome {
		t.Errorf("состояние = %q/%d/%q, ожидалось %q/%d/%q",
			found.Status, found.Score, found.Outcome, want.Status, want.Score, want.Outcome)
	}

	if !found.StartedAt.Equal(want.StartedAt) || !found.UpdatedAt.Equal(want.UpdatedAt) {
		t.Errorf("метки времени = %v/%v, ожидались %v/%v",
			found.StartedAt, found.UpdatedAt, want.StartedAt, want.UpdatedAt)
	}

	if found.Version != want.Version {
		t.Errorf("версия = %d, ожидалась %d", found.Version, want.Version)
	}

	assertSameStrings(t, "раскрытые узлы", found.RevealedNodeIDs, want.RevealedNodeIDs)

	if len(found.Decisions) != len(want.Decisions) {
		t.Errorf("решений = %d, ожидалось %d", len(found.Decisions), len(want.Decisions))
	}
}

func assertSameDecision(t *testing.T, found, want attempt.Decision) {
	t.Helper()

	if found.NodeID != want.NodeID || found.ChoiceID != want.ChoiceID {
		t.Errorf("выбор = %q/%q, ожидался %q/%q",
			found.NodeID, found.ChoiceID, want.NodeID, want.ChoiceID)
	}

	if found.ChoiceLabel != want.ChoiceLabel {
		t.Errorf("подпись = %q, ожидалась %q", found.ChoiceLabel, want.ChoiceLabel)
	}

	if found.IdempotencyKey != want.IdempotencyKey {
		t.Errorf("ключ повтора = %q, ожидался %q", found.IdempotencyKey, want.IdempotencyKey)
	}

	if found.Consequence != want.Consequence {
		t.Errorf("последствие = %+v, ожидалось %+v", found.Consequence, want.Consequence)
	}

	if found.Criticality != want.Criticality {
		t.Errorf("критичность = %q, ожидалась %q", found.Criticality, want.Criticality)
	}

	assertSameStrings(t, "метки риска", found.RiskTags, want.RiskTags)
	assertSameStrings(t, "раскрытые узлы решения", found.RevealedNodeIDs, want.RevealedNodeIDs)

	if len(found.SkillEffects) != len(want.SkillEffects) {
		t.Fatalf("эффектов навыков = %d, ожидалось %d", len(found.SkillEffects), len(want.SkillEffects))
	}

	for i, effect := range want.SkillEffects {
		if found.SkillEffects[i] != effect {
			t.Errorf("эффект навыка %d = %+v, ожидался %+v", i, found.SkillEffects[i], effect)
		}
	}

	if found.ScoreDelta != want.ScoreDelta || found.ScoreAfter != want.ScoreAfter {
		t.Errorf("результат = %d/%d, ожидался %d/%d",
			found.ScoreDelta, found.ScoreAfter, want.ScoreDelta, want.ScoreAfter)
	}

	if !found.CreatedAt.Equal(want.CreatedAt) {
		t.Errorf("время решения = %v, ожидалось %v", found.CreatedAt, want.CreatedAt)
	}

	if found.ResultingNodeID != want.ResultingNodeID {
		t.Errorf("итоговый узел = %q, ожидался %q", found.ResultingNodeID, want.ResultingNodeID)
	}

	if found.Completed != want.Completed || found.Outcome != want.Outcome {
		t.Errorf("завершение = %v/%q, ожидалось %v/%q",
			found.Completed, found.Outcome, want.Completed, want.Outcome)
	}
}

func assertSameStrings[T ~string](t *testing.T, name string, found, want []T) {
	t.Helper()

	if len(found) != len(want) {
		t.Errorf("%s: %v, ожидалось %v", name, found, want)

		return
	}

	for i, value := range want {
		if found[i] != value {
			t.Errorf("%s[%d] = %q, ожидалось %q", name, i, found[i], value)
		}
	}
}
