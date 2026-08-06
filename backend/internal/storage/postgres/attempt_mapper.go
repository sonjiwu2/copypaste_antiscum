package postgres

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/attempt"
	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/profile"
	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/scenario"
)

// Структуры ниже описывают формат хранения вложенных значений в JSONB.
// Они отделены от доменной модели, чтобы изменение схемы хранения
// не тянуло теги сериализации в домен.

type consequenceRow struct {
	Severity      string `json:"severity"`
	Title         string `json:"title"`
	Explanation   string `json:"explanation"`
	RealWorldRule string `json:"realWorldRule"`
}

type skillEffectRow struct {
	Skill string `json:"skill"`
	Delta int    `json:"delta"`
}

// attemptRow — строка таблицы attempts.
type attemptRow struct {
	ID                  string
	ProfileID           string
	ScenarioID          string
	ScenarioVersion     int32
	CurrentNodeID       string
	Status              string
	Score               int32
	Outcome             *string
	StartedAt           time.Time
	UpdatedAt           time.Time
	CompletedAt         *time.Time
	RevealedNodeIDs     []string
	AppliedSkillEffects []byte
	Version             int64
}

// decisionRow — строка таблицы attempt_decisions.
type decisionRow struct {
	Ordinal         int32
	NodeID          string
	ChoiceID        string
	ChoiceLabel     string
	IdempotencyKey  string
	Consequence     []byte
	Criticality     string
	RiskTags        []string
	SkillEffects    []byte
	ScoreDelta      int32
	CreatedAt       time.Time
	RevealedNodeIDs []string
	ResultingNodeID string
	Completed       bool
	Outcome         *string
	ScoreAfter      int32
}

// attemptRowOf раскладывает агрегат по колонкам.
func attemptRowOf(source attempt.Attempt) (attemptRow, error) {
	appliedSkillEffects, err := marshalSkillTotals(source.AppliedSkillEffects)
	if err != nil {
		return attemptRow{}, err
	}

	return attemptRow{
		ID:                  string(source.ID),
		ProfileID:           string(source.ProfileID),
		ScenarioID:          string(source.ScenarioID),
		ScenarioVersion:     int32(source.ScenarioVersion),
		CurrentNodeID:       string(source.CurrentNodeID),
		Status:              string(source.Status),
		Score:               int32(source.Score),
		Outcome:             nullableString(string(source.Outcome)),
		StartedAt:           source.StartedAt,
		UpdatedAt:           source.UpdatedAt,
		CompletedAt:         source.CompletedAt,
		RevealedNodeIDs:     stringsOf(source.RevealedNodeIDs),
		AppliedSkillEffects: appliedSkillEffects,
		Version:             int64(source.Version),
	}, nil
}

// attemptOf собирает агрегат из строки таблицы и уже загруженных решений.
//
// Пустые массивы восстанавливаются как непустые срезы, а карта эффектов —
// как непустая карта: проекции попытки рассчитывают, что этих значений
// достаточно для чтения без дополнительных проверок на nil.
func attemptOf(row attemptRow, decisions []attempt.Decision) (attempt.Attempt, error) {
	appliedSkillEffects, err := unmarshalSkillTotals(row.AppliedSkillEffects)
	if err != nil {
		return attempt.Attempt{}, err
	}

	if decisions == nil {
		decisions = []attempt.Decision{}
	}

	return attempt.Attempt{
		ID:                  attempt.ID(row.ID),
		ProfileID:           profile.ID(row.ProfileID),
		ScenarioID:          scenario.ID(row.ScenarioID),
		ScenarioVersion:     scenario.Version(row.ScenarioVersion),
		CurrentNodeID:       scenario.NodeID(row.CurrentNodeID),
		Status:              attempt.Status(row.Status),
		Score:               int(row.Score),
		Outcome:             scenario.OutcomeType(stringValue(row.Outcome)),
		StartedAt:           row.StartedAt,
		UpdatedAt:           row.UpdatedAt,
		CompletedAt:         row.CompletedAt,
		RevealedNodeIDs:     nodeIDsOf(row.RevealedNodeIDs),
		Decisions:           decisions,
		AppliedSkillEffects: appliedSkillEffects,
		Version:             int(row.Version),
	}, nil
}

// decisionRowOf раскладывает решение по колонкам. Порядковый номер задаётся
// хранилищем: в домене позиция решения выражена индексом в срезе.
func decisionRowOf(source attempt.Decision, ordinal int) (decisionRow, error) {
	consequence, err := json.Marshal(consequenceRow{
		Severity:      string(source.Consequence.Severity),
		Title:         source.Consequence.Title,
		Explanation:   source.Consequence.Explanation,
		RealWorldRule: source.Consequence.RealWorldRule,
	})
	if err != nil {
		return decisionRow{}, fmt.Errorf("сериализовать последствие решения: %w", err)
	}

	skillEffects, err := marshalSkillEffects(source.SkillEffects)
	if err != nil {
		return decisionRow{}, err
	}

	return decisionRow{
		Ordinal:         int32(ordinal),
		NodeID:          string(source.NodeID),
		ChoiceID:        string(source.ChoiceID),
		ChoiceLabel:     source.ChoiceLabel,
		IdempotencyKey:  string(source.IdempotencyKey),
		Consequence:     consequence,
		Criticality:     string(source.Criticality),
		RiskTags:        stringsOf(source.RiskTags),
		SkillEffects:    skillEffects,
		ScoreDelta:      int32(source.ScoreDelta),
		CreatedAt:       source.CreatedAt,
		RevealedNodeIDs: stringsOf(source.RevealedNodeIDs),
		ResultingNodeID: string(source.ResultingNodeID),
		Completed:       source.Completed,
		Outcome:         nullableString(string(source.Outcome)),
		ScoreAfter:      int32(source.ScoreAfter),
	}, nil
}

