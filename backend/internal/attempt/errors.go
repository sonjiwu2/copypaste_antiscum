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
