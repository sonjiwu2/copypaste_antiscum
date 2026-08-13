package memory

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/profile"
)

// ProfileRepository хранит анонимные профили в оперативной памяти.
type ProfileRepository struct {
	mu       sync.RWMutex
	profiles map[profile.ID]profile.Profile
}

// NewProfileRepository создаёт пустое хранилище профилей.
func NewProfileRepository() *ProfileRepository {
	return &ProfileRepository{profiles: make(map[profile.ID]profile.Profile)}
}

// Ensure создаёт профиль, если его ещё нет.
//
// Время создания у существующего профиля не переписывается: оно показывает,
// когда пользователь пришёл впервые.
func (r *ProfileRepository) Ensure(ctx context.Context, created profile.Profile) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	if created.ID == "" {
		return profile.ErrEmptyID
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	stored, exists := r.profiles[created.ID]
	if exists {
		stored.UpdatedAt = created.UpdatedAt
		r.profiles[created.ID] = stored

		return nil
	}

	r.profiles[created.ID] = created

	return nil
}

func (r *ProfileRepository) Reset(ctx context.Context, id profile.ID) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if id == "" {
		return profile.ErrEmptyID
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.profiles, id)

	return nil
}

func (r *ProfileRepository) UpdateIdentity(
	ctx context.Context,
	id profile.ID,
	identity profile.Identity,
	updatedAt time.Time,
) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	stored, found := r.profiles[id]
	if !found {
		return profile.ErrInvalidID
	}
	stored.Identity = identity
	stored.UpdatedAt = updatedAt
	r.profiles[id] = stored
	return nil
}

// В memory-режиме нет общего durable-прогресса. Метод оставлен полноценным
// по контракту и возвращает профили без рейтинга только для тестов/демо.
func (r *ProfileRepository) Leaderboard(
	ctx context.Context,
	current profile.ID,
	limit int,
) (profile.Leaderboard, error) {
	if err := ctx.Err(); err != nil {
		return profile.Leaderboard{}, err
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	leaders := make([]profile.Leader, 0, len(r.profiles))
	for id, stored := range r.profiles {
		if stored.Identity.DisplayName == "" {
			continue
		}
		leaders = append(leaders, profile.Leader{
			DisplayName:   stored.Identity.DisplayName,
			Avatar:        stored.Identity.Avatar,
			CurrentPlayer: id == current,
		})
	}
	sort.Slice(leaders, func(i, j int) bool { return leaders[i].DisplayName < leaders[j].DisplayName })
	if limit < len(leaders) {
		leaders = leaders[:limit]
	}
	for index := range leaders {
		leaders[index].Rank = index + 1
	}
	return profile.Leaderboard{Leaders: leaders}, nil
}

// Count возвращает количество профилей. Нужен тестам изоляции.
func (r *ProfileRepository) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.profiles)
}
