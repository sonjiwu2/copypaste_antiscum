package progress_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/attempt"
	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/profile"
	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/progress"
	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/scenario"
)

const owner = profile.ID("progress-owner-profile")

var base = time.Date(2026, time.August, 3, 12, 0, 0, 0, time.UTC)

// stubRepository отдаёт заранее подготовленные факты.
type stubRepository struct {
	facts   progress.Facts
	failure error
}

func (s stubRepository) Facts(_ context.Context, _ profile.ID) (progress.Facts, error) {
	if s.failure != nil {
		return progress.Facts{}, s.failure
	}

	return s.facts, nil
}

// stubCatalog отдаёт публичные метаданные сценариев.
type stubCatalog struct {
	scenarios []scenario.Metadata
	failure   error
}

func (s stubCatalog) List(_ context.Context, _ scenario.Role) ([]scenario.Metadata, error) {
	if s.failure != nil {
		return nil, s.failure
	}

	return s.scenarios, nil
}

func catalog() stubCatalog {
	return stubCatalog{scenarios: []scenario.Metadata{
		{ID: "buyer-one", Version: 1, Role: scenario.RoleBuyer, Title: "Покупатель 1",
			Difficulty: scenario.DifficultyEasy, EstimatedMinutes: 4},
		{ID: "buyer-two", Version: 1, Role: scenario.RoleBuyer, Title: "Покупатель 2",
			Difficulty: scenario.DifficultyMedium, EstimatedMinutes: 5},
		{ID: "seller-one", Version: 1, Role: scenario.RoleSeller, Title: "Продавец 1",
			Difficulty: scenario.DifficultyHard, EstimatedMinutes: 6},
	}}
}

func newService(facts progress.Facts) *progress.Service {
	return progress.NewService(stubRepository{facts: facts}, catalog())
}

// completed собирает завершённую попытку.
func completed(
	id attempt.ID,
	scenarioID scenario.ID,
	score int,
	minutes int,
	decisions ...progress.DecisionFact,
) progress.CompletedAttempt {
	completedAt := base.Add(time.Duration(minutes) * time.Minute)

	return progress.CompletedAttempt{
		AttemptID:   id,
		ScenarioID:  scenarioID,
		Score:       score,
		Outcome:     scenario.OutcomeSafe,
		StartedAt:   completedAt.Add(-5 * time.Minute),
		CompletedAt: completedAt,
		Decisions:   decisions,
	}
}

func decision(delta int, criticality scenario.Criticality, minutes int,
	tags []scenario.RiskTag, effects []scenario.SkillEffect,
) progress.DecisionFact {
	return progress.DecisionFact{
		Criticality:  criticality,
		RiskTags:     tags,
		SkillEffects: effects,
		ScoreDelta:   delta,
		CreatedAt:    base.Add(time.Duration(minutes) * time.Minute),
	}
}

// Новый пользователь обязан получить корректный пустой ответ, а не отказ.
func TestEmptyProfile(t *testing.T) {
	found, err := newService(progress.Facts{}).Of(context.Background(), owner)
	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}

	if found.Summary != (progress.Summary{}) {
		t.Errorf("сводка = %+v, ожидалась нулевая", found.Summary)
	}

	if len(found.ScenarioProgress) != 0 || len(found.Skills) != 0 ||
		len(found.WeakRiskTags) != 0 || len(found.RecentAttempts) != 0 ||
		len(found.ActiveAttempts) != 0 {
		t.Error("списки пустого профиля должны быть пустыми")
	}

	// Новичку есть что предложить: все сценарии ещё не пройдены.
	if len(found.Recommendations) == 0 {
		t.Error("пустому профилю нужна рекомендация с чего начать")
	}
}

func TestSummary(t *testing.T) {
	service := newService(progress.Facts{Completed: []progress.CompletedAttempt{
		completed("a-1", "buyer-one", 70, 10),
		completed("a-2", "buyer-one", 85, 20),
		completed("a-3", "seller-one", 90, 30),
	}})

	found, err := service.Of(context.Background(), owner)
	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}

	want := progress.Summary{
		CompletedAttempts:  3,
		CompletedScenarios: 2,
		AverageScore:       82, // (70+85+90)/3 = 81.67 -> 82
		BestScore:          90,
		LatestScore:        90,
	}

	if found.Summary != want {
		t.Errorf("сводка = %+v, ожидалась %+v", found.Summary, want)
	}
}

