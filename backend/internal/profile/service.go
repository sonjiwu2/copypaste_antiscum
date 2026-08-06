package profile

import (
	"context"
	"fmt"

	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/platform/clock"
	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/platform/identifier"
)

// Repository хранит профили.
//
// Ensure идемпотентен: повторный вызов с тем же идентификатором не создаёт
// второй записи и не считается ошибкой.
type Repository interface {
	Ensure(ctx context.Context, profile Profile) error
}

// Service выдаёт и подтверждает анонимные профили.
type Service struct {
	repository Repository
	clock      clock.Clock
	ids        identifier.Generator
}

// NewService создаёт сервис профилей.
func NewService(repository Repository, clock clock.Clock, ids identifier.Generator) *Service {
	return &Service{repository: repository, clock: clock, ids: ids}
}

// Resolved — результат разбора запроса.
type Resolved struct {
	ID ID

	// Issued сообщает, что профиль выдан только что и клиенту нужно
	// сохранить идентификатор.
	Issued bool
}

// Resolve возвращает профиль запроса, создавая его при необходимости.
//
// Отсутствующий или синтаксически невалидный идентификатор заменяется новым.
// Валидная cookie — bearer-секрет доступа к анонимной истории: репозиторий
// идемпотентно подтверждает профиль или создаёт его при первом обращении.
func (s *Service) Resolve(ctx context.Context, presented ID) (Resolved, error) {
	issued := false

	if !presented.Valid() {
		generated, err := s.ids.NewID()
		if err != nil {
			return Resolved{}, fmt.Errorf("создать идентификатор профиля: %w", err)
		}

		presented = ID(generated)
		issued = true
	}

	now := s.clock.Now()

	// Профиль создаётся лениво: до первого запроса записи в базе нет.
	if err := s.repository.Ensure(ctx, Profile{
		ID:        presented,
		CreatedAt: now,
		UpdatedAt: now,
	}); err != nil {
		return Resolved{}, fmt.Errorf("сохранить профиль: %w", err)
	}

	return Resolved{ID: presented, Issued: issued}, nil
}
