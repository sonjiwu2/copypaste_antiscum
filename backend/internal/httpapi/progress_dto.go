package httpapi

import (
	"time"

	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/progress"
)

// Структуры ниже — публичный контракт прогресса.
//
// Наружу уходит только собственная статистика пользователя: содержимое
// непройденных сценариев, их метки риска и веса ответов сюда не попадают.

type progressResponse struct {
	Summary          progressSummary          `json:"summary"`
	ActiveAttempts   []progressActiveAttempt  `json:"activeAttempts"`
	ScenarioProgress []progressScenario       `json:"scenarioProgress"`
	Skills           []progressSkill          `json:"skills"`
	WeakRiskTags     []progressWeakRiskTag    `json:"weakRiskTags"`
	RecentAttempts   []progressHistoryEntry   `json:"recentAttempts"`
	Recommendations  []progressRecommendation `json:"recommendedScenarios"`
}

type progressSummary struct {
	CompletedAttempts  int `json:"completedAttempts"`
	CompletedScenarios int `json:"completedScenarios"`
	AverageScore       int `json:"averageScore"`
	BestScore          int `json:"bestScore"`
	LatestScore        int `json:"latestScore"`
}

type progressActiveAttempt struct {
	AttemptID  string `json:"attemptId"`
	ScenarioID string `json:"scenarioId"`
	UpdatedAt  string `json:"updatedAt"`
}

type progressScenario struct {
	ScenarioID      string  `json:"scenarioId"`
	Role            string  `json:"role,omitempty"`
	Title           string  `json:"title,omitempty"`
	Attempts        int     `json:"attempts"`
	LastScore       int     `json:"lastScore"`
	BestScore       int     `json:"bestScore"`
	PreviousScore   *int    `json:"previousScore,omitempty"`
	Improvement     *int    `json:"improvement,omitempty"`
	LastCompletedAt *string `json:"lastCompletedAt,omitempty"`
	ActiveAttemptID string  `json:"activeAttemptId,omitempty"`
}

type progressSkill struct {
	Code      string `json:"code"`
	Value     int    `json:"value"`
	Positive  int    `json:"positive"`
	Negative  int    `json:"negative"`
	Decisions int    `json:"decisions"`
}

type progressWeakRiskTag struct {
	Code             string `json:"code"`
	Mistakes         int    `json:"mistakes"`
	WeightedSeverity int    `json:"weightedSeverity"`
	LastOccurredAt   string `json:"lastOccurredAt"`
}

type progressHistoryEntry struct {
	AttemptID   string `json:"attemptId"`
	ScenarioID  string `json:"scenarioId"`
	Role        string `json:"role,omitempty"`
	Title       string `json:"title,omitempty"`
	Score       int    `json:"score"`
	Outcome     string `json:"outcome,omitempty"`
	StartedAt   string `json:"startedAt"`
	CompletedAt string `json:"completedAt"`
	Decisions   int    `json:"decisions"`
}

type progressRecommendation struct {
	ScenarioID       string `json:"scenarioId"`
	Role             string `json:"role"`
	Title            string `json:"title"`
	Difficulty       string `json:"difficulty"`
	EstimatedMinutes int    `json:"estimatedMinutes"`
	Reason           string `json:"reason"`
}

func progressResponseOf(source progress.Progress) progressResponse {
	return progressResponse{
		Summary: progressSummary{
			CompletedAttempts:  source.Summary.CompletedAttempts,
			CompletedScenarios: source.Summary.CompletedScenarios,
			AverageScore:       source.Summary.AverageScore,
			BestScore:          source.Summary.BestScore,
			LatestScore:        source.Summary.LatestScore,
		},
		ActiveAttempts:   progressActiveAttemptsOf(source.ActiveAttempts),
		ScenarioProgress: progressScenariosOf(source.ScenarioProgress),
		Skills:           progressSkillsOf(source.Skills),
		WeakRiskTags:     progressWeakRiskTagsOf(source.WeakRiskTags),
		RecentAttempts:   progressHistoryOf(source.RecentAttempts),
		Recommendations:  progressRecommendationsOf(source.Recommendations),
	}
}

func progressActiveAttemptsOf(source []progress.ActiveAttempt) []progressActiveAttempt {
	converted := make([]progressActiveAttempt, 0, len(source))

	for _, item := range source {
		converted = append(converted, progressActiveAttempt{
			AttemptID:  string(item.AttemptID),
			ScenarioID: string(item.ScenarioID),
			UpdatedAt:  item.UpdatedAt.Format(time.RFC3339),
		})
	}

	return converted
}

func progressScenariosOf(source []progress.ScenarioProgress) []progressScenario {
	converted := make([]progressScenario, 0, len(source))

	for _, item := range source {
		entry := progressScenario{
			ScenarioID:      string(item.ScenarioID),
			Role:            string(item.Role),
			Title:           item.Title,
			Attempts:        item.Attempts,
			LastScore:       item.LastScore,
			BestScore:       item.BestScore,
			PreviousScore:   item.PreviousScore,
			Improvement:     item.Improvement,
			ActiveAttemptID: string(item.ActiveAttemptID),
		}

		// Нулевое время означает, что сценарий ещё не завершался.
		if !item.LastCompletedAt.IsZero() {
			completedAt := item.LastCompletedAt.Format(time.RFC3339)
			entry.LastCompletedAt = &completedAt
		}

		converted = append(converted, entry)
	}

	return converted
}

func progressSkillsOf(source []progress.Skill) []progressSkill {
	converted := make([]progressSkill, 0, len(source))

	for _, item := range source {
		converted = append(converted, progressSkill{
			Code:      string(item.Code),
			Value:     item.Value,
			Positive:  item.Positive,
			Negative:  item.Negative,
			Decisions: item.Decisions,
		})
	}

	return converted
}

func progressWeakRiskTagsOf(source []progress.WeakRiskTag) []progressWeakRiskTag {
	converted := make([]progressWeakRiskTag, 0, len(source))

	for _, item := range source {
		converted = append(converted, progressWeakRiskTag{
			Code:             string(item.Code),
			Mistakes:         item.Mistakes,
			WeightedSeverity: item.WeightedSeverity,
			LastOccurredAt:   item.LastOccurredAt.Format(time.RFC3339),
		})
	}

	return converted
}

func progressHistoryOf(source []progress.HistoryEntry) []progressHistoryEntry {
	converted := make([]progressHistoryEntry, 0, len(source))

	for _, item := range source {
		converted = append(converted, progressHistoryEntry{
			AttemptID:   string(item.AttemptID),
			ScenarioID:  string(item.ScenarioID),
			Role:        string(item.Role),
			Title:       item.Title,
			Score:       item.Score,
			Outcome:     string(item.Outcome),
			StartedAt:   item.StartedAt.Format(time.RFC3339),
			CompletedAt: item.CompletedAt.Format(time.RFC3339),
			Decisions:   item.Decisions,
		})
	}

	return converted
}

func progressRecommendationsOf(source []progress.Recommendation) []progressRecommendation {
	converted := make([]progressRecommendation, 0, len(source))

	for _, item := range source {
		converted = append(converted, progressRecommendation{
			ScenarioID:       string(item.ScenarioID),
			Role:             string(item.Role),
			Title:            item.Title,
			Difficulty:       string(item.Difficulty),
			EstimatedMinutes: item.EstimatedMinutes,
			Reason:           string(item.Reason),
		})
	}

	return converted
}
