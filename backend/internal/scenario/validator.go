package scenario

import "fmt"

// minDecisionChoices — узел решения без выбора не является решением.
const minDecisionChoices = 2

// validate проверяет метаданные, узлы и связность графа.
// Параметр seed содержит нарушения, найденные ещё при сборке сценария.
func validate(scenario Scenario, seed []ValidationError) error {
	violations := &violationCollector{scenarioID: scenario.ID, found: seed}

	validateMetadata(scenario, violations)
	validateNodes(scenario, violations)
	validateGraph(scenario, violations)

	if len(violations.found) == 0 {
		return nil
	}

	return ValidationErrors(violations.found)
}

func validateMetadata(scenario Scenario, violations *violationCollector) {
	if scenario.ID == "" {
		violations.add(ValidationError{
			Rule:   RuleScenarioIDRequired,
			Detail: "идентификатор сценария обязателен",
		})
	}

	if scenario.Version < 1 {
		violations.add(ValidationError{
			Rule:   RuleScenarioVersionInvalid,
			Detail: fmt.Sprintf("версия должна быть положительной, получено %d", scenario.Version),
		})
	}

	if !scenario.Role.Valid() {
		violations.add(ValidationError{
			Rule:   RuleScenarioRoleUnsupported,
			Detail: fmt.Sprintf("роль %q не поддерживается", scenario.Role),
		})
	}

	if scenario.Title == "" {
		violations.add(ValidationError{
			Rule:   RuleScenarioTitleRequired,
			Detail: "заголовок сценария обязателен",
		})
	}

	if !scenario.Difficulty.Valid() {
		violations.add(ValidationError{
			Rule:   RuleScenarioDifficultyUnsupported,
			Detail: fmt.Sprintf("сложность %q не поддерживается", scenario.Difficulty),
		})
	}
}

func validateNodes(scenario Scenario, violations *violationCollector) {
	for _, nodeID := range scenario.order {
		node, _ := scenario.Node(nodeID)

		if node.ID == "" {
			violations.add(ValidationError{
				Rule:   RuleNodeIDRequired,
				Detail: "идентификатор узла обязателен",
			})

			continue
		}

		if !node.Type.Valid() {
			violations.add(ValidationError{
				NodeID: node.ID,
				Rule:   RuleNodeTypeUnsupported,
				Detail: fmt.Sprintf("тип узла %q не поддерживается", node.Type),
			})

			continue
		}

		switch node.Type {
		case NodeTypeMessage:
			validateMessageNode(node, violations)
		case NodeTypeDecision:
			validateDecisionNode(node, violations)
		case NodeTypeTerminal:
			validateTerminalNode(node, violations)
		}

		validateTransitionTargets(scenario, node, violations)
	}
}

func validateMessageNode(node Node, violations *violationCollector) {
	if node.NextNodeID == "" {
		violations.add(ValidationError{
			NodeID: node.ID,
			Rule:   RuleMessageNodeWithoutNext,
			Detail: "узел-сообщение обязан указывать следующий узел",
		})
	}

	if len(node.Choices) > 0 {
		violations.add(ValidationError{
			NodeID: node.ID,
			Rule:   RuleMessageNodeWithChoices,
			Detail: "узел-сообщение не может содержать варианты выбора",
		})
	}
}

func validateDecisionNode(node Node, violations *violationCollector) {
	if len(node.Choices) < minDecisionChoices {
		violations.add(ValidationError{
			NodeID: node.ID,
			Rule:   RuleDecisionNodeTooFewChoices,
			Detail: fmt.Sprintf("узел решения требует минимум %d варианта, найдено %d",
				minDecisionChoices, len(node.Choices)),
		})
	}

	seen := make(map[ChoiceID]struct{}, len(node.Choices))

	for _, choice := range node.Choices {
		if _, exists := seen[choice.ID]; exists {
			violations.add(ValidationError{
				NodeID:   node.ID,
				ChoiceID: choice.ID,
				Rule:     RuleDuplicateChoiceID,
				Detail:   "идентификатор выбора повторяется внутри узла",
			})
		}

		seen[choice.ID] = struct{}{}

		if choice.Label == "" {
			violations.add(ValidationError{
				NodeID:   node.ID,
				ChoiceID: choice.ID,
				Rule:     RuleChoiceLabelRequired,
				Detail:   "у варианта выбора должна быть подпись",
			})
		}

		validateConsequence(node.ID, choice, violations)
	}
}

// validateConsequence следит за учебной ценностью сценария: выбор без
// объяснения ничему не учит, поэтому такой сценарий считается неполным.
func validateConsequence(nodeID NodeID, choice Choice, violations *violationCollector) {
	consequence := choice.Consequence

	switch {
	case !consequence.Severity.Valid():
		violations.add(ValidationError{
			NodeID:   nodeID,
			ChoiceID: choice.ID,
			Rule:     RuleConsequenceIncomplete,
			Detail:   fmt.Sprintf("степень опасности %q не поддерживается", consequence.Severity),
		})
	case consequence.Title == "":
		violations.add(ValidationError{
			NodeID:   nodeID,
			ChoiceID: choice.ID,
			Rule:     RuleConsequenceIncomplete,
			Detail:   "у последствия должен быть заголовок",
		})
	case consequence.Explanation == "":
		violations.add(ValidationError{
			NodeID:   nodeID,
			ChoiceID: choice.ID,
			Rule:     RuleConsequenceIncomplete,
			Detail:   "у последствия должно быть объяснение",
		})
	}
}

