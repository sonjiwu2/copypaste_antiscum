package profile

import (
	"context"
	"fmt"
	"time"

	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/platform/clock"
	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/platform/identifier"
)

// Repository хранит профили.
//
// Ensure идемпотентен: повторный вызов с тем же идентификатором не создаёт
// второй записи и не считается ошибкой.
type Repository interface {
	Ensure(ctx context.Context, profile Profile) error
	Reset(ctx context.Context, id ID) error
	UpdateIdentity(ctx context.Context, id ID, identity Identity, updatedAt time.Time) error
	Leaderboard(ctx context.Context, current ID, limit int) (Leaderboard, error)
}

const LeaderboardLimit = 5

func (s *Service) UpdateIdentity(ctx context.Context, id ID, identity Identity) (Identity, error) {
	if !id.Valid() {
		return Identity{}, ErrInvalidID
	}
	normalized, err := NormalizeIdentity(identity)
	if err != nil {
		return Identity{}, err
	}
	if err := s.repository.UpdateIdentity(ctx, id, normalized, s.clock.Now()); err != nil {
		return Identity{}, fmt.Errorf("обновить публичный профиль: %w", err)
	}
	return normalized, nil
}

func (s *Service) Leaderboard(ctx context.Context, current ID) (Leaderboard, error) {
	if !current.Valid() {
		return Leaderboard{}, ErrInvalidID
	}
	board, err := s.repository.Leaderboard(ctx, current, LeaderboardLimit)
	if err != nil {
		return Leaderboard{}, fmt.Errorf("прочитать таблицу лидеров: %w", err)
	}
	return board, nil
}

// Reset удаляет серверные данные анонимного профиля. После этого HTTP-слой
// стирает cookie, и следующий запрос получает совершенно новый профиль.
func (s *Service) Reset(ctx context.Context, id ID) error {
	if !id.Valid() {
		return ErrInvalidID
	}

	if err := s.repository.Reset(ctx, id); err != nil {
		return fmt.Errorf("удалить данные профиля: %w", err)
	}

	return nil
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
