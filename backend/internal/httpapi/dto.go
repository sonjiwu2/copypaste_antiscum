package httpapi

import (
	"time"

	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/attempt"
	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/scenario"
)

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

// startAttemptRequest — тело запроса на создание попытки.
type startAttemptRequest struct {
	ScenarioID string `json:"scenarioId"`
}

// attemptResponse — публичное состояние попытки.
type attemptResponse struct {
	AttemptID     string             `json:"attemptId"`
	Scenario      attemptScenarioRef `json:"scenario"`
	Status        string             `json:"status"`
	Score         int                `json:"score"`
	Outcome       string             `json:"outcome,omitempty"`
	CurrentNodeID string             `json:"currentNodeId"`
	RevealedNodes []attemptNode      `json:"revealedNodes"`
	Decisions     []attemptDecision  `json:"decisions"`
	StartedAt     string             `json:"startedAt"`
	UpdatedAt     string             `json:"updatedAt"`
	CompletedAt   *string            `json:"completedAt,omitempty"`
}

type attemptScenarioRef struct {
	ID      string `json:"id"`
	Version int    `json:"version"`
	Slug    string `json:"slug"`
	Role    string `json:"role"`
	Title   string `json:"title"`
}

type attemptNode struct {
	ID      string          `json:"id"`
	Type    string          `json:"type"`
	Sender  string          `json:"sender,omitempty"`
	Text    string          `json:"text,omitempty"`
	Prompt  string          `json:"prompt,omitempty"`
	Choices []attemptChoice `json:"choices,omitempty"`
	Outcome *attemptOutcome `json:"outcome,omitempty"`
}

type attemptChoice struct {
	ID    string `json:"id"`
	Label string `json:"label"`
}

type attemptOutcome struct {
	Type        string `json:"type"`
	Title       string `json:"title"`
	Explanation string `json:"explanation"`
}

type attemptDecision struct {
	NodeID      string             `json:"nodeId"`
	ChoiceID    string             `json:"choiceId"`
	Label       string             `json:"label"`
	Consequence attemptConsequence `json:"consequence"`
}

type attemptConsequence struct {
	Severity      string `json:"severity"`
	Title         string `json:"title"`
	Explanation   string `json:"explanation"`
	RealWorldRule string `json:"realWorldRule,omitempty"`
}

func attemptResponseOf(view attempt.View) attemptResponse {
	response := attemptResponse{
		AttemptID: string(view.ID),
		Scenario: attemptScenarioRef{
			ID:      string(view.Scenario.ID),
			Version: int(view.Scenario.Version),
			Slug:    view.Scenario.Slug,
			Role:    string(view.Scenario.Role),
			Title:   view.Scenario.Title,
		},
		Status:        string(view.Status),
		Score:         view.Score,
		Outcome:       string(view.Outcome),
		CurrentNodeID: string(view.CurrentNodeID),
		RevealedNodes: attemptNodesOf(view.RevealedNodes),
		Decisions:     attemptDecisionsOf(view.Decisions),
		StartedAt:     view.StartedAt.Format(time.RFC3339),
		UpdatedAt:     view.UpdatedAt.Format(time.RFC3339),
	}

	if view.CompletedAt != nil {
		completedAt := view.CompletedAt.Format(time.RFC3339)
		response.CompletedAt = &completedAt
	}

	return response
}

// submitChoiceRequest — тело запроса на применение выбора.
//
// Следующего узла, изменения результата и эффектов здесь нет намеренно:
// эти значения считает только сервер.
type submitChoiceRequest struct {
	NodeID         string `json:"nodeId"`
	ChoiceID       string `json:"choiceId"`
	IdempotencyKey string `json:"idempotencyKey"`
}

// transitionResponse — результат одного шага прохождения.
type transitionResponse struct {
	AttemptID      string             `json:"attemptId"`
	Status         string             `json:"status"`
	Score          int                `json:"score"`
	Outcome        string             `json:"outcome,omitempty"`
	AcceptedChoice acceptedChoice     `json:"acceptedChoice"`
	Consequence    attemptConsequence `json:"consequence"`
	RevealedNodes  []attemptNode      `json:"revealedNodes"`
	CurrentNodeID  string             `json:"currentNodeId"`
	CompletedAt    *string            `json:"completedAt,omitempty"`
}

type acceptedChoice struct {
	NodeID   string `json:"nodeId"`
	ChoiceID string `json:"choiceId"`
	Label    string `json:"label"`
}

func transitionResponseOf(transition attempt.Transition) transitionResponse {
	response := transitionResponse{
		AttemptID: string(transition.AttemptID),
		Status:    string(transition.Status),
		Score:     transition.Score,
		Outcome:   string(transition.Outcome),
		AcceptedChoice: acceptedChoice{
			NodeID:   string(transition.Accepted.NodeID),
			ChoiceID: string(transition.Accepted.ChoiceID),
			Label:    transition.Accepted.Label,
		},
		Consequence: attemptConsequence{
			Severity:      string(transition.Consequence.Severity),
			Title:         transition.Consequence.Title,
			Explanation:   transition.Consequence.Explanation,
			RealWorldRule: transition.Consequence.RealWorldRule,
		},
		RevealedNodes: attemptNodesOf(transition.RevealedNodes),
		CurrentNodeID: string(transition.CurrentNodeID),
	}

	if transition.CompletedAt != nil {
		completedAt := transition.CompletedAt.Format(time.RFC3339)
		response.CompletedAt = &completedAt
	}

	return response
}

func attemptNodesOf(nodes []attempt.PublicNode) []attemptNode {
	converted := make([]attemptNode, 0, len(nodes))

	for _, node := range nodes {
		item := attemptNode{
			ID:     string(node.ID),
			Type:   string(node.Type),
			Sender: node.Sender,
			Text:   node.Text,
			Prompt: node.Prompt,
		}

		if len(node.Choices) > 0 {
			item.Choices = make([]attemptChoice, 0, len(node.Choices))
			for _, choice := range node.Choices {
				item.Choices = append(item.Choices, attemptChoice{
					ID:    string(choice.ID),
					Label: choice.Label,
				})
			}
		}

		if node.Outcome != nil {
			item.Outcome = &attemptOutcome{
				Type:        string(node.Outcome.Type),
				Title:       node.Outcome.Title,
				Explanation: node.Outcome.Explanation,
			}
		}

		converted = append(converted, item)
	}

	return converted
}

func attemptDecisionsOf(decisions []attempt.PublicDecision) []attemptDecision {
	converted := make([]attemptDecision, 0, len(decisions))

	for _, decision := range decisions {
		converted = append(converted, attemptDecision{
			NodeID:   string(decision.NodeID),
			ChoiceID: string(decision.ChoiceID),
			Label:    decision.ChoiceLabel,
			Consequence: attemptConsequence{
				Severity:      string(decision.Consequence.Severity),
				Title:         decision.Consequence.Title,
				Explanation:   decision.Consequence.Explanation,
				RealWorldRule: decision.Consequence.RealWorldRule,
			},
		})
	}

	return converted
}
