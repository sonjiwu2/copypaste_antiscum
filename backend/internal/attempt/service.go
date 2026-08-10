package attempt

import (
	"context"
	"errors"
	"fmt"

	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/platform/clock"
	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/platform/identifier"
	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/profile"
	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/scenario"
)

// ScenarioCatalog — доступ к сценариям, нужный движку попыток.
// Интерфейс объявлен рядом с потребителем и намеренно уже, чем каталог целиком.
type ScenarioCatalog interface {
	Get(ctx context.Context, id scenario.ID) (scenario.Scenario, error)
}

// ScenarioVersions — точный архив версий для уже начатых попыток.
type ScenarioVersions interface {
	GetVersion(ctx context.Context, id scenario.ID, version scenario.Version) (scenario.Scenario, error)
}

// Service выполняет сценарии использования прохождения.
type Service struct {
	scenarios ScenarioCatalog
	versions  ScenarioVersions
	attempts  Repository
	clock     clock.Clock
	ids       identifier.Generator
}

// NewService создаёт сервис прохождения.
func NewService(
	scenarios ScenarioCatalog,
	versions ScenarioVersions,
	attempts Repository,
	clock clock.Clock,
	ids identifier.Generator,
) *Service {
	return &Service{
		scenarios: scenarios,
		versions:  versions,
		attempts:  attempts,
		clock:     clock,
		ids:       ids,
	}
}

// Start создаёт попытку и раскрывает первые узлы сценария.
//
// Владелец приходит параметром, а не из тела запроса: клиент не должен иметь
// возможности начать прохождение от чужого имени.
func (s *Service) Start(ctx context.Context, owner profile.ID, scenarioID scenario.ID) (View, error) {
	if owner == "" {
		return View{}, ErrEmptyProfileID
	}

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
		ProfileID:       owner,
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
func (s *Service) Get(ctx context.Context, owner profile.ID, id ID) (View, error) {
	found, err := s.attempts.Get(ctx, id)
	if err != nil {
		return View{}, fmt.Errorf("получить попытку: %w", err)
	}

	if err := found.EnsureOwnedBy(owner); err != nil {
		return View{}, err
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
	definition, err := s.versions.GetVersion(ctx, source.ScenarioID, source.ScenarioVersion)
	if err != nil {
		if errors.Is(err, scenario.ErrNotFound) || errors.Is(err, scenario.ErrVersionCorrupt) {
			return scenario.Scenario{}, fmt.Errorf("%w: сценарий %q версии %d: %w",
				ErrScenarioVersionChanged, source.ScenarioID, source.ScenarioVersion, err)
		}

		return scenario.Scenario{}, fmt.Errorf("получить версию сценария попытки: %w", err)
	}

	return definition, nil
}
