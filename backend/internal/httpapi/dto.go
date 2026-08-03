package httpapi

import "github.com/sonjiwu2/copypaste_antiscum/backend/internal/scenario"

// Структуры ниже — публичный контракт API.
// Доменные значения не сериализуются напрямую: клиент получает только то,
// что перечислено здесь явно.

// scenarioListResponse — ответ каталога сценариев.
type scenarioListResponse struct {
	Scenarios []scenarioSummary `json:"scenarios"`
}

// scenarioSummary — краткое описание сценария без содержимого графа.
type scenarioSummary struct {
	ID               string `json:"id"`
	Version          int    `json:"version"`
	Slug             string `json:"slug"`
	Role             string `json:"role"`
	Title            string `json:"title"`
	Description      string `json:"description"`
	Difficulty       string `json:"difficulty"`
	EstimatedMinutes int    `json:"estimatedMinutes"`
}

func scenarioSummaryOf(metadata scenario.Metadata) scenarioSummary {
	return scenarioSummary{
		ID:               string(metadata.ID),
		Version:          int(metadata.Version),
		Slug:             metadata.Slug,
		Role:             string(metadata.Role),
		Title:            metadata.Title,
		Description:      metadata.Description,
		Difficulty:       string(metadata.Difficulty),
		EstimatedMinutes: metadata.EstimatedMinutes,
	}
}

func scenarioSummariesOf(catalog []scenario.Metadata) []scenarioSummary {
	summaries := make([]scenarioSummary, 0, len(catalog))
	for _, metadata := range catalog {
		summaries = append(summaries, scenarioSummaryOf(metadata))
	}

	return summaries
}
