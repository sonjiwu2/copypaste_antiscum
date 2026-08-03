package httpapi

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

// Коды публичных ошибок API. Клиент опирается на код, а не на текст сообщения.
const (
	CodeNotFound      = "NOT_FOUND"
	CodeInternalError = "INTERNAL_ERROR"
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
