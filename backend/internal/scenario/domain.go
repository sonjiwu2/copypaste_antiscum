// Package scenario содержит доменную модель учебного сценария и проверку его графа.
//
// Пакет не зависит от HTTP, хранилищ и транспортных DTO: сценарий описывается
// только доменными типами.
package scenario

import "slices"

// ID — устойчивый идентификатор сценария.
type ID string

// Version — версия содержимого сценария. Начатая попытка закрепляется
// за конкретной версией, поэтому правка фикстуры не меняет уже идущее прохождение.
type Version int

// Role — роль, в которой пользователь проходит сценарий.
type Role string

// Роли, поддерживаемые тренажёром.
const (
	RoleBuyer  Role = "buyer"
	RoleSeller Role = "seller"
)

// Valid сообщает, поддерживается ли роль.
func (r Role) Valid() bool {
	return r == RoleBuyer || r == RoleSeller
}

// Difficulty — субъективная сложность сценария для пользователя.
type Difficulty string

// Уровни сложности.
const (
	DifficultyEasy   Difficulty = "easy"
	DifficultyMedium Difficulty = "medium"
	DifficultyHard   Difficulty = "hard"
)

// Valid сообщает, поддерживается ли уровень сложности.
func (d Difficulty) Valid() bool {
	return d == DifficultyEasy || d == DifficultyMedium || d == DifficultyHard
}

// NodeID — идентификатор узла внутри сценария.
type NodeID string

// ChoiceID — идентификатор варианта выбора внутри узла решения.
type ChoiceID string

// NodeType — тип узла сценария.
type NodeType string

// Типы узлов MVP.
const (
	NodeTypeMessage  NodeType = "message"
	NodeTypeDecision NodeType = "decision"
	NodeTypeTerminal NodeType = "terminal"
)

// Valid сообщает, поддерживается ли тип узла.
func (t NodeType) Valid() bool {
	return t == NodeTypeMessage || t == NodeTypeDecision || t == NodeTypeTerminal
}

// OutcomeType — характер финала сценария.
type OutcomeType string

// Возможные финалы.
const (
	OutcomeSafe   OutcomeType = "safe"
	OutcomeUnsafe OutcomeType = "unsafe"
)

// Valid сообщает, поддерживается ли тип финала.
func (o OutcomeType) Valid() bool {
	return o == OutcomeSafe || o == OutcomeUnsafe
}

// Severity — насколько опасным оказался выбор пользователя.
type Severity string

// Степени опасности выбора.
const (
	SeveritySafe      Severity = "safe"
	SeverityWarning   Severity = "warning"
	SeverityDangerous Severity = "dangerous"
)

// Valid сообщает, поддерживается ли степень опасности.
func (s Severity) Valid() bool {
	return s == SeveritySafe || s == SeverityWarning || s == SeverityDangerous
}

// Criticality — вес ошибки для будущего подсчёта прогресса.
// Значение внутреннее и клиенту не отдаётся.
type Criticality string

// Уровни критичности выбора.
const (
	CriticalityLow    Criticality = "low"
	CriticalityMedium Criticality = "medium"
	CriticalityHigh   Criticality = "high"
)

// Valid сообщает, поддерживается ли уровень критичности.
func (c Criticality) Valid() bool {
	return c == CriticalityLow || c == CriticalityMedium || c == CriticalityHigh
}

// RiskTag — признак мошеннической схемы, к которой относится выбор.
type RiskTag string

// SkillCode — код навыка, который тренирует выбор.
type SkillCode string

// SkillEffect — изменение навыка после выбора. Значение внутреннее.
type SkillEffect struct {
	Skill SkillCode
	Delta int
}

// Consequence — учебное объяснение последствий выбора.
// Раскрывается пользователю только после того, как выбор сделан.
type Consequence struct {
	Severity      Severity
	Title         string
	Explanation   string
	RealWorldRule string
}

// Outcome — публичный итог сценария в терминальном узле.
type Outcome struct {
	Type        OutcomeType
	Title       string
	Explanation string
}

// Choice — один вариант действия в узле решения.
//
// SafetyScore, Criticality, RiskTags и SkillEffects — серверные данные:
// они участвуют в подсчёте результата и не публикуются до выбора.
type Choice struct {
	ID           ChoiceID
	Label        string
	NextNodeID   NodeID
	SafetyScore  int
	Criticality  Criticality
	RiskTags     []RiskTag
	SkillEffects []SkillEffect
	Consequence  Consequence
}

