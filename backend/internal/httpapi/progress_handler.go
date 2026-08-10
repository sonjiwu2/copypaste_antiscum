package httpapi

import (
	"net/http"

	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/progress"
)

// progressHandler обслуживает прогресс текущего профиля.
type progressHandler struct {
	progress *progress.Service
}

// get возвращает прогресс анонимного профиля запроса.
//
// Чужой профиль запросить нельзя: владелец берётся из cookie, а не из
// параметров запроса.
func (h *progressHandler) get(w http.ResponseWriter, r *http.Request) {
	found, err := h.progress.Of(r.Context(), ProfileIDFrom(r.Context()))
	if err != nil {
		writeDomainError(w, r, err)

		return
	}

	writeJSON(w, r, http.StatusOK, progressResponseOf(found))
}
