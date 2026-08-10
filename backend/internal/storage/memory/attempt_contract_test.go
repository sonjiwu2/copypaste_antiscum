package memory_test

import (
	"testing"

	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/attempt"
	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/storage/attempttest"
	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/storage/memory"
)

// Тот же контракт прогоняется против PostgreSQL-адаптера. Расхождение
// поведения хранилищ проявится здесь, а не в продакшене после переключения.
func TestAttemptRepositoryContract(t *testing.T) {
	attempttest.RunRepositoryContract(t, func(_ *testing.T) attempt.Repository {
		return memory.NewAttemptRepository()
	})
}
