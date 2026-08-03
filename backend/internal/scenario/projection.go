package scenario

// Metadata — публичное описание сценария для каталога.
//
// Проекция намеренно не содержит узлов, вариантов выбора и последствий:
// граф сценария не должен покидать сервер до начала попытки.
type Metadata struct {
	ID               ID
	Version          Version
	Slug             string
	Role             Role
	Title            string
	Description      string
	Difficulty       Difficulty
	EstimatedMinutes int
}

// MetadataOf строит публичное описание сценария.
func MetadataOf(scenario Scenario) Metadata {
	return Metadata{
		ID:               scenario.ID,
		Version:          scenario.Version,
		Slug:             scenario.Slug,
		Role:             scenario.Role,
		Title:            scenario.Title,
		Description:      scenario.Description,
		Difficulty:       scenario.Difficulty,
		EstimatedMinutes: scenario.EstimatedMinutes,
	}
}
