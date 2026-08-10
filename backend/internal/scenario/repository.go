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
// Интерфейс объявлен рядом с потребителем. Текущий публичный каталог собирается
// из проверенных embedded-файлов, а точные исторические версии обслуживает
// отдельный VersionRepository.
type Repository interface {
	List(ctx context.Context, filter Filter) ([]Scenario, error)
	Get(ctx context.Context, id ID) (Scenario, error)
}

// VersionRepository возвращает точную неизменяемую версию сценария.
//
// Каталог и архив разделены намеренно: публичный список знает только текущие
// версии, а уже начатая попытка должна продолжаться по своей старой версии.
type VersionRepository interface {
	GetVersion(ctx context.Context, id ID, version Version) (Scenario, error)
}
