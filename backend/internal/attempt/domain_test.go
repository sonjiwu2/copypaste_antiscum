package attempt_test

import (
	"errors"
	"testing"
	"time"

	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/attempt"
	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/scenario"
)

var startMoment = time.Date(2026, time.August, 3, 12, 0, 0, 0, time.UTC)

func validStartParams() attempt.StartParams {
	return attempt.StartParams{
		ID:              "attempt-1",
		ScenarioID:      "buyer-fake-delivery",
		ScenarioVersion: 1,
		StartNodeID:     "greeting",
		RevealedNodeIDs: []scenario.NodeID{"greeting", "channel-decision"},
		CurrentNodeID:   "channel-decision",
		StartedAt:       startMoment,
	}
}

func mustStart(t *testing.T) attempt.Attempt {
	t.Helper()

	started, err := attempt.Start(validStartParams())
	if err != nil {
		t.Fatalf("не удалось создать попытку: %v", err)
	}

	return started
}

func decision(key attempt.IdempotencyKey, delta int, next scenario.NodeID) attempt.Decision {
	return attempt.Decision{
		NodeID:         "channel-decision",
		ChoiceID:       "move-to-messenger",
		ChoiceLabel:    "Перейти в мессенджер",
		IdempotencyKey: key,
		Consequence: scenario.Consequence{
			Severity:    scenario.SeverityDangerous,
			Title:       "Защита потеряна",
			Explanation: "Вне площадки переписка не поможет.",
		},
		Criticality:     scenario.CriticalityHigh,
		RiskTags:        []scenario.RiskTag{"off_platform"},
		SkillEffects:    []scenario.SkillEffect{{Skill: "channel_safety", Delta: -1}},
		ScoreDelta:      delta,
		RevealedNodeIDs: []scenario.NodeID{"delivery-message", "link-decision"},
		ResultingNodeID: next,
	}
}

func TestStartCreatesAttemptInProgress(t *testing.T) {
	started := mustStart(t)

	if started.Status != attempt.StatusInProgress {
		t.Errorf("статус = %q, ожидался in_progress", started.Status)
	}

	if started.Score != attempt.InitialScore {
		t.Errorf("score = %d, ожидался %d", started.Score, attempt.InitialScore)
	}

	if started.Version != 1 {
		t.Errorf("version = %d, ожидалась 1", started.Version)
	}

	if started.CompletedAt != nil {
		t.Error("незавершённая попытка не должна иметь время завершения")
	}

	if len(started.Decisions) != 0 {
		t.Errorf("решений = %d, ожидалось 0", len(started.Decisions))
	}

	if started.ScenarioVersion != 1 {
		t.Errorf("версия сценария = %d, ожидалась 1", started.ScenarioVersion)
	}
}

func TestStartValidatesInput(t *testing.T) {
	testCases := []struct {
		name    string
		mutate  func(params *attempt.StartParams)
		wantErr error
	}{
		{
			name:    "пустой идентификатор попытки",
			mutate:  func(p *attempt.StartParams) { p.ID = "" },
			wantErr: attempt.ErrEmptyAttemptID,
		},
		{
			name:    "пустой идентификатор сценария",
			mutate:  func(p *attempt.StartParams) { p.ScenarioID = "" },
			wantErr: attempt.ErrEmptyScenarioID,
		},
		{
			name:    "неположительная версия сценария",
			mutate:  func(p *attempt.StartParams) { p.ScenarioVersion = 0 },
			wantErr: attempt.ErrInvalidScenarioVersion,
		},
		{
			name:    "пустой текущий узел",
			mutate:  func(p *attempt.StartParams) { p.CurrentNodeID = "" },
			wantErr: attempt.ErrEmptyNode,
		},
		{
			name:    "нулевое время начала",
			mutate:  func(p *attempt.StartParams) { p.StartedAt = time.Time{} },
			wantErr: attempt.ErrEmptyStartTime,
		},
		{
			name:    "ни одного раскрытого узла",
			mutate:  func(p *attempt.StartParams) { p.RevealedNodeIDs = nil },
			wantErr: attempt.ErrNothingRevealed,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			params := validStartParams()
			testCase.mutate(&params)

			if _, err := attempt.Start(params); !errors.Is(err, testCase.wantErr) {
				t.Errorf("ошибка = %v, ожидалась %v", err, testCase.wantErr)
			}
		})
	}
}

