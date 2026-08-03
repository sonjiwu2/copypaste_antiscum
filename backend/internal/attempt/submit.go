package attempt

import (
	"context"
	"errors"
	"fmt"

	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/scenario"
)

// SubmitChoiceCommand — запрос на применение выбора.
//
// Клиент присылает только то, что не может быть вычислено сервером:
// какой узел он видел, что выбрал и каким ключом защищает повтор.
// Следующий узел, изменение результата и эффекты сюда не входят намеренно.
type SubmitChoiceCommand struct {
	AttemptID      ID
	NodeID         scenario.NodeID
	ChoiceID       scenario.ChoiceID
	IdempotencyKey IdempotencyKey
}

func (c SubmitChoiceCommand) validate() error {
	switch {
	case c.AttemptID == "":
		return ErrEmptyAttemptID
	case c.NodeID == "" || c.ChoiceID == "":
		return ErrChoiceNotFound
	case c.IdempotencyKey == "":
		return ErrEmptyIdempotencyKey
	}

	return nil
}

// SubmitChoice применяет выбор пользователя и продвигает прохождение.
func (s *Service) SubmitChoice(ctx context.Context, command SubmitChoiceCommand) (Transition, error) {
	if err := command.validate(); err != nil {
		return Transition{}, err
	}

	current, err := s.attempts.Get(ctx, command.AttemptID)
	if err != nil {
		return Transition{}, fmt.Errorf("получить попытку: %w", err)
	}

	definition, err := s.scenarioOfAttempt(ctx, current)
	if err != nil {
		return Transition{}, err
	}

	// Повтор проверяется раньше состояния попытки: сетевой таймаут после
	// последнего выбора не должен превращаться в «попытка уже завершена».
	if replay, replayed, err := replayOf(current, definition, command); replayed || err != nil {
		return replay, err
	}

	transition, err := s.applyChoice(ctx, current, definition, command)
	if !errors.Is(err, ErrConcurrentUpdate) {
		return transition, err
	}

	return s.resolveConflict(ctx, command)
}

// applyChoice проверяет выбор по данным сценария и сохраняет переход.
func (s *Service) applyChoice(
	ctx context.Context,
	current Attempt,
	definition scenario.Scenario,
	command SubmitChoiceCommand,
) (Transition, error) {
	if err := current.EnsureInProgress(); err != nil {
		return Transition{}, err
	}

	if err := current.EnsureCurrentNode(command.NodeID); err != nil {
		return Transition{}, err
	}

	// Узел и выбор берутся из серверного сценария, а не из запроса: иначе
	// подменой идентификатора можно было бы прыгнуть в скрытую ветку.
	node, found := definition.Node(current.CurrentNodeID)
	if !found {
		return Transition{}, fmt.Errorf("%w: узел %q сценария %q",
			ErrBrokenScenario, current.CurrentNodeID, definition.ID)
	}

	if node.Type != scenario.NodeTypeDecision {
		return Transition{}, ErrNodeNotDecision
	}

	choice, found := node.Choice(command.ChoiceID)
	if !found {
		return Transition{}, ErrChoiceNotFound
	}

	revealed, err := revealFrom(definition, choice.NextNodeID)
	if err != nil {
		return Transition{}, err
	}

	recorded, err := current.Record(decisionOf(command, choice, revealed), s.clock.Now())
	if err != nil {
		return Transition{}, err
	}

	saved, err := s.attempts.Update(ctx, current)
	if err != nil {
		return Transition{}, err
	}

	return transitionOf(saved, definition, recorded)
}

// decisionOf переносит эффекты выбора из сценария в запись решения.
// Все значения серверные: клиент не влияет ни на результат, ни на навыки.
func decisionOf(command SubmitChoiceCommand, choice scenario.Choice, revealed reveal) Decision {
	return Decision{
		NodeID:          command.NodeID,
		ChoiceID:        choice.ID,
		ChoiceLabel:     choice.Label,
		IdempotencyKey:  command.IdempotencyKey,
		Consequence:     choice.Consequence,
		Criticality:     choice.Criticality,
		RiskTags:        choice.RiskTags,
		SkillEffects:    choice.SkillEffects,
		ScoreDelta:      choice.SafetyScore,
		RevealedNodeIDs: revealed.NodeIDs,
		ResultingNodeID: revealed.StopNodeID,
		Completed:       revealed.Completed,
		Outcome:         revealed.Outcome,
	}
}

// replayOf возвращает ранее сохранённый результат, если ключ повтора уже
// использован. Тот же ключ с другими данными — конфликт, а не новый переход.
func replayOf(
	current Attempt,
	definition scenario.Scenario,
	command SubmitChoiceCommand,
) (Transition, bool, error) {
	previous, found := current.DecisionByIdempotencyKey(command.IdempotencyKey)
	if !found {
		return Transition{}, false, nil
	}

	if previous.NodeID != command.NodeID || previous.ChoiceID != command.ChoiceID {
		return Transition{}, true, ErrIdempotencyConflict
	}

	transition, err := transitionOf(current, definition, previous)

	return transition, true, err
}

// resolveConflict разбирает ситуацию, когда попытку успел изменить другой запрос.
//
// Одинаковый ключ означает повторную доставку того же запроса — тогда
// возвращается уже применённый результат. Иначе это действительно
// конкурентный переход, и клиент должен перечитать состояние попытки.
func (s *Service) resolveConflict(ctx context.Context, command SubmitChoiceCommand) (Transition, error) {
	current, err := s.attempts.Get(ctx, command.AttemptID)
	if err != nil {
		return Transition{}, fmt.Errorf("перечитать попытку: %w", err)
	}

	definition, err := s.scenarioOfAttempt(ctx, current)
	if err != nil {
		return Transition{}, err
	}

	if replay, replayed, err := replayOf(current, definition, command); replayed {
		return replay, err
	}

	return Transition{}, ErrConcurrentUpdate
}
