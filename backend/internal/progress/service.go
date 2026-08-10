package progress

import (
	"context"
	"fmt"
	"sort"

	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/attempt"
	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/profile"
	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/scenario"
)

// Repository отдаёт сохранённые факты одного профиля.
//
// Интерфейс объявлен рядом с потребителем: подсчёт живёт в домене, а SQL —
// в адаптере хранения.
type Repository interface {
	Facts(ctx context.Context, owner profile.ID) (Facts, error)
}

// Catalog — публичные метаданные сценариев, нужные прогрессу.
type Catalog interface {
	List(ctx context.Context, role scenario.Role) ([]scenario.Metadata, error)
}

// Service собирает прогресс профиля.
type Service struct {
	repository Repository
	catalog    Catalog
}

// NewService создаёт сервис прогресса.
func NewService(repository Repository, catalog Catalog) *Service {
	return &Service{repository: repository, catalog: catalog}
}

// Of возвращает прогресс профиля.
//
// Пустой профиль — не ошибка: новый пользователь получает нулевую сводку и
// пустые списки, а не отказ.
func (s *Service) Of(ctx context.Context, owner profile.ID) (Progress, error) {
	if owner == "" {
		return Progress{}, profile.ErrEmptyID
	}

	facts, err := s.repository.Facts(ctx, owner)
	if err != nil {
		return Progress{}, fmt.Errorf("получить данные прогресса: %w", err)
	}

	catalog, err := s.catalog.List(ctx, "")
	if err != nil {
		return Progress{}, fmt.Errorf("получить каталог сценариев: %w", err)
	}

	metadata := make(map[scenario.ID]scenario.Metadata, len(catalog))
	for _, item := range catalog {
		metadata[item.ID] = item
	}

	// Порядок хранилища не должен влиять на ответ: сначала приводим факты
	// к детерминированному виду, потом считаем.
	completed := sortedCompleted(facts.Completed)
	active := sortedActive(facts.Active)

	scenarios := scenarioProgressOf(completed, active, metadata)

	return Progress{
		Summary:          summaryOf(completed),
		ActiveAttempts:   active,
		ScenarioProgress: scenarios,
		Skills:           skillsOf(completed),
		WeakRiskTags:     weakRiskTagsOf(completed),
		RecentAttempts:   historyOf(completed, metadata),
		Recommendations:  recommendationsOf(catalog, scenarios),
	}, nil
}

// sortedCompleted упорядочивает завершённые попытки по времени завершения.
// Идентификатор разрывает совпадения, иначе порядок был бы неустойчив.
func sortedCompleted(source []CompletedAttempt) []CompletedAttempt {
	completed := make([]CompletedAttempt, len(source))
	copy(completed, source)

	sort.SliceStable(completed, func(i, j int) bool {
		if !completed[i].CompletedAt.Equal(completed[j].CompletedAt) {
			return completed[i].CompletedAt.Before(completed[j].CompletedAt)
		}

		return completed[i].AttemptID < completed[j].AttemptID
	})

	return completed
}

// sortedActive показывает недавно тронутые прохождения первыми.
func sortedActive(source []ActiveAttempt) []ActiveAttempt {
	active := make([]ActiveAttempt, len(source))
	copy(active, source)

	sort.SliceStable(active, func(i, j int) bool {
		if !active[i].UpdatedAt.Equal(active[j].UpdatedAt) {
			return active[i].UpdatedAt.After(active[j].UpdatedAt)
		}

		return active[i].AttemptID > active[j].AttemptID
	})

	return active
}

// summaryOf считает сводку по завершённым попыткам.
// Незавершённые прохождения в статистику не входят.
func summaryOf(completed []CompletedAttempt) Summary {
	summary := Summary{}
	if len(completed) == 0 {
		return summary
	}

	scenarios := make(map[scenario.ID]struct{}, len(completed))
	total := 0

	for _, item := range completed {
		scenarios[item.ScenarioID] = struct{}{}
		total += item.Score

		if item.Score > summary.BestScore {
			summary.BestScore = item.Score
		}
	}

	summary.CompletedAttempts = len(completed)
	summary.CompletedScenarios = len(scenarios)
	summary.AverageScore = roundedAverage(total, len(completed))
	// Попытки уже упорядочены по времени завершения.
	summary.LatestScore = completed[len(completed)-1].Score

	return summary
}

// roundedAverage округляет по правилу «половина вверх».
// Правило зафиксировано здесь одно на весь продукт.
func roundedAverage(total, count int) int {
	if count == 0 {
		return 0
	}

	return (2*total + count) / (2 * count)
}