func TestStartCanFinishImmediatelyOnTerminalScenario(t *testing.T) {
	params := validStartParams()
	params.Completed = true
	params.Outcome = scenario.OutcomeSafe

	started, err := attempt.Start(params)
	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}

	if started.Status != attempt.StatusCompleted {
		t.Errorf("статус = %q, ожидался completed", started.Status)
	}

	if started.CompletedAt == nil {
		t.Fatal("завершённая попытка должна иметь время завершения")
	}

	if started.Outcome != scenario.OutcomeSafe {
		t.Errorf("итог = %q, ожидался safe", started.Outcome)
	}
}

func TestRecordAppliesDecision(t *testing.T) {
	started := mustStart(t)
	now := startMoment.Add(time.Minute)

	recorded, err := started.Record(decision("key-1", -20, "link-decision"), now)
	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}

	if started.Score != 80 {
		t.Errorf("score = %d, ожидался 80", started.Score)
	}

	if recorded.ScoreAfter != 80 {
		t.Errorf("ScoreAfter = %d, ожидался 80", recorded.ScoreAfter)
	}

	if started.CurrentNodeID != "link-decision" {
		t.Errorf("текущий узел = %q, ожидался link-decision", started.CurrentNodeID)
	}

	if started.AppliedSkillEffects["channel_safety"] != -1 {
		t.Errorf("эффект навыка = %d, ожидался -1", started.AppliedSkillEffects["channel_safety"])
	}

	if len(started.RevealedNodeIDs) != 4 {
		t.Errorf("раскрытых узлов = %d, ожидалось 4", len(started.RevealedNodeIDs))
	}

	if len(started.Decisions) != 1 {
		t.Fatalf("решений = %d, ожидалось 1", len(started.Decisions))
	}

	if !started.Decisions[0].CreatedAt.Equal(now) {
		t.Errorf("время решения = %v, ожидалось %v", started.Decisions[0].CreatedAt, now)
	}

	if started.Status != attempt.StatusInProgress {
		t.Errorf("статус = %q, ожидался in_progress", started.Status)
	}
}

func TestRecordRejectsInvalidTransitions(t *testing.T) {
	testCases := []struct {
		name    string
		prepare func(t *testing.T) (attempt.Attempt, attempt.Decision)
		wantErr error
	}{
		{
			name: "устаревший узел",
			prepare: func(t *testing.T) (attempt.Attempt, attempt.Decision) {
				t.Helper()

				step := decision("key-1", -10, "link-decision")
				step.NodeID = "greeting"

				return mustStart(t), step
			},
			wantErr: attempt.ErrStaleNode,
		},
		{
			name: "завершённая попытка",
			prepare: func(t *testing.T) (attempt.Attempt, attempt.Decision) {
				t.Helper()

				params := validStartParams()
				params.Completed = true
				params.Outcome = scenario.OutcomeUnsafe

				finished, err := attempt.Start(params)
				if err != nil {
					t.Fatalf("неожиданная ошибка: %v", err)
				}

				return finished, decision("key-1", -10, "link-decision")
			},
			wantErr: attempt.ErrAlreadyCompleted,
		},
		{
			name: "пустой ключ повтора",
			prepare: func(t *testing.T) (attempt.Attempt, attempt.Decision) {
				t.Helper()

				return mustStart(t), decision("", -10, "link-decision")
			},
			wantErr: attempt.ErrEmptyIdempotencyKey,
		},
		{
			name: "пустой следующий узел",
			prepare: func(t *testing.T) (attempt.Attempt, attempt.Decision) {
				t.Helper()

				return mustStart(t), decision("key-1", -10, "")
			},
			wantErr: attempt.ErrEmptyNode,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			target, step := testCase.prepare(t)
			scoreBefore := target.Score
			decisionsBefore := len(target.Decisions)

			if _, err := target.Record(step, startMoment.Add(time.Minute)); !errors.Is(err, testCase.wantErr) {
				t.Fatalf("ошибка = %v, ожидалась %v", err, testCase.wantErr)
			}

			if target.Score != scoreBefore {
				t.Errorf("score изменился на %d при отклонённом решении", target.Score-scoreBefore)
			}

			if len(target.Decisions) != decisionsBefore {
				t.Error("отклонённое решение не должно попадать в историю")
			}
		})
	}
}

