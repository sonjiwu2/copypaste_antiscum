package attempt

import "errors"

// Доменные ошибки попытки. Перевод в HTTP-статусы выполняется в слое httpapi.
var (
	// ErrNotFound — попытка не найдена.
	ErrNotFound = errors.New("попытка не найдена")

	// ErrAlreadyCompleted — завершённую попытку изменять нельзя.
	ErrAlreadyCompleted = errors.New("попытка уже завершена")

	// ErrStaleNode — присланный узел не совпадает с текущим узлом попытки.
	ErrStaleNode = errors.New("узел попытки устарел")

	// ErrConcurrentUpdate — попытку успел изменить другой запрос.
	ErrConcurrentUpdate = errors.New("попытка изменена другим запросом")

	// ErrIdempotencyConflict — ключ повтора уже использован с другими данными.
	ErrIdempotencyConflict = errors.New("ключ повтора использован с другими данными")

	// ErrDuplicateAttempt — попытка с таким идентификатором уже существует.
	ErrDuplicateAttempt = errors.New("попытка с таким идентификатором уже существует")

	// ErrBrokenScenario — сценарий попытки повреждён. После проверки графа при
	// загрузке это означает ошибку сервера, а не некорректный запрос клиента.
	ErrBrokenScenario = errors.New("сценарий попытки повреждён")

	// ErrScenarioVersionChanged — сценарий изменился после начала попытки.
	ErrScenarioVersionChanged = errors.New("версия сценария попытки недоступна")

	// ErrNodeNotDecision — на текущем узле выбор не принимается.
	ErrNodeNotDecision = errors.New("текущий узел не является узлом решения")

	// ErrChoiceNotFound — выбор не принадлежит текущему узлу попытки.
	ErrChoiceNotFound = errors.New("вариант выбора недоступен на текущем узле")
)

// Ошибки некорректных исходных данных попытки.
var (
	ErrEmptyAttemptID         = errors.New("идентификатор попытки обязателен")
	ErrEmptyScenarioID        = errors.New("идентификатор сценария обязателен")
	ErrInvalidScenarioVersion = errors.New("версия сценария должна быть положительной")
	ErrEmptyNode              = errors.New("узел попытки обязателен")
	ErrEmptyStartTime         = errors.New("время начала попытки обязательно")
	ErrNothingRevealed        = errors.New("попытка должна начинаться хотя бы с одного раскрытого узла")
	ErrEmptyIdempotencyKey    = errors.New("ключ повтора обязателен")
)
