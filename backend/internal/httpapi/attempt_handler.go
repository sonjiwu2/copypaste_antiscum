package httpapi

import (
	"encoding/json"
	"net/http"

	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/attempt"
	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/scenario"
)

// attemptHandler обслуживает прохождение сценария.
type attemptHandler struct {
	attempts *attempt.Service
}

// start создаёт попытку и возвращает первые раскрытые узлы.
func (h *attemptHandler) start(w http.ResponseWriter, r *http.Request) {
	var request startAttemptRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeError(w, r, http.StatusBadRequest, CodeInvalidRequest,
			"Тело запроса должно быть корректным JSON.")

		return
	}

	if request.ScenarioID == "" {
		writeError(w, r, http.StatusBadRequest, CodeInvalidRequest,
			"Поле scenarioId обязательно.")

		return
	}

	view, err := h.attempts.Start(r.Context(), scenario.ID(request.ScenarioID))
	if err != nil {
		writeDomainError(w, r, err)

		return
	}

	writeJSON(w, r, http.StatusCreated, attemptResponseOf(view))
}

// get возвращает текущее состояние попытки.
// Тем же ответом фронтенд восстанавливает экран после перезагрузки страницы.
func (h *attemptHandler) get(w http.ResponseWriter, r *http.Request) {
	attemptID := attempt.ID(r.PathValue("attemptId"))

	view, err := h.attempts.Get(r.Context(), attemptID)
	if err != nil {
		writeDomainError(w, r, err)

		return
	}

	writeJSON(w, r, http.StatusOK, attemptResponseOf(view))
}
