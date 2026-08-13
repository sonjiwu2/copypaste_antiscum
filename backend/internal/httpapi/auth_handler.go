package httpapi

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/auth"
	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/profile"
)

type authHandler struct {
	auth   *auth.Service
	cookie CookieSettings
}

type registerRequest struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	DisplayName string `json:"displayName"`
	Avatar      string `json:"avatar"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type sessionResponse struct {
	Email        string `json:"email"`
	DisplayName  string `json:"displayName"`
	Avatar       string `json:"avatar"`
	RegisteredAt string `json:"registeredAt"`
	ExpiresAt    string `json:"expiresAt"`
}

func (h *authHandler) register(w http.ResponseWriter, r *http.Request) {
	var request registerRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeDecodeError(w, r, err)
		return
	}
	current, token, err := h.auth.Register(r.Context(), ProfileIDFrom(r.Context()), auth.RegisterInput{
		Email: request.Email, Password: request.Password,
		Identity: profile.Identity{DisplayName: request.DisplayName, Avatar: profile.Avatar(request.Avatar)},
	})
	if err != nil {
		writeDomainError(w, r, err)
		return
	}
	setSessionCookie(w, h.cookie, token)
	writeJSON(w, r, http.StatusCreated, sessionResponseOf(current))
}

func (h *authHandler) login(w http.ResponseWriter, r *http.Request) {
	var request loginRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeDecodeError(w, r, err)
		return
	}
	current, token, err := h.auth.Login(r.Context(), request.Email, request.Password)
	if err != nil {
		writeDomainError(w, r, err)
		return
	}
	setSessionCookie(w, h.cookie, token)
	writeJSON(w, r, http.StatusOK, sessionResponseOf(current))
}

func (h *authHandler) current(w http.ResponseWriter, r *http.Request) {
	current, ok := AccountFrom(r.Context())
	if !ok {
		writeError(w, r, http.StatusUnauthorized, CodeAuthRequired, "Войдите в аккаунт.")
		return
	}
	writeJSON(w, r, http.StatusOK, sessionResponseOf(current))
}

func (h *authHandler) logout(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie(h.cookie.Name); err == nil {
		if err := h.auth.Logout(r.Context(), cookie.Value); err != nil {
			writeDomainError(w, r, err)
			return
		}
	}
	clearCookie(w, h.cookie)
	writeJSON(w, r, http.StatusOK, map[string]string{"status": "signed_out"})
}

func sessionResponseOf(current auth.Current) sessionResponse {
	return sessionResponse{Email: current.Email, DisplayName: current.Identity.DisplayName,
		Avatar: string(current.Identity.Avatar), RegisteredAt: current.CreatedAt.UTC().Format(time.RFC3339),
		ExpiresAt: current.ExpiresAt.UTC().Format(time.RFC3339)}
}
