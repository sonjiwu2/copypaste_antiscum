package httpapi

import (
	"net/http"
	"time"

	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/profile"
)

type profileHandler struct {
	profiles *profile.Service
	cookie   CookieSettings
}

func (h *profileHandler) reset(w http.ResponseWriter, r *http.Request) {
	if err := h.profiles.Reset(r.Context(), ProfileIDFrom(r.Context())); err != nil {
		writeDomainError(w, r, err)

		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     h.cookie.Name,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		Expires:  time.Unix(1, 0).UTC(),
		HttpOnly: true,
		Secure:   h.cookie.Secure,
		SameSite: h.cookie.sameSite(),
	})

	writeJSON(w, r, http.StatusOK, map[string]string{"status": "reset"})
}
