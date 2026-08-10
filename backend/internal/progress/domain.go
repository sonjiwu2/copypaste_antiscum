// Package progress считает прогресс пользователя по сохранённым попыткам.
//
// Второго алгоритма подсчёта результата здесь нет: score каждой попытки уже
// вычислен движком прохождения и хранится в базе. Пакет только агрегирует
// сохранённые факты, поэтому расхождение между экраном результата и экраном
// прогресса невозможно.
package progress

import (
	"time"

	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/attempt"
	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/scenario"
)

// Веса критичности для оценки слабых мест.
//
// Значения внутренние: наружу уходит только сумма по уже сделанным ошибкам,
// поэтому по ней нельзя восстановить веса непройденных веток сценария.
const (
	weightLow    = 1
	weightMedium = 2
	weightHigh   = 3
)

// recentAttemptsLimit ограничивает историю в ответе: экран показывает
// последние прохождения, а не весь архив.
const recentAttemptsLimit = 10

// recommendationLimit — сколько сценариев предлагать дальше.
const recommendationLimit = 3

// DecisionFact — сохранённое решение, нужное для аналитики.
//
// Тексты последствий сюда не входят: прогресс считает статистику, а не
// пересказывает прохождение.
type DecisionFact struct {
	Criticality  scenario.Criticality
	RiskTags     []scenario.RiskTag
	SkillEffects []scenario.SkillEffect
	ScoreDelta   int
	CreatedAt    time.Time
}

// Mistake сообщает, был ли выбор ошибочным.
//
// Признак берётся из серверной дельты результата: клиент на него не влияет,
// и определение не зависит от текста сценария.
func (d DecisionFact) Mistake() bool {
	return d.ScoreDelta < 0
}

// Weight переводит критичность решения в вес ошибки.
func (d DecisionFact) Weight() int {
	switch d.Criticality {
	case scenario.CriticalityHigh:
		return weightHigh
	case scenario.CriticalityMedium:
		return weightMedium
	case scenario.CriticalityLow:
		return weightLow
	default:
		return weightLow
	}
}

// CompletedAttempt — завершённое прохождение.
type CompletedAttempt struct {
	AttemptID   attempt.ID
	ScenarioID  scenario.ID
	Score       int
	Outcome     scenario.OutcomeType
	StartedAt   time.Time
	CompletedAt time.Time
	Decisions   []DecisionFact
}

// ActiveAttempt — незавершённое прохождение.
// В статистику оно не входит и показывается отдельно.
type ActiveAttempt struct {
	AttemptID  attempt.ID
	ScenarioID scenario.ID
	UpdatedAt  time.Time
}

// Facts — сохранённые данные профиля, из которых считается прогресс.
type Facts struct {
	Completed []CompletedAttempt
	Active    []ActiveAttempt
}

// Summary — сводка по завершённым прохождениям.
type Summary struct {
	CompletedAttempts  int
	CompletedScenarios int
	AverageScore       int
	BestScore          int
	LatestScore        int
}

// ScenarioProgress — прогресс по одному сценарию.
type ScenarioProgress struct {
	ScenarioID      scenario.ID
	Role            scenario.Role
	Title           string
	Attempts        int
	LastScore       int
	BestScore       int
	LastCompletedAt time.Time

	// PreviousScore и Improvement заполняются только со второго прохождения:
	// сравнивать первый результат не с чем.
	PreviousScore *int
	Improvement   *int

	// ActiveAttemptID указывает на незавершённое прохождение этого сценария.
	ActiveAttemptID attempt.ID
}

// Skill — накопленное состояние навыка безопасности.
type Skill struct {
	Code      scenario.SkillCode
	Value     int
	Positive  int
	Negative  int
	Decisions int
}

// WeakRiskTag — мошенническая схема, на которой пользователь ошибается.
type WeakRiskTag struct {
	Code             scenario.RiskTag
	Mistakes         int
	WeightedSeverity int
	LastOccurredAt   time.Time
}

// HistoryEntry — запись истории прохождений.
type HistoryEntry struct {
	AttemptID   attempt.ID
	ScenarioID  scenario.ID
	Role        scenario.Role
	Title       string
	Score       int
	Outcome     scenario.OutcomeType
	StartedAt   time.Time
	CompletedAt time.Time
	Decisions   int
}

// Recommendation — предложение следующего сценария.
//
// Состав полей ограничен публичными метаданными каталога: рекомендация не
// имеет права намекать на содержимое непройденного сценария.
type Recommendation struct {
	ScenarioID       scenario.ID
	Role             scenario.Role
	Title            string
	Difficulty       scenario.Difficulty
	EstimatedMinutes int
	Reason           Reason
}

// Reason — почему сценарий предложен.
type Reason string

// Причины рекомендации.
const (
	// ReasonNotCompleted — сценарий ещё ни разу не пройден до конца.
	ReasonNotCompleted Reason = "not_completed"

	// ReasonImproveScore — сценарий пройден, но результат можно улучшить.
	ReasonImproveScore Reason = "improve_score"
)

// Progress — полный прогресс профиля.
type Progress struct {
	Summary          Summary
	ActiveAttempts   []ActiveAttempt
	ScenarioProgress []ScenarioProgress
	Skills           []Skill
	WeakRiskTags     []WeakRiskTag
	RecentAttempts   []HistoryEntry
	Recommendations  []Recommendation
}
