package httpapi

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/attempt"
	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/auth"
	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/profile"
	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/scenario"
	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/weeklytest"
)

// Коды публичных ошибок API. Клиент опирается на код, а не на текст сообщения.
const (
	CodeNotFound         = "NOT_FOUND"
	CodeInternalError    = "INTERNAL_ERROR"
	CodeInvalidRequest   = "INVALID_REQUEST"
	CodeScenarioNotFound = "SCENARIO_NOT_FOUND"
	CodeUnsupportedRole  = "UNSUPPORTED_ROLE"
	CodeAttemptNotFound  = "ATTEMPT_NOT_FOUND"
	CodeAttemptForbidden = "ATTEMPT_FORBIDDEN"

	CodePayloadTooLarge    = "PAYLOAD_TOO_LARGE"
	CodeInvalidProfile     = "INVALID_PROFILE"
	CodeAuthRequired       = "AUTH_REQUIRED"
	CodeInvalidAuthData    = "INVALID_AUTH_DATA"
	CodeInvalidCredentials = "INVALID_CREDENTIALS"
	CodeEmailTaken         = "EMAIL_TAKEN"

	CodeAttemptAlreadyCompleted  = "ATTEMPT_ALREADY_COMPLETED"
	CodeStaleNode                = "STALE_NODE"
	CodeIdempotencyKeyConflict   = "IDEMPOTENCY_KEY_CONFLICT"
	CodeConcurrentTransition     = "CONCURRENT_TRANSITION"
	CodeNodeNotDecision          = "NODE_NOT_DECISION"
	CodeChoiceNotFound           = "CHOICE_NOT_FOUND"
	CodeWeeklyTestNotFound       = "WEEKLY_TEST_NOT_FOUND"
	CodeWeeklyTestForbidden      = "WEEKLY_TEST_FORBIDDEN"
	CodeWeeklyTestInvalidAnswers = "WEEKLY_TEST_INVALID_ANSWERS"
	CodeWeeklyTestCompleted      = "WEEKLY_TEST_ALREADY_COMPLETED"
)

// errorEnvelope — единый формат ошибки для всех endpoint.
type errorEnvelope struct {
	Error errorBody `json:"error"`
}

