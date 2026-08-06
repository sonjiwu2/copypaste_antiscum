package memory

import (
	"context"

	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/attempt"
	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/profile"
	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/progress"
)

// ProgressRepository собирает факты прогресса из хранилища попыток в памяти.
//
// Отдельного состояния у него нет: прогресс всегда считается по тем же
// попыткам, что видит движок прохождения, поэтому расхождение невозможно.
type ProgressRepository struct {
	attempts *AttemptRepository
}

// NewProgressRepository создаёт хранилище прогресса поверх попыток в памяти.
func NewProgressRepository(attempts *AttemptRepository) *ProgressRepository {
	return &ProgressRepository{attempts: attempts}
}

// Facts возвращает завершённые и активные прохождения профиля.
func (r *ProgressRepository) Facts(ctx context.Context, owner profile.ID) (progress.Facts, error) {
	if err := ctx.Err(); err != nil {
		return progress.Facts{}, err
	}

	if owner == "" {
		return progress.Facts{}, profile.ErrEmptyID
	}

	facts := progress.Facts{
		Completed: make([]progress.CompletedAttempt, 0),
		Active:    make([]progress.ActiveAttempt, 0),
	}

	for _, stored := range r.attempts.ownedBy(owner) {
		if stored.Status == attempt.StatusInProgress {
			facts.Active = append(facts.Active, progress.ActiveAttempt{
				AttemptID:  stored.ID,
				ScenarioID: stored.ScenarioID,
				UpdatedAt:  stored.UpdatedAt,
			})

			continue
		}

		facts.Completed = append(facts.Completed, completedAttemptOf(stored))
	}

	return facts, nil
}

func completedAttemptOf(stored attempt.Attempt) progress.CompletedAttempt {
	completed := progress.CompletedAttempt{
		AttemptID:  stored.ID,
		ScenarioID: stored.ScenarioID,
		Score:      stored.Score,
		Outcome:    stored.Outcome,
		StartedAt:  stored.StartedAt,
		Decisions:  make([]progress.DecisionFact, 0, len(stored.Decisions)),
	}

	if stored.CompletedAt != nil {
		completed.CompletedAt = *stored.CompletedAt
	}

	for _, decision := range stored.Decisions {
		completed.Decisions = append(completed.Decisions, progress.DecisionFact{
			Criticality:  decision.Criticality,
			RiskTags:     decision.RiskTags,
			SkillEffects: decision.SkillEffects,
			ScoreDelta:   decision.ScoreDelta,
			CreatedAt:    decision.CreatedAt,
		})
	}

	return completed
}
