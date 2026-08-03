package attempt

import (
	"fmt"
	"time"

	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/scenario"
)

// Типы ниже описывают всё, что попытка вообще может показать пользователю.
//
// Скрытыми остаются переходы, дельты результата, критичность, метки риска и
// эффекты навыков: зная их заранее, пользователь видел бы правильный ответ.

// PublicChoice — доступное действие в текущем узле решения.
type PublicChoice struct {
	ID    scenario.ChoiceID
	Label string
}

// PublicOutcome — итог сценария в терминальном узле.
type PublicOutcome struct {
	Type        scenario.OutcomeType
	Title       string
	Explanation string
}

// PublicNode — раскрытый узел сценария.
type PublicNode struct {
	ID      scenario.NodeID
	Type    scenario.NodeType
	Sender  string
	Text    string
	Prompt  string
	Choices []PublicChoice
	Outcome *PublicOutcome
}

// PublicDecision — уже сделанный выбор вместе с показанным объяснением.
type PublicDecision struct {
	NodeID      scenario.NodeID
	ChoiceID    scenario.ChoiceID
	ChoiceLabel string
	Consequence scenario.Consequence
}

// ScenarioRef — краткая ссылка на сценарий попытки.
type ScenarioRef struct {
	ID      scenario.ID
	Version scenario.Version
	Slug    string
	Role    scenario.Role
	Title   string
}

// View — публичное состояние попытки. По нему фронтенд полностью
// восстанавливает экран после перезагрузки страницы.
type View struct {
	ID            ID
	Scenario      ScenarioRef
	Status        Status
	Score         int
	Outcome       scenario.OutcomeType
	CurrentNodeID scenario.NodeID
	RevealedNodes []PublicNode
	Decisions     []PublicDecision
	StartedAt     time.Time
	UpdatedAt     time.Time
	CompletedAt   *time.Time
}

// AcceptedChoice — выбор, который сервер принял и применил.
type AcceptedChoice struct {
	NodeID   scenario.NodeID
	ChoiceID scenario.ChoiceID
	Label    string
}

// Transition — результат одного шага прохождения.
//
// Состояние берётся из снимка, сохранённого вместе с решением, поэтому повтор
// запроса с тем же ключом возвращает точно такой же ответ.
type Transition struct {
	AttemptID     ID
	Status        Status
	Score         int
	Outcome       scenario.OutcomeType
	Accepted      AcceptedChoice
	Consequence   scenario.Consequence
	RevealedNodes []PublicNode
	CurrentNodeID scenario.NodeID
	CompletedAt   *time.Time
}

// transitionOf собирает публичный результат шага по записанному решению.
func transitionOf(source Attempt, definition scenario.Scenario, decision Decision) (Transition, error) {
	revealedNodes := make([]PublicNode, 0, len(decision.RevealedNodeIDs))

	for _, nodeID := range decision.RevealedNodeIDs {
		node, found := definition.Node(nodeID)
		if !found {
			return Transition{}, fmt.Errorf("%w: узел %q сценария %q", ErrBrokenScenario, nodeID, definition.ID)
		}

		revealedNodes = append(revealedNodes, publicNodeOf(node))
	}

	transition := Transition{
		AttemptID: source.ID,
		Status:    StatusInProgress,
		Score:     decision.ScoreAfter,
		Accepted: AcceptedChoice{
			NodeID:   decision.NodeID,
			ChoiceID: decision.ChoiceID,
			Label:    decision.ChoiceLabel,
		},
		Consequence:   decision.Consequence,
		RevealedNodes: revealedNodes,
		CurrentNodeID: decision.ResultingNodeID,
	}

	if decision.Completed {
		transition.Status = StatusCompleted
		transition.Outcome = decision.Outcome
		transition.CompletedAt = source.CompletedAt
	}

	return transition, nil
}

// viewOf собирает публичное состояние попытки по её истории и сценарию.
func viewOf(source Attempt, definition scenario.Scenario) (View, error) {
	revealedNodes := make([]PublicNode, 0, len(source.RevealedNodeIDs))

	for _, nodeID := range source.RevealedNodeIDs {
		node, found := definition.Node(nodeID)
		if !found {
			return View{}, fmt.Errorf("%w: узел %q сценария %q", ErrBrokenScenario, nodeID, definition.ID)
		}

		revealedNodes = append(revealedNodes, publicNodeOf(node))
	}

	decisions := make([]PublicDecision, 0, len(source.Decisions))
	for _, decision := range source.Decisions {
		decisions = append(decisions, PublicDecision{
			NodeID:      decision.NodeID,
			ChoiceID:    decision.ChoiceID,
			ChoiceLabel: decision.ChoiceLabel,
			Consequence: decision.Consequence,
		})
	}

	return View{
		ID: source.ID,
		Scenario: ScenarioRef{
			ID:      definition.ID,
			Version: source.ScenarioVersion,
			Slug:    definition.Slug,
			Role:    definition.Role,
			Title:   definition.Title,
		},
		Status:        source.Status,
		Score:         source.Score,
		Outcome:       source.Outcome,
		CurrentNodeID: source.CurrentNodeID,
		RevealedNodes: revealedNodes,
		Decisions:     decisions,
		StartedAt:     source.StartedAt,
		UpdatedAt:     source.UpdatedAt,
		CompletedAt:   source.CompletedAt,
	}, nil
}

// publicNodeOf оставляет от узла только то, что можно показать пользователю.
func publicNodeOf(node scenario.Node) PublicNode {
	public := PublicNode{
		ID:     node.ID,
		Type:   node.Type,
		Sender: node.Sender,
		Text:   node.Text,
		Prompt: node.DecisionPrompt,
	}

	if len(node.Choices) > 0 {
		public.Choices = make([]PublicChoice, 0, len(node.Choices))
		for _, choice := range node.Choices {
			// Наружу уходят только идентификатор и подпись: следующий узел,
			// дельта результата и эффекты остаются на сервере.
			public.Choices = append(public.Choices, PublicChoice{ID: choice.ID, Label: choice.Label})
		}
	}

	if node.TerminalOutcome != nil {
		public.Outcome = &PublicOutcome{
			Type:        node.TerminalOutcome.Type,
			Title:       node.TerminalOutcome.Title,
			Explanation: node.TerminalOutcome.Explanation,
		}
	}

	return public
}
