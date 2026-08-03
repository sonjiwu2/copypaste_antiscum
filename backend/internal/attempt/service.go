package attempt

import (
	"context"
	"fmt"

	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/platform/clock"
	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/platform/identifier"
	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/scenario"
)

// ScenarioCatalog — доступ к сценариям, нужный движку попыток.
// Интерфейс объявлен рядом с потребителем и намеренно уже, чем каталог целиком.
type ScenarioCatalog interface {
	Get(ctx context.Context, id scenario.ID) (scenario.Scenario, error)
}

// Service выполняет сценарии использования прохождения.
type Service struct {
	scenarios ScenarioCatalog
	attempts  Repository
	clock     clock.Clock
	ids       identifier.Generator
}

// NewService создаёт сервис прохождения.
func NewService(
	scenarios ScenarioCatalog,
	attempts Repository,
	clock clock.Clock,
	ids identifier.Generator,
) *Service {
	return &Service{scenarios: scenarios, attempts: attempts, clock: clock, ids: ids}
}

// Start создаёт попытку и раскрывает первые узлы сценария.
func (s *Service) Start(ctx context.Context, scenarioID scenario.ID) (View, error) {
	definition, err := s.scenarios.Get(ctx, scenarioID)
	if err != nil {
		return View{}, fmt.Errorf("получить сценарий: %w", err)
	}

	// Отключённый сценарий недоступен для прохождения и выглядит отсутствующим.
	if !definition.IsActive {
		return View{}, scenario.ErrNotFound
	}

	revealed, err := revealFrom(definition, definition.StartNodeID)
	if err != nil {
		return View{}, err
	}

	attemptID, err := s.ids.NewID()
	if err != nil {
		return View{}, fmt.Errorf("создать идентификатор попытки: %w", err)
	}

	started, err := Start(StartParams{
		ID:              ID(attemptID),
		ScenarioID:      definition.ID,
		ScenarioVersion: definition.Version,
		StartNodeID:     definition.StartNodeID,
		RevealedNodeIDs: revealed.NodeIDs,
		CurrentNodeID:   revealed.StopNodeID,
		Completed:       revealed.Completed,
		Outcome:         revealed.Outcome,
		StartedAt:       s.clock.Now(),
	})
	if err != nil {
		return View{}, err
	}

	if err := s.attempts.Create(ctx, started); err != nil {
		return View{}, fmt.Errorf("сохранить попытку: %w", err)
	}

	return viewOf(started, definition)
}

// Get возвращает текущее публичное состояние попытки.
// Тот же метод обслуживает восстановление после перезагрузки страницы.
func (s *Service) Get(ctx context.Context, id ID) (View, error) {
	found, err := s.attempts.Get(ctx, id)
	if err != nil {
		return View{}, fmt.Errorf("получить попытку: %w", err)
	}

	definition, err := s.scenarioOfAttempt(ctx, found)
	if err != nil {
		return View{}, err
	}

	return viewOf(found, definition)
}

// scenarioOfAttempt возвращает сценарий той версии, с которой попытка начата.
//
// Начатое прохождение обязано доигрываться на зафиксированной версии, поэтому
// расхождение версий — ошибка сервера, а не повод молча подставить новый текст.
func (s *Service) scenarioOfAttempt(ctx context.Context, source Attempt) (scenario.Scenario, error) {
	definition, err := s.scenarios.Get(ctx, source.ScenarioID)
	if err != nil {
		return scenario.Scenario{}, fmt.Errorf("получить сценарий попытки: %w", err)
	}

	if definition.Version != source.ScenarioVersion {
		return scenario.Scenario{}, fmt.Errorf("%w: попытка начата на версии %d, доступна версия %d",
			ErrScenarioVersionChanged, source.ScenarioVersion, definition.Version)
	}

	return definition, nil
}