// Округление среднего зафиксировано правилом «половина вверх».
func TestAverageRounding(t *testing.T) {
	testCases := []struct {
		name   string
		scores []int
		want   int
	}{
		{name: "точное деление", scores: []int{80, 90}, want: 85},
		{name: "половина вверх", scores: []int{80, 81}, want: 81},
		{name: "округление вниз", scores: []int{80, 80, 81}, want: 80},
		{name: "одна попытка", scores: []int{73}, want: 73},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			attempts := make([]progress.CompletedAttempt, 0, len(testCase.scores))
			for i, score := range testCase.scores {
				attempts = append(attempts,
					completed(attempt.ID("a-"+string(rune('a'+i))), "buyer-one", score, i))
			}

			found, err := newService(progress.Facts{Completed: attempts}).Of(context.Background(), owner)
			if err != nil {
				t.Fatalf("неожиданная ошибка: %v", err)
			}

			if found.Summary.AverageScore != testCase.want {
				t.Errorf("среднее = %d, ожидалось %d", found.Summary.AverageScore, testCase.want)
			}
		})
	}
}

// Улучшение считается относительно предыдущего завершённого прохождения
// того же сценария.
func TestScenarioImprovement(t *testing.T) {
	testCases := []struct {
		name            string
		scores          []int
		wantLast        int
		wantBest        int
		wantPrevious    *int
		wantImprovement *int
	}{
		{name: "первое прохождение", scores: []int{70}, wantLast: 70, wantBest: 70},
		{
			name: "улучшение", scores: []int{70, 85}, wantLast: 85, wantBest: 85,
			wantPrevious: intPtr(70), wantImprovement: intPtr(15),
		},
		{
			name: "ухудшение", scores: []int{90, 60}, wantLast: 60, wantBest: 90,
			wantPrevious: intPtr(90), wantImprovement: intPtr(-30),
		},
		{
			name: "без изменений", scores: []int{75, 75}, wantLast: 75, wantBest: 75,
			wantPrevious: intPtr(75), wantImprovement: intPtr(0),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			attempts := make([]progress.CompletedAttempt, 0, len(testCase.scores))
			for i, score := range testCase.scores {
				attempts = append(attempts,
					completed(attempt.ID("a-"+string(rune('a'+i))), "buyer-one", score, i))
			}

			found, err := newService(progress.Facts{Completed: attempts}).Of(context.Background(), owner)
			if err != nil {
				t.Fatalf("неожиданная ошибка: %v", err)
			}

			if len(found.ScenarioProgress) != 1 {
				t.Fatalf("сценариев = %d, ожидался 1", len(found.ScenarioProgress))
			}

			item := found.ScenarioProgress[0]

			if item.Attempts != len(testCase.scores) {
				t.Errorf("попыток = %d, ожидалось %d", item.Attempts, len(testCase.scores))
			}

			if item.LastScore != testCase.wantLast || item.BestScore != testCase.wantBest {
				t.Errorf("последний/лучший = %d/%d, ожидалось %d/%d",
					item.LastScore, item.BestScore, testCase.wantLast, testCase.wantBest)
			}

			assertIntPtr(t, "предыдущий результат", item.PreviousScore, testCase.wantPrevious)
			assertIntPtr(t, "улучшение", item.Improvement, testCase.wantImprovement)
		})
	}
}

// Незавершённое прохождение показывается отдельно и не влияет на статистику.
func TestActiveAttemptsAreSeparate(t *testing.T) {
	service := newService(progress.Facts{
		Completed: []progress.CompletedAttempt{completed("a-1", "buyer-one", 70, 10)},
		Active: []progress.ActiveAttempt{
			{AttemptID: "a-2", ScenarioID: "buyer-two", UpdatedAt: base.Add(time.Hour)},
		},
	})

	found, err := service.Of(context.Background(), owner)
	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}

	if found.Summary.CompletedAttempts != 1 {
		t.Errorf("завершённых попыток = %d, активная попала в статистику", found.Summary.CompletedAttempts)
	}

	if len(found.ActiveAttempts) != 1 || found.ActiveAttempts[0].AttemptID != "a-2" {
		t.Errorf("активные попытки = %+v", found.ActiveAttempts)
	}

	// Сценарий с активным прохождением виден, чтобы к нему можно было вернуться.
	var active *progress.ScenarioProgress

	for i, item := range found.ScenarioProgress {
		if item.ScenarioID == "buyer-two" {
			active = &found.ScenarioProgress[i]
		}
	}

	if active == nil {
		t.Fatal("сценарий с активным прохождением отсутствует в прогрессе")
	}

	if active.Attempts != 0 || active.ActiveAttemptID != "a-2" {
		t.Errorf("сценарий с активным прохождением = %+v", *active)
	}
}