type errorBody struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"requestId,omitempty"`
}

// writeJSON сериализует тело ответа. Ошибка записи означает разорванное
// соединение, поэтому статус менять уже поздно — только пишем в лог.
func writeJSON(w http.ResponseWriter, r *http.Request, status int, body any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)

	if body == nil {
		return
	}

	if err := json.NewEncoder(w).Encode(body); err != nil {
		loggerFrom(r.Context()).ErrorContext(r.Context(), "не удалось записать тело ответа",
			slog.String("error", err.Error()))
	}
}

// writeError отправляет клиенту ошибку в едином формате.
func writeError(w http.ResponseWriter, r *http.Request, status int, code, message string) {
	writeJSON(w, r, status, errorEnvelope{Error: errorBody{
		Code:      code,
		Message:   message,
		RequestID: RequestIDFrom(r.Context()),
	}})
}

// writeDecodeError объясняет клиенту, почему тело запроса не прочитано.
//
// Превышение лимита размера отделено от синтаксической ошибки: иначе клиент,
// приславший слишком большое тело, получал бы сообщение о некорректном JSON.
func writeDecodeError(w http.ResponseWriter, r *http.Request, err error) {
	var tooLarge *http.MaxBytesError

	if errors.As(err, &tooLarge) {
		writeError(w, r, http.StatusRequestEntityTooLarge, CodePayloadTooLarge,
			"Тело запроса превышает допустимый размер.")

		return
	}

	writeError(w, r, http.StatusBadRequest, CodeInvalidRequest,
		"Тело запроса должно быть корректным JSON.")
}

// writeDomainError переводит доменную ошибку в публичный ответ.
//
// Это единственное место перевода: тексты и коды ошибок не расползаются
// по handler'ам, а внутренние подробности не доходят до клиента.
func writeDomainError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, auth.ErrRequired):
		writeError(w, r, http.StatusUnauthorized, CodeAuthRequired, "Войдите в аккаунт.")
	case errors.Is(err, auth.ErrInvalidCredentials):
		writeError(w, r, http.StatusUnauthorized, CodeInvalidCredentials, "Неверная электронная почта или пароль.")
	case errors.Is(err, auth.ErrEmailTaken):
		writeError(w, r, http.StatusConflict, CodeEmailTaken, "Эта почта уже используется.")
	case errors.Is(err, auth.ErrProfileClaimed):
		writeError(w, r, http.StatusConflict, CodeInvalidAuthData,
			"На этом устройстве профиль уже зарегистрирован.")
	case errors.Is(err, auth.ErrInvalidEmail), errors.Is(err, auth.ErrWeakPassword):
		writeError(w, r, http.StatusUnprocessableEntity, CodeInvalidAuthData,
			"Проверьте электронную почту, имя и пароль.")
	case errors.Is(err, profile.ErrInvalidName), errors.Is(err, profile.ErrInvalidAvatar):
		writeError(w, r, http.StatusUnprocessableEntity, CodeInvalidProfile,
			"Имя или аватар игрока указаны некорректно.")
	case errors.Is(err, profile.ErrEmptyID):
		// Профиль присваивает middleware, поэтому его отсутствие — дефект
		// сборки приложения, а не ошибка клиента.
		loggerFrom(r.Context()).ErrorContext(r.Context(), "запрос дошёл до обработчика без профиля")

		writeError(w, r, http.StatusInternalServerError, CodeInternalError, "Внутренняя ошибка сервера.")
	case errors.Is(err, scenario.ErrNotFound):
		writeError(w, r, http.StatusNotFound, CodeScenarioNotFound, "Сценарий не найден.")
	case errors.Is(err, scenario.ErrUnsupportedRole):
		writeError(w, r, http.StatusBadRequest, CodeUnsupportedRole, "Указана неподдерживаемая роль.")
	case errors.Is(err, attempt.ErrNotFound):
		writeError(w, r, http.StatusNotFound, CodeAttemptNotFound, "Попытка не найдена.")
	case errors.Is(err, attempt.ErrForbidden):
		writeError(w, r, http.StatusForbidden, CodeAttemptForbidden,
			"Попытка принадлежит другому пользователю.")
	case errors.Is(err, attempt.ErrAlreadyCompleted):
		writeError(w, r, http.StatusConflict, CodeAttemptAlreadyCompleted,
			"Попытка уже завершена.")
	case errors.Is(err, attempt.ErrStaleNode):
		writeError(w, r, http.StatusConflict, CodeStaleNode,
			"Присланный узел не совпадает с текущим. Перечитайте состояние попытки.")
	case errors.Is(err, attempt.ErrIdempotencyConflict):
		writeError(w, r, http.StatusConflict, CodeIdempotencyKeyConflict,
			"Ключ повтора уже использован с другими данными.")
	case errors.Is(err, attempt.ErrConcurrentUpdate):
		writeError(w, r, http.StatusConflict, CodeConcurrentTransition,
			"Попытку изменил другой запрос. Перечитайте состояние попытки.")
	case errors.Is(err, attempt.ErrNodeNotDecision):
		writeError(w, r, http.StatusUnprocessableEntity, CodeNodeNotDecision,
			"На текущем узле выбор не принимается.")
	case errors.Is(err, attempt.ErrChoiceNotFound):
		writeError(w, r, http.StatusUnprocessableEntity, CodeChoiceNotFound,
			"Такой вариант выбора недоступен на текущем узле.")
	case errors.Is(err, weeklytest.ErrNotFound):
		writeError(w, r, http.StatusNotFound, CodeWeeklyTestNotFound,
			"Еженедельный тест не найден.")
	case errors.Is(err, weeklytest.ErrForbidden):
		writeError(w, r, http.StatusForbidden, CodeWeeklyTestForbidden,
			"Еженедельный тест принадлежит другому пользователю.")
	case errors.Is(err, weeklytest.ErrInvalidAnswers):
		writeError(w, r, http.StatusUnprocessableEntity, CodeWeeklyTestInvalidAnswers,
			"Переданы некорректные ответы на тест.")
	case errors.Is(err, weeklytest.ErrAlreadyCompleted):
		writeError(w, r, http.StatusConflict, CodeWeeklyTestCompleted,
			"Еженедельный тест уже завершён.")
	default:
		// Неожиданная ошибка логируется один раз, на границе HTTP.
		loggerFrom(r.Context()).ErrorContext(r.Context(), "необработанная ошибка запроса",
			slog.String("error", err.Error()))

		writeError(w, r, http.StatusInternalServerError, CodeInternalError, "Внутренняя ошибка сервера.")
	}
}
