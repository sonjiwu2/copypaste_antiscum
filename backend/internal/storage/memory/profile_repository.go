package memory

import (
	"context"
	"sync"

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

// Count возвращает количество профилей. Нужен тестам изоляции.
func (r *ProfileRepository) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.profiles)
}
