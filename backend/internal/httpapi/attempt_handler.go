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
		writeDecodeError(w, r, err)

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

// submitChoice применяет выбор пользователя и возвращает результат шага.
func (h *attemptHandler) submitChoice(w http.ResponseWriter, r *http.Request) {
	var request submitChoiceRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeDecodeError(w, r, err)

		return
	}

	if missing := missingChoiceField(request); missing != "" {
		writeError(w, r, http.StatusBadRequest, CodeInvalidRequest,
			"Поле "+missing+" обязательно.")

		return
	}

	transition, err := h.attempts.SubmitChoice(r.Context(), attempt.SubmitChoiceCommand{
		AttemptID:      attempt.ID(r.PathValue("attemptId")),
		NodeID:         scenario.NodeID(request.NodeID),
		ChoiceID:       scenario.ChoiceID(request.ChoiceID),
		IdempotencyKey: attempt.IdempotencyKey(request.IdempotencyKey),
	})
	if err != nil {
		writeDomainError(w, r, err)

		return
	}

	writeJSON(w, r, http.StatusOK, transitionResponseOf(transition))
}

// missingChoiceField возвращает имя первого незаполненного обязательного поля.
func missingChoiceField(request submitChoiceRequest) string {
	switch {
	case request.NodeID == "":
		return "nodeId"
	case request.ChoiceID == "":
		return "choiceId"
	case request.IdempotencyKey == "":
		return "idempotencyKey"
	default:
		return ""
	}
}