func validateTerminalNode(node Node, violations *violationCollector) {
	if node.NextNodeID != "" {
		violations.add(ValidationError{
			NodeID: node.ID,
			Rule:   RuleTerminalNodeWithNext,
			Detail: "терминальный узел не может ссылаться на следующий узел",
		})
	}

	if len(node.Choices) > 0 {
		violations.add(ValidationError{
			NodeID: node.ID,
			Rule:   RuleTerminalNodeWithChoices,
			Detail: "терминальный узел не может содержать варианты выбора",
		})
	}

	if node.TerminalOutcome == nil {
		violations.add(ValidationError{
			NodeID: node.ID,
			Rule:   RuleTerminalNodeWithoutOutcome,
			Detail: "терминальный узел обязан содержать итог",
		})

		return
	}

	outcome := *node.TerminalOutcome

	if !outcome.Type.Valid() || outcome.Title == "" || outcome.Explanation == "" {
		violations.add(ValidationError{
			NodeID: node.ID,
			Rule:   RuleTerminalNodeWithoutOutcome,
			Detail: "итог должен содержать поддерживаемый тип, заголовок и объяснение",
		})
	}
}

// validateTransitionTargets проверяет инвариант «все переходы ведут в существующий узел».
func validateTransitionTargets(scenario Scenario, node Node, violations *violationCollector) {
	if node.Type == NodeTypeMessage && node.NextNodeID != "" {
		if _, found := scenario.Node(node.NextNodeID); !found {
			violations.add(ValidationError{
				NodeID: node.ID,
				Rule:   RuleTransitionTargetMissing,
				Detail: fmt.Sprintf("следующий узел %q не существует", node.NextNodeID),
			})
		}
	}

	if node.Type != NodeTypeDecision {
		return
	}

	for _, choice := range node.Choices {
		if _, found := scenario.Node(choice.NextNodeID); found {
			continue
		}

		violations.add(ValidationError{
			NodeID:   node.ID,
			ChoiceID: choice.ID,
			Rule:     RuleTransitionTargetMissing,
			Detail:   fmt.Sprintf("следующий узел %q не существует", choice.NextNodeID),
		})
	}
}

func validateGraph(scenario Scenario, violations *violationCollector) {
	if _, found := scenario.Node(scenario.StartNodeID); !found {
		violations.add(ValidationError{
			NodeID: scenario.StartNodeID,
			Rule:   RuleStartNodeMissing,
			Detail: "стартовый узел не найден среди узлов сценария",
		})

		// Без стартового узла достижимость считать не от чего.
		return
	}

	if cycleNodeID, hasCycle := findCycle(scenario); hasCycle {
		violations.add(ValidationError{
			NodeID: cycleNodeID,
			Rule:   RuleGraphHasCycle,
			Detail: "граф содержит цикл: в MVP сценарий обязан завершаться",
		})
	}

	reachable := reachableFrom(scenario, scenario.StartNodeID)

	for _, nodeID := range scenario.order {
		if _, visited := reachable[nodeID]; !visited {
			violations.add(ValidationError{
				NodeID: nodeID,
				Rule:   RuleNodeUnreachable,
				Detail: "узел недостижим из стартового узла",
			})
		}
	}

	if !hasReachableTerminal(scenario, reachable) {
		violations.add(ValidationError{
			Rule:   RuleTerminalUnreachable,
			Detail: "из стартового узла не достижим ни один терминальный узел",
		})
	}
}

func reachableFrom(scenario Scenario, start NodeID) map[NodeID]struct{} {
	reachable := map[NodeID]struct{}{start: {}}
	queue := []NodeID{start}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		node, found := scenario.Node(current)
		if !found {
			continue
		}

		for _, target := range node.transitions() {
			if _, visited := reachable[target]; visited {
				continue
			}

			if _, exists := scenario.Node(target); !exists {
				continue
			}

			reachable[target] = struct{}{}
			queue = append(queue, target)
		}
	}

	return reachable
}

func hasReachableTerminal(scenario Scenario, reachable map[NodeID]struct{}) bool {
	for nodeID := range reachable {
		node, found := scenario.Node(nodeID)
		if found && node.Type == NodeTypeTerminal {
			return true
		}
	}

	return false
}

// Цвета обхода в глубину: белый — не посещён, серый — в текущем пути,
// чёрный — полностью обработан. Серый повторно означает цикл.
type traversalColor int

const (
	colorWhite traversalColor = iota
	colorGray
	colorBlack
)

// findCycle возвращает узел, на котором замыкается первый найденный цикл.
func findCycle(scenario Scenario) (NodeID, bool) {
	colors := make(map[NodeID]traversalColor, scenario.NodeCount())

	var walk func(NodeID) (NodeID, bool)

	walk = func(nodeID NodeID) (NodeID, bool) {
		colors[nodeID] = colorGray

		node, found := scenario.Node(nodeID)
		if found {
			for _, target := range node.transitions() {
				if _, exists := scenario.Node(target); !exists {
					continue
				}

				switch colors[target] {
				case colorGray:
					return target, true
				case colorWhite:
					if cycleNodeID, hasCycle := walk(target); hasCycle {
						return cycleNodeID, true
					}
				case colorBlack:
				}
			}
		}

		colors[nodeID] = colorBlack

		return "", false
	}

	// Порядок объявления узлов делает результат детерминированным.
	for _, nodeID := range scenario.order {
		if colors[nodeID] != colorWhite {
			continue
		}

		if cycleNodeID, hasCycle := walk(nodeID); hasCycle {
			return cycleNodeID, true
		}
	}

	return "", false
}

// violationCollector накапливает нарушения и проставляет им идентификатор сценария.
type violationCollector struct {
	scenarioID ID
	found      []ValidationError
}

func (c *violationCollector) add(violation ValidationError) {
	violation.ScenarioID = c.scenarioID
	c.found = append(c.found, violation)
}
