package scenario

import "context"

// Filter ограничивает выборку сценариев.
type Filter struct {
	// Role пустая — сценарии всех ролей.
	Role Role
	// OnlyActive исключает сценарии, отключённые из каталога.
	OnlyActive bool
}

// Repository хранит сценарии и отдаёт их по запросу.
//
// Интерфейс объявлен рядом с потребителем: реализация на PostgreSQL появится
// в зоне Backend №2 и заменит in-memory адаптер без изменений в сервисе.
type Repository interface {
	List(ctx context.Context, filter Filter) ([]Scenario, error)
	Get(ctx context.Context, id ID) (Scenario, error)
}