func TestRecordKeepsScoreInsideBounds(t *testing.T) {
	testCases := []struct {
		name      string
		deltas    []int
		wantScore int
	}{
		{name: "нижняя граница", deltas: []int{-60, -60}, wantScore: attempt.MinScore},
		{name: "верхняя граница", deltas: []int{40}, wantScore: attempt.MaxScore},
		{name: "нулевая дельта", deltas: []int{0}, wantScore: attempt.InitialScore},
		{name: "обычное уменьшение", deltas: []int{-15, -25}, wantScore: 60},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			target := mustStart(t)

			for i, delta := range testCase.deltas {
				step := decision(attempt.IdempotencyKey(string(rune('a'+i))), delta, "channel-decision")

				if _, err := target.Record(step, startMoment.Add(time.Duration(i)*time.Minute)); err != nil {
					t.Fatalf("шаг %d вернул ошибку: %v", i, err)
				}
			}

			if target.Score != testCase.wantScore {
				t.Errorf("score = %d, ожидался %d", target.Score, testCase.wantScore)
			}
		})
	}
}

func TestRecordCompletesAttemptOnce(t *testing.T) {
	target := mustStart(t)
	completionMoment := startMoment.Add(2 * time.Minute)

	final := decision("key-final", -30, "unsafe-ending")
	final.Completed = true
	final.Outcome = scenario.OutcomeUnsafe

	if _, err := target.Record(final, completionMoment); err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}

	if target.Status != attempt.StatusCompleted {
		t.Fatalf("статус = %q, ожидался completed", target.Status)
	}

	if target.CompletedAt == nil || !target.CompletedAt.Equal(completionMoment) {
		t.Fatalf("время завершения = %v, ожидалось %v", target.CompletedAt, completionMoment)
	}

	// Завершённая попытка неизменяема: повторная запись обязана быть отклонена,
	// а время завершения — остаться прежним.
	if _, err := target.Record(decision("key-next", -10, "safe-ending"), completionMoment.Add(time.Minute)); !errors.Is(err, attempt.ErrAlreadyCompleted) {
		t.Fatalf("ошибка = %v, ожидалась ErrAlreadyCompleted", err)
	}

	if !target.CompletedAt.Equal(completionMoment) {
		t.Error("время завершения не должно переустанавливаться")
	}
}

func TestDecisionByIdempotencyKey(t *testing.T) {
	target := mustStart(t)

	if _, found := target.DecisionByIdempotencyKey("key-1"); found {
		t.Error("до записи решение не должно находиться")
	}

	if _, err := target.Record(decision("key-1", -20, "link-decision"), startMoment.Add(time.Minute)); err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}

	found, exists := target.DecisionByIdempotencyKey("key-1")
	if !exists {
		t.Fatal("записанное решение должно находиться по ключу")
	}

	if found.ScoreAfter != 80 {
		t.Errorf("ScoreAfter = %d, ожидался 80", found.ScoreAfter)
	}

	if _, exists := target.DecisionByIdempotencyKey("key-2"); exists {
		t.Error("чужой ключ не должен находиться")
	}
}

// Копия попытки не должна разделять память с оригиналом,
// иначе изменение копии портит сохранённое состояние.
func TestCloneIsIndependent(t *testing.T) {
	target := mustStart(t)

	if _, err := target.Record(decision("key-1", -20, "link-decision"), startMoment.Add(time.Minute)); err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}

	cloned := target.Clone()

	cloned.RevealedNodeIDs[0] = "подменённый-узел"
	cloned.Decisions[0].ChoiceLabel = "подменённая подпись"
	cloned.Decisions[0].RiskTags[0] = "подменённый-тег"
	cloned.AppliedSkillEffects["channel_safety"] = 999

	if target.RevealedNodeIDs[0] == "подменённый-узел" {
		t.Error("раскрытые узлы разделяют память с копией")
	}

	if target.Decisions[0].ChoiceLabel == "подменённая подпись" {
		t.Error("история решений разделяет память с копией")
	}

	if target.Decisions[0].RiskTags[0] == "подменённый-тег" {
		t.Error("метки риска разделяют память с копией")
	}

	if target.AppliedSkillEffects["channel_safety"] == 999 {
		t.Error("эффекты навыков разделяют память с копией")
	}
}

func TestCloneCopiesCompletionTime(t *testing.T) {
	params := validStartParams()
	params.Completed = true
	params.Outcome = scenario.OutcomeSafe

	finished, err := attempt.Start(params)
	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}

	cloned := finished.Clone()
	if cloned.CompletedAt == finished.CompletedAt {
		t.Error("указатель на время завершения должен быть независимым")
	}

	if !cloned.CompletedAt.Equal(*finished.CompletedAt) {
		t.Error("значение времени завершения должно совпадать")
	}
}
