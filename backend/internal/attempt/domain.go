// Package attempt содержит доменную модель одного прохождения сценария.
//
// Пакет не зависит от HTTP и хранилищ: правила перехода и подсчёта результата
// живут здесь, а сохранение — в адаптерах.
package attempt

import (
	"maps"
	"slices"
	"time"

	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/scenario"
)

// ID — идентификатор попытки. Аутентификации нет, поэтому идентификатор
// одновременно служит секретом доступа к попытке.
type ID string

// Status — состояние прохождения.
type Status string

// Возможные состояния попытки.
const (
	StatusInProgress Status = "in_progress"
	StatusCompleted  Status = "completed"
)

// IdempotencyKey — ключ повторной отправки выбора, заданный клиентом.
type IdempotencyKey string

// Границы результата попытки.
const (
	MinScore     = 0
	MaxScore     = 100
	InitialScore = MaxScore
)

// Decision — принятое решение вместе с рассчитанным сервером переходом.
//
// Снимок последствия и эффектов хранится вместе с решением: при повторной
// отправке того же ключа ответ должен быть точно таким же, даже если
// содержимое сценария позже изменится.
type Decision struct {
	NodeID         scenario.NodeID
	ChoiceID       scenario.ChoiceID
	ChoiceLabel    string
	IdempotencyKey IdempotencyKey
	Consequence    scenario.Consequence
	Criticality    scenario.Criticality
	RiskTags       []scenario.RiskTag
	SkillEffects   []scenario.SkillEffect
	ScoreDelta     int
	CreatedAt      time.Time

	// Поля ниже описывают результат перехода и заполняются при записи решения.
	RevealedNodeIDs []scenario.NodeID
	ResultingNodeID scenario.NodeID
	Completed       bool
	Outcome         scenario.OutcomeType
	ScoreAfter      int
}

func (d Decision) clone() Decision {
	cloned := d
	cloned.RiskTags = slices.Clone(d.RiskTags)
	cloned.SkillEffects = slices.Clone(d.SkillEffects)
	cloned.RevealedNodeIDs = slices.Clone(d.RevealedNodeIDs)

	return cloned
}

// Attempt — одно прохождение зафиксированной версии сценария.
type Attempt struct {
	ID              ID
	ScenarioID      scenario.ID
	ScenarioVersion scenario.Version
	CurrentNodeID   scenario.NodeID
	Status          Status
	Score           int
	Outcome         scenario.OutcomeType
	StartedAt       time.Time
	UpdatedAt       time.Time
	CompletedAt     *time.Time

	// RevealedNodeIDs хранит порядок раскрытия узлов: по нему восстанавливается
	// переписка после перезагрузки страницы.
	RevealedNodeIDs []scenario.NodeID
	Decisions       []Decision

	// AppliedSkillEffects накапливает изменения навыков для Backend №2.
	AppliedSkillEffects map[scenario.SkillCode]int

	// Version — оптимистичная блокировка. Хранилище принимает обновление
	// только если версия совпадает с сохранённой.
	Version int
}

// StartParams — данные для создания попытки.
type StartParams struct {
	ID              ID
	ScenarioID      scenario.ID
	ScenarioVersion scenario.Version
	StartNodeID     scenario.NodeID
	RevealedNodeIDs []scenario.NodeID
	CurrentNodeID   scenario.NodeID
	Completed       bool
	Outcome         scenario.OutcomeType
	StartedAt       time.Time
}

// Start создаёт попытку в состоянии, полученном после раскрытия первых узлов.
func Start(params StartParams) (Attempt, error) {
	if err := params.validate(); err != nil {
		return Attempt{}, err
	}

	started := Attempt{
		ID:                  params.ID,
		ScenarioID:          params.ScenarioID,
		ScenarioVersion:     params.ScenarioVersion,
		CurrentNodeID:       params.CurrentNodeID,
		Status:              StatusInProgress,
		Score:               InitialScore,
		StartedAt:           params.StartedAt,
		UpdatedAt:           params.StartedAt,
		RevealedNodeIDs:     slices.Clone(params.RevealedNodeIDs),
		Decisions:           []Decision{},
		AppliedSkillEffects: map[scenario.SkillCode]int{},
		Version:             1,
	}

	if params.Completed {
		started.complete(params.Outcome, params.StartedAt)
	}

	return started, nil
}

