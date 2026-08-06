package progress

import (
	"sort"

	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/scenario"
)

// perfectScore — результат, выше которого улучшать нечего.
const perfectScore = 100

// recommendationsOf предлагает, что пройти дальше.
//
// Рекомендация опирается только на публичные метаданные каталога и на
// собственную статистику пользователя. Содержимое непройденного сценария,
// его метки риска и веса ответов сюда не попадают — иначе подсказка
// раскрывала бы правильные действия заранее.
//
// Порядок приоритетов:
//  1. сценарий ещё ни разу не пройден до конца;
//  2. сценарий пройден, но результат неидеален — от худшего к лучшему.
//
// Отключённые сценарии в каталог не попадают, поэтому предложены не будут.
func recommendationsOf(catalog []scenario.Metadata, progress []ScenarioProgress) []Recommendation {
	best := make(map[scenario.ID]int, len(progress))
	completed := make(map[scenario.ID]bool, len(progress))

	for _, item := range progress {
		if item.Attempts == 0 {
			continue
		}

		completed[item.ScenarioID] = true
		best[item.ScenarioID] = item.BestScore
	}

	candidates := make([]Recommendation, 0, len(catalog))

	for _, item := range catalog {
		if completed[item.ID] && best[item.ID] >= perfectScore {
			continue
		}

		reason := ReasonNotCompleted
		if completed[item.ID] {
			reason = ReasonImproveScore
		}

		candidates = append(candidates, Recommendation{
			ScenarioID:       item.ID,
			Role:             item.Role,
			Title:            item.Title,
			Difficulty:       item.Difficulty,
			EstimatedMinutes: item.EstimatedMinutes,
			Reason:           reason,
		})
	}

	// Сначала непройденные, затем — с худшим результатом.
	// Идентификатор разрывает совпадения, чтобы порядок был устойчив.
	sort.Slice(candidates, func(i, j int) bool {
		left, right := candidates[i], candidates[j]

		if (left.Reason == ReasonNotCompleted) != (right.Reason == ReasonNotCompleted) {
			return left.Reason == ReasonNotCompleted
		}

		if left.Reason == ReasonImproveScore && best[left.ScenarioID] != best[right.ScenarioID] {
			return best[left.ScenarioID] < best[right.ScenarioID]
		}

		return left.ScenarioID < right.ScenarioID
	})

	if len(candidates) > recommendationLimit {
		candidates = candidates[:recommendationLimit]
	}

	return candidates
}