func TestSkillsAggregation(t *testing.T) {
	service := newService(progress.Facts{Completed: []progress.CompletedAttempt{
		completed("a-1", "buyer-one", 60, 10,
			decision(-20, scenario.CriticalityHigh, 1, nil,
				[]scenario.SkillEffect{{Skill: "payment_safety", Delta: -1}}),
			decision(0, scenario.CriticalityLow, 2, nil,
				[]scenario.SkillEffect{{Skill: "link_hygiene", Delta: 1}}),
		),
		completed("a-2", "buyer-two", 80, 20,
			decision(0, scenario.CriticalityLow, 3, nil,
				[]scenario.SkillEffect{{Skill: "payment_safety", Delta: 2}}),
		),
	}})

	found, err := service.Of(context.Background(), owner)
	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}

	if len(found.Skills) != 2 {
		t.Fatalf("навыков = %d, ожидалось 2", len(found.Skills))
	}

	// Слабые навыки идут первыми.
	if found.Skills[0].Code != "link_hygiene" {
		t.Errorf("первый навык = %q, ожидался link_hygiene", found.Skills[0].Code)
	}

	payment := found.Skills[1]
	if payment.Code != "payment_safety" {
		t.Fatalf("второй навык = %q, ожидался payment_safety", payment.Code)
	}

	if payment.Value != 1 || payment.Positive != 2 || payment.Negative != 1 || payment.Decisions != 2 {
		t.Errorf("навык = %+v, ожидалось value 1, positive 2, negative 1, decisions 2", payment)
	}
}

// Слабым считается то, на чём пользователь ошибался: признак берётся из
// серверной дельты результата, а не из текста сценария.
func TestWeakRiskTags(t *testing.T) {
	service := newService(progress.Facts{Completed: []progress.CompletedAttempt{
		completed("a-1", "buyer-one", 40, 10,
			decision(-20, scenario.CriticalityHigh, 1,
				[]scenario.RiskTag{"phishing_link"}, nil),
			decision(-10, scenario.CriticalityMedium, 2,
				[]scenario.RiskTag{"urgency"}, nil),
			// Безопасный выбор с той же меткой не считается ошибкой.
			decision(0, scenario.CriticalityLow, 3,
				[]scenario.RiskTag{"urgency"}, nil),
		),
		completed("a-2", "buyer-two", 80, 20,
			decision(-5, scenario.CriticalityLow, 4,
				[]scenario.RiskTag{"phishing_link"}, nil),
		),
	}})

	found, err := service.Of(context.Background(), owner)
	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}

	if len(found.WeakRiskTags) != 2 {
		t.Fatalf("меток = %d, ожидалось 2", len(found.WeakRiskTags))
	}

	// Самое опасное — первым: high (3) + low (1) = 4 против medium (2).
	weakest := found.WeakRiskTags[0]
	if weakest.Code != "phishing_link" || weakest.Mistakes != 2 || weakest.WeightedSeverity != 4 {
		t.Errorf("слабая метка = %+v, ожидалось phishing_link / 2 / 4", weakest)
	}

	if !weakest.LastOccurredAt.Equal(base.Add(4 * time.Minute)) {
		t.Errorf("последняя ошибка = %v, ожидалась %v", weakest.LastOccurredAt, base.Add(4*time.Minute))
	}

	if found.WeakRiskTags[1].Code != "urgency" || found.WeakRiskTags[1].Mistakes != 1 {
		t.Errorf("вторая метка = %+v, ожидалось urgency / 1", found.WeakRiskTags[1])
	}
}

func TestHistoryIsNewestFirst(t *testing.T) {
	service := newService(progress.Facts{Completed: []progress.CompletedAttempt{
		completed("a-1", "buyer-one", 70, 10, decision(-30, scenario.CriticalityHigh, 1, nil, nil)),
		completed("a-2", "seller-one", 90, 20),
	}})

	found, err := service.Of(context.Background(), owner)
	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}

	if len(found.RecentAttempts) != 2 {
		t.Fatalf("записей истории = %d, ожидалось 2", len(found.RecentAttempts))
	}

	if found.RecentAttempts[0].AttemptID != "a-2" {
		t.Errorf("первая запись = %q, ожидалась a-2", found.RecentAttempts[0].AttemptID)
	}

	// Метаданные сценария подставляются из каталога.
	if found.RecentAttempts[0].Title != "Продавец 1" || found.RecentAttempts[0].Role != scenario.RoleSeller {
		t.Errorf("метаданные записи = %+v", found.RecentAttempts[0])
	}

	if found.RecentAttempts[1].Decisions != 1 {
		t.Errorf("решений в записи = %d, ожидалось 1", found.RecentAttempts[1].Decisions)
	}
}

