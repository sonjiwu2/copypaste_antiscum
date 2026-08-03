package scenario

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/fs"
	"sort"
)

// fixturePattern — маска файлов сценариев внутри переданной файловой системы.
const fixturePattern = "*.json"

// Структуры ниже описывают формат файла сценария.
// Они отделены от доменной модели, чтобы изменение формата хранения
// не тянуло теги сериализации в домен.
type fixtureScenario struct {
	ID               string        `json:"id"`
	Version          int           `json:"version"`
	Slug             string        `json:"slug"`
	Role             string        `json:"role"`
	Title            string        `json:"title"`
	Description      string        `json:"description"`
	Difficulty       string        `json:"difficulty"`
	EstimatedMinutes int           `json:"estimatedMinutes"`
	StartNodeID      string        `json:"startNodeId"`
	IsActive         bool          `json:"isActive"`
	Nodes            []fixtureNode `json:"nodes"`
}

type fixtureNode struct {
	ID             string          `json:"id"`
	Type           string          `json:"type"`
	Sender         string          `json:"sender,omitempty"`
	Text           string          `json:"text,omitempty"`
	NextNodeID     string          `json:"nextNodeId,omitempty"`
	DecisionPrompt string          `json:"decisionPrompt,omitempty"`
	Choices        []fixtureChoice `json:"choices,omitempty"`
	Outcome        *fixtureOutcome `json:"outcome,omitempty"`
}

type fixtureChoice struct {
	ID           string               `json:"id"`
	Label        string               `json:"label"`
	NextNodeID   string               `json:"nextNodeId"`
	SafetyScore  int                  `json:"safetyScore"`
	Criticality  string               `json:"criticality"`
	RiskTags     []string             `json:"riskTags,omitempty"`
	SkillEffects []fixtureSkillEffect `json:"skillEffects,omitempty"`
	Consequence  fixtureConsequence   `json:"consequence"`
}

type fixtureSkillEffect struct {
	Skill string `json:"skill"`
	Delta int    `json:"delta"`
}

type fixtureConsequence struct {
	Severity      string `json:"severity"`
	Title         string `json:"title"`
	Explanation   string `json:"explanation"`
	RealWorldRule string `json:"realWorldRule"`
}

type fixtureOutcome struct {
	Type        string `json:"type"`
	Title       string `json:"title"`
	Explanation string `json:"explanation"`
}

// LoadFS читает и проверяет все сценарии из переданной файловой системы.
//
// Ошибка любого файла останавливает загрузку целиком: приложение не должно
// запускаться с частично исправным каталогом.
func LoadFS(fsys fs.FS) ([]Scenario, error) {
	names, err := fs.Glob(fsys, fixturePattern)
	if err != nil {
		return nil, fmt.Errorf("найти файлы сценариев: %w", err)
	}

	if len(names) == 0 {
		return nil, fmt.Errorf("не найдено ни одного файла сценария по маске %q", fixturePattern)
	}

	// Порядок обхода файловой системы не гарантирован, а каталог должен
	// собираться одинаково при каждом запуске.
	sort.Strings(names)

	scenarios := make([]Scenario, 0, len(names))

	for _, name := range names {
		loaded, err := loadFile(fsys, name)
		if err != nil {
			return nil, err
		}

		scenarios = append(scenarios, loaded)
	}

	return scenarios, nil
}

func loadFile(fsys fs.FS, name string) (Scenario, error) {
	raw, err := fs.ReadFile(fsys, name)
	if err != nil {
		return Scenario{}, fmt.Errorf("прочитать файл сценария %q: %w", name, err)
	}

	decoder := json.NewDecoder(bytes.NewReader(raw))
	// Неизвестное поле почти всегда означает опечатку в фикстуре,
	// поэтому такой файл отклоняется, а не читается частично.
	decoder.DisallowUnknownFields()

	var file fixtureScenario
	if err := decoder.Decode(&file); err != nil {
		return Scenario{}, fmt.Errorf("разобрать файл сценария %q: %w", name, err)
	}

	built, err := New(file.draft())
	if err != nil {
		return Scenario{}, fmt.Errorf("файл сценария %q не прошёл проверку: %w", name, err)
	}

	return built, nil
}

func (f fixtureScenario) draft() Draft {
	nodes := make([]Node, 0, len(f.Nodes))
	for _, node := range f.Nodes {
		nodes = append(nodes, node.domainNode())
	}

	return Draft{
		ID:               ID(f.ID),
		Version:          Version(f.Version),
		Slug:             f.Slug,
		Role:             Role(f.Role),
		Title:            f.Title,
		Description:      f.Description,
		Difficulty:       Difficulty(f.Difficulty),
		EstimatedMinutes: f.EstimatedMinutes,
		StartNodeID:      NodeID(f.StartNodeID),
		IsActive:         f.IsActive,
		Nodes:            nodes,
	}
}

func (n fixtureNode) domainNode() Node {
	node := Node{
		ID:             NodeID(n.ID),
		Type:           NodeType(n.Type),
		Sender:         n.Sender,
		Text:           n.Text,
		NextNodeID:     NodeID(n.NextNodeID),
		DecisionPrompt: n.DecisionPrompt,
	}

	if len(n.Choices) > 0 {
		node.Choices = make([]Choice, 0, len(n.Choices))
		for _, choice := range n.Choices {
			node.Choices = append(node.Choices, choice.domainChoice())
		}
	}

	if n.Outcome != nil {
		node.TerminalOutcome = &Outcome{
			Type:        OutcomeType(n.Outcome.Type),
			Title:       n.Outcome.Title,
			Explanation: n.Outcome.Explanation,
		}
	}

	return node
}

func (c fixtureChoice) domainChoice() Choice {
	choice := Choice{
		ID:          ChoiceID(c.ID),
		Label:       c.Label,
		NextNodeID:  NodeID(c.NextNodeID),
		SafetyScore: c.SafetyScore,
		Criticality: Criticality(c.Criticality),
		Consequence: Consequence{
			Severity:      Severity(c.Consequence.Severity),
			Title:         c.Consequence.Title,
			Explanation:   c.Consequence.Explanation,
			RealWorldRule: c.Consequence.RealWorldRule,
		},
	}

	if len(c.RiskTags) > 0 {
		choice.RiskTags = make([]RiskTag, 0, len(c.RiskTags))
		for _, tag := range c.RiskTags {
			choice.RiskTags = append(choice.RiskTags, RiskTag(tag))
		}
	}

	if len(c.SkillEffects) > 0 {
		choice.SkillEffects = make([]SkillEffect, 0, len(c.SkillEffects))
		for _, effect := range c.SkillEffects {
			choice.SkillEffects = append(choice.SkillEffects, SkillEffect{
				Skill: SkillCode(effect.Skill),
				Delta: effect.Delta,
			})
		}
	}

	return choice
}
