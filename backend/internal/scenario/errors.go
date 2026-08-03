package scenario

import (
	"errors"
	"fmt"
	"strings"
)

// Доменные ошибки каталога сценариев.
var (
	// ErrNotFound возвращается и для отсутствующего, и для отключённого
	// сценария: клиенту не сообщается о существовании скрытого содержимого.
	ErrNotFound = errors.New("сценарий не найден")

	// ErrUnsupportedRole сообщает о запросе с неизвестной ролью.
	ErrUnsupportedRole = errors.New("роль не поддерживается")
)

// ValidationRule — нарушенное правило описания сценария.
// Код правила стабилен, поэтому тесты и логи не зависят от текста сообщения.
type ValidationRule string

// Правила проверки графа сценария.
const (
	RuleScenarioIDRequired            ValidationRule = "scenario_id_required"
	RuleScenarioVersionInvalid        ValidationRule = "scenario_version_invalid"
	RuleScenarioRoleUnsupported       ValidationRule = "scenario_role_unsupported"
	RuleScenarioTitleRequired         ValidationRule = "scenario_title_required"
	RuleScenarioDifficultyUnsupported ValidationRule = "scenario_difficulty_unsupported"
	RuleStartNodeMissing              ValidationRule = "start_node_missing"
	RuleDuplicateNodeID               ValidationRule = "duplicate_node_id"
	RuleNodeIDRequired                ValidationRule = "node_id_required"
	RuleNodeTypeUnsupported           ValidationRule = "node_type_unsupported"
	RuleMessageNodeWithoutNext        ValidationRule = "message_node_without_next"
	RuleMessageNodeWithChoices        ValidationRule = "message_node_with_choices"
	RuleDecisionNodeTooFewChoices     ValidationRule = "decision_node_too_few_choices"
	RuleDuplicateChoiceID             ValidationRule = "duplicate_choice_id"
	RuleChoiceLabelRequired           ValidationRule = "choice_label_required"
	RuleTransitionTargetMissing       ValidationRule = "transition_target_missing"
	RuleTerminalNodeWithNext          ValidationRule = "terminal_node_with_next"
	RuleTerminalNodeWithChoices       ValidationRule = "terminal_node_with_choices"
	RuleTerminalNodeWithoutOutcome    ValidationRule = "terminal_node_without_outcome"
	RuleTerminalUnreachable           ValidationRule = "terminal_unreachable"
	RuleNodeUnreachable               ValidationRule = "node_unreachable"
	RuleGraphHasCycle                 ValidationRule = "graph_has_cycle"
	RuleConsequenceIncomplete         ValidationRule = "consequence_incomplete"
)

// ValidationError описывает одно нарушение с точным местом в сценарии.
type ValidationError struct {
	ScenarioID ID
	NodeID     NodeID
	ChoiceID   ChoiceID
	Rule       ValidationRule
	Detail     string
}

// Error возвращает описание нарушения для логов и запуска приложения.
// Текст предназначен разработчику и не публикуется через HTTP.
func (e ValidationError) Error() string {
	location := fmt.Sprintf("сценарий %q", e.ScenarioID)

	if e.NodeID != "" {
		location += fmt.Sprintf(", узел %q", e.NodeID)
	}

	if e.ChoiceID != "" {
		location += fmt.Sprintf(", выбор %q", e.ChoiceID)
	}

	return fmt.Sprintf("%s: [%s] %s", location, e.Rule, e.Detail)
}

// ValidationErrors — все нарушения одного сценария.
type ValidationErrors []ValidationError

// Error перечисляет нарушения по одному в строке.
func (e ValidationErrors) Error() string {
	messages := make([]string, len(e))
	for i, violation := range e {
		messages[i] = violation.Error()
	}

	return strings.Join(messages, "; ")
}

// Has сообщает, встречается ли указанное правило среди нарушений.
func (e ValidationErrors) Has(rule ValidationRule) bool {
	for _, violation := range e {
		if violation.Rule == rule {
			return true
		}
	}

	return false
}

// Rules возвращает коды нарушенных правил в порядке обнаружения.
func (e ValidationErrors) Rules() []ValidationRule {
	rules := make([]ValidationRule, len(e))
	for i, violation := range e {
		rules[i] = violation.Rule
	}

	return rules
}