// Порядок ответа не должен зависеть от порядка строк в хранилище.
func TestOrderingIsDeterministic(t *testing.T) {
	forward := progress.Facts{Completed: []progress.CompletedAttempt{
		completed("a-1", "buyer-one", 70, 10),
		completed("a-2", "buyer-two", 80, 20),
		completed("a-3", "seller-one", 90, 30),
	}}

	reversed := progress.Facts{Completed: []progress.CompletedAttempt{
		forward.Completed[2], forward.Completed[1], forward.Completed[0],
	}}

	first, err := newService(forward).Of(context.Background(), owner)
	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}

	second, err := newService(reversed).Of(context.Background(), owner)
	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}

	if first.Summary != second.Summary {
		t.Errorf("сводки различаются: %+v и %+v", first.Summary, second.Summary)
	}

	for i := range first.ScenarioProgress {
		if first.ScenarioProgress[i].ScenarioID != second.ScenarioProgress[i].ScenarioID {
			t.Fatalf("порядок сценариев различается: %q и %q",
				first.ScenarioProgress[i].ScenarioID, second.ScenarioProgress[i].ScenarioID)
		}
	}

	for i := range first.RecentAttempts {
		if first.RecentAttempts[i].AttemptID != second.RecentAttempts[i].AttemptID {
			t.Fatalf("порядок истории различается: %q и %q",
				first.RecentAttempts[i].AttemptID, second.RecentAttempts[i].AttemptID)
		}
	}
}

// Рекомендация опирается только на публичные метаданные каталога.
func TestRecommendations(t *testing.T) {
	service := newService(progress.Facts{Completed: []progress.CompletedAttempt{
		completed("a-1", "buyer-one", 60, 10),
		completed("a-2", "seller-one", 100, 20),
	}})

	found, err := service.Of(context.Background(), owner)
	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}

	if len(found.Recommendations) != 2 {
		t.Fatalf("рекомендаций = %d, ожидалось 2: %+v", len(found.Recommendations), found.Recommendations)
	}

	// Сначала непройденный сценарий.
	if found.Recommendations[0].ScenarioID != "buyer-two" ||
		found.Recommendations[0].Reason != progress.ReasonNotCompleted {
		t.Errorf("первая рекомендация = %+v, ожидался непройденный buyer-two", found.Recommendations[0])
	}

	// Затем пройденный с неидеальным результатом.
	if found.Recommendations[1].ScenarioID != "buyer-one" ||
		found.Recommendations[1].Reason != progress.ReasonImproveScore {
		t.Errorf("вторая рекомендация = %+v, ожидался buyer-one на улучшение", found.Recommendations[1])
	}

	// Идеально пройденный сценарий не предлагается.
	for _, item := range found.Recommendations {
		if item.ScenarioID == "seller-one" {
			t.Error("сценарий с идеальным результатом не должен рекомендоваться")
		}
	}
}

// Повтор запроса не удваивает статистику: она считается по сохранённым
// фактам, а не накапливается счётчиками.
func TestRepeatedReadIsStable(t *testing.T) {
	service := newService(progress.Facts{Completed: []progress.CompletedAttempt{
		completed("a-1", "buyer-one", 70, 10,
			decision(-30, scenario.CriticalityHigh, 1,
				[]scenario.RiskTag{"phishing_link"},
				[]scenario.SkillEffect{{Skill: "link_hygiene", Delta: -1}})),
	}})

	first, err := service.Of(context.Background(), owner)
	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}

	second, err := service.Of(context.Background(), owner)
	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}

	if first.Summary != second.Summary {
		t.Errorf("сводка изменилась при повторном чтении: %+v и %+v", first.Summary, second.Summary)
	}

	if first.Skills[0] != second.Skills[0] || first.WeakRiskTags[0] != second.WeakRiskTags[0] {
		t.Error("аналитика изменилась при повторном чтении")
	}
}

func TestProgressRequiresProfile(t *testing.T) {
	_, err := newService(progress.Facts{}).Of(context.Background(), "")
	if !errors.Is(err, profile.ErrEmptyID) {
		t.Errorf("ошибка = %v, ожидалась ErrEmptyID", err)
	}
}

func TestProgressPropagatesFailures(t *testing.T) {
	failure := errors.New("хранилище недоступно")

	service := progress.NewService(stubRepository{failure: failure}, catalog())
	if _, err := service.Of(context.Background(), owner); !errors.Is(err, failure) {
		t.Errorf("ошибка = %v, ожидалась причина хранилища", err)
	}

	catalogFailure := errors.New("каталог недоступен")

	service = progress.NewService(stubRepository{}, stubCatalog{failure: catalogFailure})
	if _, err := service.Of(context.Background(), owner); !errors.Is(err, catalogFailure) {
		t.Errorf("ошибка = %v, ожидалась причина каталога", err)
	}
}

func intPtr(value int) *int { return &value }

func assertIntPtr(t *testing.T, name string, got, want *int) {
	t.Helper()

	switch {
	case want == nil && got != nil:
		t.Errorf("%s = %d, ожидалось отсутствие значения", name, *got)
	case want != nil && got == nil:
		t.Errorf("%s отсутствует, ожидалось %d", name, *want)
	case want != nil && got != nil && *want != *got:
		t.Errorf("%s = %d, ожидалось %d", name, *got, *want)
	}
}