// decisionOf собирает решение из строки таблицы.
func decisionOf(row decisionRow) (attempt.Decision, error) {
	var consequence consequenceRow
	if err := json.Unmarshal(row.Consequence, &consequence); err != nil {
		return attempt.Decision{}, fmt.Errorf("разобрать последствие решения: %w", err)
	}

	skillEffects, err := unmarshalSkillEffects(row.SkillEffects)
	if err != nil {
		return attempt.Decision{}, err
	}

	riskTags := make([]scenario.RiskTag, 0, len(row.RiskTags))
	for _, tag := range row.RiskTags {
		riskTags = append(riskTags, scenario.RiskTag(tag))
	}

	return attempt.Decision{
		NodeID:         scenario.NodeID(row.NodeID),
		ChoiceID:       scenario.ChoiceID(row.ChoiceID),
		ChoiceLabel:    row.ChoiceLabel,
		IdempotencyKey: attempt.IdempotencyKey(row.IdempotencyKey),
		Consequence: scenario.Consequence{
			Severity:      scenario.Severity(consequence.Severity),
			Title:         consequence.Title,
			Explanation:   consequence.Explanation,
			RealWorldRule: consequence.RealWorldRule,
		},
		Criticality:     scenario.Criticality(row.Criticality),
		RiskTags:        riskTags,
		SkillEffects:    skillEffects,
		ScoreDelta:      int(row.ScoreDelta),
		CreatedAt:       row.CreatedAt,
		RevealedNodeIDs: nodeIDsOf(row.RevealedNodeIDs),
		ResultingNodeID: scenario.NodeID(row.ResultingNodeID),
		Completed:       row.Completed,
		Outcome:         scenario.OutcomeType(stringValue(row.Outcome)),
		ScoreAfter:      int(row.ScoreAfter),
	}, nil
}

func marshalSkillTotals(totals map[scenario.SkillCode]int) ([]byte, error) {
	// Пустая карта сохраняется как {}, а не как null: колонка объявлена NOT NULL.
	plain := make(map[string]int, len(totals))
	for skill, value := range totals {
		plain[string(skill)] = value
	}

	encoded, err := json.Marshal(plain)
	if err != nil {
		return nil, fmt.Errorf("сериализовать накопленные эффекты навыков: %w", err)
	}

	return encoded, nil
}

func unmarshalSkillTotals(raw []byte) (map[scenario.SkillCode]int, error) {
	plain := map[string]int{}

	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &plain); err != nil {
			return nil, fmt.Errorf("разобрать накопленные эффекты навыков: %w", err)
		}
	}

	totals := make(map[scenario.SkillCode]int, len(plain))
	for skill, value := range plain {
		totals[scenario.SkillCode(skill)] = value
	}

	return totals, nil
}

func marshalSkillEffects(effects []scenario.SkillEffect) ([]byte, error) {
	plain := make([]skillEffectRow, 0, len(effects))
	for _, effect := range effects {
		plain = append(plain, skillEffectRow{Skill: string(effect.Skill), Delta: effect.Delta})
	}

	encoded, err := json.Marshal(plain)
	if err != nil {
		return nil, fmt.Errorf("сериализовать эффекты навыков решения: %w", err)
	}

	return encoded, nil
}

func unmarshalSkillEffects(raw []byte) ([]scenario.SkillEffect, error) {
	var plain []skillEffectRow

	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &plain); err != nil {
			return nil, fmt.Errorf("разобрать эффекты навыков решения: %w", err)
		}
	}

	effects := make([]scenario.SkillEffect, 0, len(plain))
	for _, effect := range plain {
		effects = append(effects, scenario.SkillEffect{
			Skill: scenario.SkillCode(effect.Skill),
			Delta: effect.Delta,
		})
	}

	return effects, nil
}

// stringsOf переводит доменные идентификаторы в массив строк для text[].
func stringsOf[T ~string](values []T) []string {
	converted := make([]string, 0, len(values))
	for _, value := range values {
		converted = append(converted, string(value))
	}

	return converted
}

func nodeIDsOf(values []string) []scenario.NodeID {
	converted := make([]scenario.NodeID, 0, len(values))
	for _, value := range values {
		converted = append(converted, scenario.NodeID(value))
	}

	return converted
}

// nullableString переводит пустое доменное значение в NULL.
// Пустая строка и NULL не должны сосуществовать: иначе ограничения
// согласованности статуса и итога пришлось бы дублировать в двух вариантах.
func nullableString(value string) *string {
	if value == "" {
		return nil
	}

	return &value
}

func stringValue(value *string) string {
	if value == nil {
		return ""
	}

	return *value
}
