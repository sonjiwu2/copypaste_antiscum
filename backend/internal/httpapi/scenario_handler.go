package httpapi

import (
	"net/http"

	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/scenario"
)

// scenarioHandler обслуживает публичный каталог сценариев.
type scenarioHandler struct {
	scenarios *scenario.Service
}

// list отдаёт список доступных сценариев с необязательным фильтром по роли.
func (h *scenarioHandler) list(w http.ResponseWriter, r *http.Request) {
	role := scenario.Role(r.URL.Query().Get("role"))

	catalog, err := h.scenarios.List(r.Context(), role)
	if err != nil {
		writeDomainError(w, r, err)

		return
	}

	writeJSON(w, r, http.StatusOK, scenarioListResponse{Scenarios: scenarioSummariesOf(catalog)})
}

// get отдаёт описание одного сценария.
func (h *scenarioHandler) get(w http.ResponseWriter, r *http.Request) {
	scenarioID := scenario.ID(r.PathValue("scenarioId"))

	metadata, err := h.scenarios.Get(r.Context(), scenarioID)
	if err != nil {
		writeDomainError(w, r, err)

		return
	}

	writeJSON(w, r, http.StatusOK, scenarioSummaryOf(metadata))
}
