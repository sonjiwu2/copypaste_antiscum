package scenario

import (
	"context"
	"fmt"
	"sort"
)

// Service реализует сценарии использования каталога.
type Service struct {
	repository Repository
}

// NewService создаёт сервис каталога сценариев.
func NewService(repository Repository) *Service {
	return &Service{repository: repository}
}

// List возвращает публичные описания активных сценариев.
// Пустая роль означает выборку без фильтра по роли.
func (s *Service) List(ctx context.Context, role Role) ([]Metadata, error) {
	if role != "" && !role.Valid() {
		return nil, fmt.Errorf("%w: %q", ErrUnsupportedRole, role)
	}

	scenarios, err := s.repository.List(ctx, Filter{Role: role, OnlyActive: true})
	if err != nil {
		return nil, fmt.Errorf("получить список сценариев: %w", err)
	}

	catalog := make([]Metadata, 0, len(scenarios))
	for _, found := range scenarios {
		catalog = append(catalog, MetadataOf(found))
	}

	// Порядок каталога не должен зависеть от порядка обхода хранилища,
	// иначе клиент видит разный список при одинаковом запросе.
	sort.Slice(catalog, func(i, j int) bool {
		return catalog[i].ID < catalog[j].ID
	})

	return catalog, nil
}

// Get возвращает публичное описание одного сценария.
// Отключённый сценарий считается отсутствующим.
func (s *Service) Get(ctx context.Context, id ID) (Metadata, error) {
	found, err := s.repository.Get(ctx, id)
	if err != nil {
		return Metadata{}, fmt.Errorf("получить сценарий: %w", err)
	}

	if !found.IsActive {
		return Metadata{}, ErrNotFound
	}

	return MetadataOf(found), nil
}
