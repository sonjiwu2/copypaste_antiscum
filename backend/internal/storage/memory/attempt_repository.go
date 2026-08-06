package memory

import (
	"context"
	"sync"

	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/attempt"
	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/profile"
)

// AttemptRepository хранит попытки в оперативной памяти.
//
// Мьютекс защищает карту, а сверка версии под тем же замком делает запись
// перехода атомарной: два параллельных выбора из одного узла не могут
// примениться оба.
type AttemptRepository struct {
	mu       sync.RWMutex
	attempts map[attempt.ID]attempt.Attempt
}

// NewAttemptRepository создаёт пустое хранилище попыток.
func NewAttemptRepository() *AttemptRepository {
	return &AttemptRepository{attempts: make(map[attempt.ID]attempt.Attempt)}
}

// Create сохраняет новую попытку.
func (r *AttemptRepository) Create(ctx context.Context, created attempt.Attempt) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.attempts[created.ID]; exists {
		return attempt.ErrDuplicateAttempt
	}

	r.attempts[created.ID] = created.Clone()

	return nil
}

// Get возвращает копию попытки.
func (r *AttemptRepository) Get(ctx context.Context, id attempt.ID) (attempt.Attempt, error) {
	if err := ctx.Err(); err != nil {
		return attempt.Attempt{}, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	stored, found := r.attempts[id]
	if !found {
		return attempt.Attempt{}, attempt.ErrNotFound
	}

	return stored.Clone(), nil
}

// Update сохраняет попытку, если её версия не устарела, и повышает версию.
func (r *AttemptRepository) Update(ctx context.Context, updated attempt.Attempt) (attempt.Attempt, error) {
	if err := ctx.Err(); err != nil {
		return attempt.Attempt{}, err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	stored, found := r.attempts[updated.ID]
	if !found {
		return attempt.Attempt{}, attempt.ErrNotFound
	}

	if stored.Version != updated.Version {
		return attempt.Attempt{}, attempt.ErrConcurrentUpdate
	}

	next := updated.Clone()
	next.Version = stored.Version + 1
	r.attempts[updated.ID] = next

	return next.Clone(), nil
}

// ownedBy возвращает копии попыток профиля.
//
// Метод нужен хранилищу прогресса в том же пакете: наружу изменяемое
// состояние по-прежнему не выходит.
func (r *AttemptRepository) ownedBy(owner profile.ID) []attempt.Attempt {
	r.mu.RLock()
	defer r.mu.RUnlock()

	owned := make([]attempt.Attempt, 0, len(r.attempts))

	for _, stored := range r.attempts {
		if stored.ProfileID != owner {
			continue
		}

		owned = append(owned, stored.Clone())
	}

	return owned
}