// Node — узел сценария. Набор заполненных полей определяется типом узла
// и проверяется валидатором.
type Node struct {
	ID              NodeID
	Type            NodeType
	Sender          string
	Text            string
	NextNodeID      NodeID
	DecisionPrompt  string
	Choices         []Choice
	TerminalOutcome *Outcome
}

// Scenario — проверенный граф учебного сценария.
// Значение создаётся только через New, поэтому невалидного сценария не существует.
type Scenario struct {
	ID               ID
	Version          Version
	Slug             string
	Role             Role
	Title            string
	Description      string
	Difficulty       Difficulty
	EstimatedMinutes int
	StartNodeID      NodeID
	IsActive         bool

	nodes map[NodeID]Node
	order []NodeID
}

// Draft — исходное описание сценария до проверки.
// Узлы передаются слайсом, чтобы поймать повторяющиеся идентификаторы.
type Draft struct {
	ID               ID
	Version          Version
	Slug             string
	Role             Role
	Title            string
	Description      string
	Difficulty       Difficulty
	EstimatedMinutes int
	StartNodeID      NodeID
	IsActive         bool
	Nodes            []Node
}

// New собирает сценарий из черновика и проверяет его граф.
// При нарушении правил возвращает ValidationErrors и пустой сценарий.
func New(draft Draft) (Scenario, error) {
	scenario := Scenario{
		ID:               draft.ID,
		Version:          draft.Version,
		Slug:             draft.Slug,
		Role:             draft.Role,
		Title:            draft.Title,
		Description:      draft.Description,
		Difficulty:       draft.Difficulty,
		EstimatedMinutes: draft.EstimatedMinutes,
		StartNodeID:      draft.StartNodeID,
		IsActive:         draft.IsActive,
		nodes:            make(map[NodeID]Node, len(draft.Nodes)),
		order:            make([]NodeID, 0, len(draft.Nodes)),
	}

	duplicates := make([]ValidationError, 0)

	for _, node := range draft.Nodes {
		if _, exists := scenario.nodes[node.ID]; exists {
			duplicates = append(duplicates, ValidationError{
				ScenarioID: draft.ID,
				NodeID:     node.ID,
				Rule:       RuleDuplicateNodeID,
				Detail:     "идентификатор узла встречается больше одного раза",
			})

			continue
		}

		scenario.nodes[node.ID] = node
		scenario.order = append(scenario.order, node.ID)
	}

	if err := validate(scenario, duplicates); err != nil {
		return Scenario{}, err
	}

	return scenario, nil
}

// Node возвращает узел сценария по идентификатору.
//
// Возвращается независимая копия: сценарий живёт в общем хранилище, и
// вызывающий код не должен иметь возможности изменить его через слайсы узла.
func (s Scenario) Node(id NodeID) (Node, bool) {
	node, found := s.nodes[id]
	if !found {
		return Node{}, false
	}

	return node.clone(), true
}

// NodeCount возвращает количество узлов сценария.
func (s Scenario) NodeCount() int {
	return len(s.nodes)
}

// Choice возвращает вариант выбора указанного узла решения.
func (n Node) Choice(id ChoiceID) (Choice, bool) {
	for _, choice := range n.Choices {
		if choice.ID == id {
			return choice.clone(), true
		}
	}

	return Choice{}, false
}

func (n Node) clone() Node {
	cloned := n

	if n.Choices != nil {
		cloned.Choices = make([]Choice, len(n.Choices))
		for i, choice := range n.Choices {
			cloned.Choices[i] = choice.clone()
		}
	}

	if n.TerminalOutcome != nil {
		outcome := *n.TerminalOutcome
		cloned.TerminalOutcome = &outcome
	}

	return cloned
}

func (c Choice) clone() Choice {
	cloned := c
	cloned.RiskTags = slices.Clone(c.RiskTags)
	cloned.SkillEffects = slices.Clone(c.SkillEffects)

	return cloned
}

// transitions возвращает узлы, в которые можно перейти из текущего.
func (n Node) transitions() []NodeID {
	switch n.Type {
	case NodeTypeMessage:
		if n.NextNodeID == "" {
			return nil
		}

		return []NodeID{n.NextNodeID}
	case NodeTypeDecision:
		targets := make([]NodeID, 0, len(n.Choices))
		for _, choice := range n.Choices {
			if choice.NextNodeID != "" {
				targets = append(targets, choice.NextNodeID)
			}
		}

		return targets
	case NodeTypeTerminal:
		return nil
	default:
		return nil
	}
}
