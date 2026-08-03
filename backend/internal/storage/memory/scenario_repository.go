// Package memory содержит адаптеры хранения в оперативной памяти.
//
// Они существуют, чтобы Backend №1 запускался и тестировался независимо,
// пока Backend №2 готовит PostgreSQL.
package memory

import (
	"context"
	"fmt"
	"sync"

	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/scenario"
)

// ScenarioRepository хранит проверенные сценарии в памяти.
//
// Каталог заполняется один раз при старте и дальше только читается, поэтому
// RWMutex защищает от гонок при параллельных HTTP-запросах и остаётся дешёвым.
type ScenarioRepository struct {
	mu        sync.RWMutex
	scenarios map[scenario.ID]scenario.Scenario
}

// NewScenarioRepository наполняет каталог загруженными сценариями.
func NewScenarioRepository(scenarios []scenario.Scenario) (*ScenarioRepository, error) {
	repository := &ScenarioRepository{
		scenarios: make(map[scenario.ID]scenario.Scenario, len(scenarios)),
	}

	for _, loaded := range scenarios {
		if _, exists := repository.scenarios[loaded.ID]; exists {
			return nil, fmt.Errorf("идентификатор сценария %q повторяется в каталоге", loaded.ID)
		}

		repository.scenarios[loaded.ID] = loaded
	}

	return repository, nil
}

// List возвращает сценарии, подходящие под фильтр.
func (r *ScenarioRepository) List(ctx context.Context, filter scenario.Filter) ([]scenario.Scenario, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	found := make([]scenario.Scenario, 0, len(r.scenarios))

	for _, stored := range r.scenarios {
		if filter.OnlyActive && !stored.IsActive {
			continue
		}

		if filter.Role != "" && stored.Role != filter.Role {
			continue
		}

		found = append(found, stored)
	}

	return found, nil
}

// Get возвращает сценарий по идентификатору.
func (r *ScenarioRepository) Get(ctx context.Context, id scenario.ID) (scenario.Scenario, error) {
	if err := ctx.Err(); err != nil {
		return scenario.Scenario{}, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	stored, found := r.scenarios[id]
	if !found {
		return scenario.Scenario{}, scenario.ErrNotFound
	}

	return stored, nil
}
