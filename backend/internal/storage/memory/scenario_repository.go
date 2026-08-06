// Package memory содержит адаптеры хранения в оперативной памяти.
//
// Они существуют, чтобы Backend №1 запускался и тестировался независимо,
// пока Backend №2 готовит PostgreSQL.
package memory

import (
	"context"
	"fmt"
	"sort"
	"sync"

	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/scenario"
)

// ScenarioRepository хранит проверенные сценарии в памяти.
//
// Каталог заполняется один раз при старте и дальше только читается, поэтому
// RWMutex защищает от гонок при параллельных HTTP-запросах и остаётся дешёвым.
type ScenarioRepository struct {
	mu       sync.RWMutex
	versions map[scenario.ID]map[scenario.Version]scenario.Scenario
	current  map[scenario.ID]scenario.Version
}

// NewScenarioRepository наполняет каталог загруженными сценариями.
func NewScenarioRepository(scenarios []scenario.Scenario) (*ScenarioRepository, error) {
	repository := &ScenarioRepository{
		versions: make(map[scenario.ID]map[scenario.Version]scenario.Scenario, len(scenarios)),
		current:  make(map[scenario.ID]scenario.Version, len(scenarios)),
	}

	for _, loaded := range scenarios {
		byVersion := repository.versions[loaded.ID]
		if byVersion == nil {
			byVersion = make(map[scenario.Version]scenario.Scenario)
			repository.versions[loaded.ID] = byVersion
		}

		if _, exists := byVersion[loaded.Version]; exists {
			return nil, fmt.Errorf("сценарий %q версии %d повторяется в каталоге", loaded.ID, loaded.Version)
		}

		if currentVersion, exists := repository.current[loaded.ID]; exists {
			current := byVersion[currentVersion]
			if loaded.IsActive && current.IsActive {
				return nil, fmt.Errorf("сценарий %q имеет несколько активных версий", loaded.ID)
			}

			if loaded.IsActive || (!current.IsActive && loaded.Version > currentVersion) {
				repository.current[loaded.ID] = loaded.Version
			}
		} else {
			repository.current[loaded.ID] = loaded.Version
		}

		byVersion[loaded.Version] = loaded
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

	found := make([]scenario.Scenario, 0, len(r.current))

	for id, version := range r.current {
		stored := r.versions[id][version]
		if filter.OnlyActive && !stored.IsActive {
			continue
		}

		if filter.Role != "" && stored.Role != filter.Role {
			continue
		}

		found = append(found, stored)
	}

	sort.Slice(found, func(i, j int) bool {
		if found[i].ID == found[j].ID {
			return found[i].Version < found[j].Version
		}

		return found[i].ID < found[j].ID
	})

	return found, nil
}

// Get возвращает сценарий по идентификатору.
func (r *ScenarioRepository) Get(ctx context.Context, id scenario.ID) (scenario.Scenario, error) {
	if err := ctx.Err(); err != nil {
		return scenario.Scenario{}, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	version, found := r.current[id]
	if !found {
		return scenario.Scenario{}, scenario.ErrNotFound
	}

	return r.versions[id][version], nil
}

// GetVersion возвращает точную версию, включая историческую неактивную.
func (r *ScenarioRepository) GetVersion(
	ctx context.Context,
	id scenario.ID,
	version scenario.Version,
) (scenario.Scenario, error) {
	if err := ctx.Err(); err != nil {
		return scenario.Scenario{}, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	byVersion := r.versions[id]
	stored, found := byVersion[version]
	if !found {
		return scenario.Scenario{}, scenario.ErrNotFound
	}

	return stored, nil
}
