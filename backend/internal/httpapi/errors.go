package httpapi

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/attempt"
	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/scenario"
)

// Коды публичных ошибок API. Клиент опирается на код, а не на текст сообщения.
const (
	CodeNotFound         = "NOT_FOUND"
	CodeInternalError    = "INTERNAL_ERROR"
	CodeInvalidRequest   = "INVALID_REQUEST"
	CodeScenarioNotFound = "SCENARIO_NOT_FOUND"
	CodeUnsupportedRole  = "UNSUPPORTED_ROLE"
	CodeAttemptNotFound  = "ATTEMPT_NOT_FOUND"
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

// writeDomainError переводит доменную ошибку в публичный ответ.
//
// Это единственное место перевода: тексты и коды ошибок не расползаются
// по handler'ам, а внутренние подробности не доходят до клиента.
func writeDomainError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, scenario.ErrNotFound):
		writeError(w, r, http.StatusNotFound, CodeScenarioNotFound, "Сценарий не найден.")
	case errors.Is(err, scenario.ErrUnsupportedRole):
		writeError(w, r, http.StatusBadRequest, CodeUnsupportedRole, "Указана неподдерживаемая роль.")
	case errors.Is(err, attempt.ErrNotFound):
		writeError(w, r, http.StatusNotFound, CodeAttemptNotFound, "Попытка не найдена.")
	default:
		// Неожиданная ошибка логируется один раз, на границе HTTP.
		loggerFrom(r.Context()).ErrorContext(r.Context(), "необработанная ошибка запроса",
			slog.String("error", err.Error()))

		writeError(w, r, http.StatusInternalServerError, CodeInternalError, "Внутренняя ошибка сервера.")
	}
}