func (p StartParams) validate() error {
	switch {
	case p.ID == "":
		return ErrEmptyAttemptID
	case p.ScenarioID == "":
		return ErrEmptyScenarioID
	case p.ScenarioVersion < 1:
		return ErrInvalidScenarioVersion
	case p.StartNodeID == "" || p.CurrentNodeID == "":
		return ErrEmptyNode
	case p.StartedAt.IsZero():
		return ErrEmptyStartTime
	case len(p.RevealedNodeIDs) == 0:
		return ErrNothingRevealed
	}

	return nil
}

// EnsureInProgress проверяет, что попытку ещё можно изменять.
func (a Attempt) EnsureInProgress() error {
	if a.Status != StatusInProgress {
		return ErrAlreadyCompleted
	}

	return nil
}

// EnsureCurrentNode проверяет, что клиент отвечает на актуальный узел.
//
// Сравнение защищает от повторной отправки устаревшего экрана: текущий узел
// хранит сервер, а присланный клиентом служит только для сверки.
func (a Attempt) EnsureCurrentNode(nodeID scenario.NodeID) error {
	if a.CurrentNodeID != nodeID {
		return ErrStaleNode
	}

	return nil
}

// DecisionByIdempotencyKey ищет ранее принятое решение по ключу повтора.
func (a Attempt) DecisionByIdempotencyKey(key IdempotencyKey) (Decision, bool) {
	for _, decision := range a.Decisions {
		if decision.IdempotencyKey == key {
			return decision.clone(), true
		}
	}

	return Decision{}, false
}

// Record применяет решение к попытке: сохраняет его, пересчитывает результат
// и переводит попытку в новый узел.
//
// Метод возвращает записанное решение с заполненными полями перехода.
func (a *Attempt) Record(decision Decision, now time.Time) (Decision, error) {
	if err := a.EnsureInProgress(); err != nil {
		return Decision{}, err
	}

	if err := a.EnsureCurrentNode(decision.NodeID); err != nil {
		return Decision{}, err
	}

	if decision.IdempotencyKey == "" {
		return Decision{}, ErrEmptyIdempotencyKey
	}

	if decision.ResultingNodeID == "" {
		return Decision{}, ErrEmptyNode
	}

	recorded := decision.clone()
	recorded.CreatedAt = now

	a.Score = clampScore(a.Score + recorded.ScoreDelta)
	recorded.ScoreAfter = a.Score

	for _, effect := range recorded.SkillEffects {
		a.AppliedSkillEffects[effect.Skill] += effect.Delta
	}

	a.RevealedNodeIDs = append(a.RevealedNodeIDs, recorded.RevealedNodeIDs...)
	a.CurrentNodeID = recorded.ResultingNodeID
	a.UpdatedAt = now
	a.Decisions = append(a.Decisions, recorded)

	if recorded.Completed {
		a.complete(recorded.Outcome, now)
	}

	return recorded.clone(), nil
}

// complete переводит попытку в завершённое состояние.
// Время завершения проставляется один раз и дальше не меняется.
func (a *Attempt) complete(outcome scenario.OutcomeType, now time.Time) {
	if a.Status == StatusCompleted {
		return
	}

	completedAt := now

	a.Status = StatusCompleted
	a.Outcome = outcome
	a.CompletedAt = &completedAt
	a.UpdatedAt = now
}

// Clone возвращает независимую копию попытки.
// Нужен хранилищу: агрегат содержит слайсы и карту, которые иначе
// остались бы общими с вызывающим кодом.
func (a Attempt) Clone() Attempt {
	cloned := a
	cloned.RevealedNodeIDs = slices.Clone(a.RevealedNodeIDs)
	cloned.AppliedSkillEffects = maps.Clone(a.AppliedSkillEffects)

	if cloned.AppliedSkillEffects == nil {
		cloned.AppliedSkillEffects = map[scenario.SkillCode]int{}
	}

	cloned.Decisions = make([]Decision, len(a.Decisions))
	for i, decision := range a.Decisions {
		cloned.Decisions[i] = decision.clone()
	}

	if a.CompletedAt != nil {
		completedAt := *a.CompletedAt
		cloned.CompletedAt = &completedAt
	}

	return cloned
}

// clampScore удерживает результат в документированных границах.
func clampScore(score int) int {
	if score < MinScore {
		return MinScore
	}

	if score > MaxScore {
		return MaxScore
	}

	return score
}