// scenarioProgressOf считает прогресс по каждому сценарию.
func scenarioProgressOf(
	completed []CompletedAttempt,
	active []ActiveAttempt,
	metadata map[scenario.ID]scenario.Metadata,
) []ScenarioProgress {
	grouped := make(map[scenario.ID][]CompletedAttempt)
	for _, item := range completed {
		grouped[item.ScenarioID] = append(grouped[item.ScenarioID], item)
	}

	// Активное прохождение показывается рядом со сценарием, но результата
	// ещё не имеет, поэтому в расчёт не входит.
	activeByScenario := make(map[scenario.ID]attempt.ID, len(active))
	for _, item := range active {
		if _, exists := activeByScenario[item.ScenarioID]; !exists {
			activeByScenario[item.ScenarioID] = item.AttemptID
		}
	}

	progress := make([]ScenarioProgress, 0, len(grouped)+len(activeByScenario))

	for scenarioID, attempts := range grouped {
		progress = append(progress, scenarioProgressItem(
			scenarioID, attempts, activeByScenario[scenarioID], metadata))
	}

	// Сценарий с одним лишь активным прохождением тоже виден: пользователь
	// должен найти, куда вернуться.
	for scenarioID, attemptID := range activeByScenario {
		if _, exists := grouped[scenarioID]; exists {
			continue
		}

		item := ScenarioProgress{ScenarioID: scenarioID, ActiveAttemptID: attemptID}
		if found, exists := metadata[scenarioID]; exists {
			item.Role = found.Role
			item.Title = found.Title
		}

		progress = append(progress, item)
	}

	sort.Slice(progress, func(i, j int) bool {
		return progress[i].ScenarioID < progress[j].ScenarioID
	})

	return progress
}

func scenarioProgressItem(
	scenarioID scenario.ID,
	attempts []CompletedAttempt,
	activeAttemptID attempt.ID,
	metadata map[scenario.ID]scenario.Metadata,
) ScenarioProgress {
	item := ScenarioProgress{
		ScenarioID:      scenarioID,
		Attempts:        len(attempts),
		ActiveAttemptID: activeAttemptID,
	}

	if found, exists := metadata[scenarioID]; exists {
		item.Role = found.Role
		item.Title = found.Title
	}

	for _, current := range attempts {
		if current.Score > item.BestScore {
			item.BestScore = current.Score
		}
	}

	last := attempts[len(attempts)-1]
	item.LastScore = last.Score
	item.LastCompletedAt = last.CompletedAt

	if len(attempts) >= 2 {
		previous := attempts[len(attempts)-2].Score
		improvement := last.Score - previous

		item.PreviousScore = &previous
		item.Improvement = &improvement
	}

	return item
}

// skillsOf суммирует эффекты навыков по завершённым прохождениям.
func skillsOf(completed []CompletedAttempt) []Skill {
	totals := make(map[scenario.SkillCode]*Skill)

	for _, item := range completed {
		for _, decision := range item.Decisions {
			for _, effect := range decision.SkillEffects {
				skill, exists := totals[effect.Skill]
				if !exists {
					skill = &Skill{Code: effect.Skill}
					totals[effect.Skill] = skill
				}

				skill.Value += effect.Delta
				skill.Decisions++

				if effect.Delta >= 0 {
					skill.Positive += effect.Delta
				} else {
					skill.Negative -= effect.Delta
				}
			}
		}
	}

	skills := make([]Skill, 0, len(totals))
	for _, skill := range totals {
		skills = append(skills, *skill)
	}

	// Слабые навыки идут первыми: экран прогресса показывает, что подтянуть.
	sort.Slice(skills, func(i, j int) bool {
		if skills[i].Value != skills[j].Value {
			return skills[i].Value < skills[j].Value
		}

		return skills[i].Code < skills[j].Code
	})

	return skills
}

// weakRiskTagsOf собирает схемы, на которых пользователь ошибался.
func weakRiskTagsOf(completed []CompletedAttempt) []WeakRiskTag {
	totals := make(map[scenario.RiskTag]*WeakRiskTag)

	for _, item := range completed {
		for _, decision := range item.Decisions {
			if !decision.Mistake() {
				continue
			}

			for _, tag := range decision.RiskTags {
				weak, exists := totals[tag]
				if !exists {
					weak = &WeakRiskTag{Code: tag}
					totals[tag] = weak
				}

				weak.Mistakes++
				weak.WeightedSeverity += decision.Weight()

				if decision.CreatedAt.After(weak.LastOccurredAt) {
					weak.LastOccurredAt = decision.CreatedAt
				}
			}
		}
	}

	weak := make([]WeakRiskTag, 0, len(totals))
	for _, tag := range totals {
		weak = append(weak, *tag)
	}

	// Самое опасное — первым.
	sort.Slice(weak, func(i, j int) bool {
		if weak[i].WeightedSeverity != weak[j].WeightedSeverity {
			return weak[i].WeightedSeverity > weak[j].WeightedSeverity
		}

		return weak[i].Code < weak[j].Code
	})

	return weak
}

// historyOf возвращает последние завершённые прохождения.
func historyOf(completed []CompletedAttempt, metadata map[scenario.ID]scenario.Metadata) []HistoryEntry {
	history := make([]HistoryEntry, 0, len(completed))

	// Обход с конца: последние прохождения показываются первыми.
	for i := len(completed) - 1; i >= 0 && len(history) < recentAttemptsLimit; i-- {
		item := completed[i]

		entry := HistoryEntry{
			AttemptID:   item.AttemptID,
			ScenarioID:  item.ScenarioID,
			Score:       item.Score,
			Outcome:     item.Outcome,
			StartedAt:   item.StartedAt,
			CompletedAt: item.CompletedAt,
			Decisions:   len(item.Decisions),
		}

		if found, exists := metadata[item.ScenarioID]; exists {
			entry.Role = found.Role
			entry.Title = found.Title
		}

		history = append(history, entry)
	}

	return history
}
