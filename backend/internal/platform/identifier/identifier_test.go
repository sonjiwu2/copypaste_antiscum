package identifier_test

import (
	"sync"
	"testing"

	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/platform/identifier"
)

func TestRandomNewIDIsUnique(t *testing.T) {
	const total = 1000

	generator := identifier.Random{}
	seen := make(map[string]struct{}, total)

	for range total {
		id, err := generator.NewID()
		if err != nil {
			t.Fatalf("неожиданная ошибка: %v", err)
		}

		if id == "" {
			t.Fatal("идентификатор не должен быть пустым")
		}

		if _, duplicate := seen[id]; duplicate {
			t.Fatalf("идентификатор %q повторился", id)
		}

		seen[id] = struct{}{}
	}
}

func TestSequentialNewIDIsRaceFree(t *testing.T) {
	const goroutines = 50

	generator := &identifier.Sequential{Prefix: "attempt"}

	var waitGroup sync.WaitGroup

	results := make(chan string, goroutines)

	for range goroutines {
		waitGroup.Add(1)

		go func() {
			defer waitGroup.Done()

			id, err := generator.NewID()
			if err != nil {
				t.Errorf("неожиданная ошибка: %v", err)

				return
			}

			results <- id
		}()
	}

	waitGroup.Wait()
	close(results)

	seen := make(map[string]struct{}, goroutines)
	for id := range results {
		if _, duplicate := seen[id]; duplicate {
			t.Fatalf("идентификатор %q повторился", id)
		}

		seen[id] = struct{}{}
	}

	if len(seen) != goroutines {
		t.Fatalf("получено %d идентификаторов, ожидалось %d", len(seen), goroutines)
	}
}
