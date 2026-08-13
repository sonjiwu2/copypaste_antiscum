package httpapi

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/profile"
)

type profileHandler struct {
	profiles      *profile.Service
	cookie        CookieSettings
	sessionCookie CookieSettings
}

type updateProfileRequest struct {
	DisplayName string `json:"displayName"`
	Avatar      string `json:"avatar"`
}

type profileIdentityResponse struct {
	DisplayName string `json:"displayName"`
	Avatar      string `json:"avatar"`
}

type leaderboardEntryResponse struct {
	Rank               int    `json:"rank"`
	DisplayName        string `json:"displayName"`
	Avatar             string `json:"avatar"`
	Rating             int    `json:"rating"`
	CompletedScenarios int    `json:"completedScenarios"`
	AverageScore       int    `json:"averageScore"`
	CurrentPlayer      bool   `json:"currentPlayer"`
}

type leaderboardResponse struct {
	Leaders []leaderboardEntryResponse `json:"leaders"`
	Current *leaderboardEntryResponse  `json:"current,omitempty"`
}

func (h *profileHandler) update(w http.ResponseWriter, r *http.Request) {
	var request updateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeDecodeError(w, r, err)
		return
	}
	identity, err := h.profiles.UpdateIdentity(r.Context(), ProfileIDFrom(r.Context()), profile.Identity{
		DisplayName: request.DisplayName,
		Avatar:      profile.Avatar(request.Avatar),
	})
	if err != nil {
		writeDomainError(w, r, err)
		return
	}
	writeJSON(w, r, http.StatusOK, profileIdentityResponse{
		DisplayName: identity.DisplayName, Avatar: string(identity.Avatar),
	})
}

func (h *profileHandler) leaderboard(w http.ResponseWriter, r *http.Request) {
	board, err := h.profiles.Leaderboard(r.Context(), ProfileIDFrom(r.Context()))
	if err != nil {
		writeDomainError(w, r, err)
		return
	}
	response := leaderboardResponse{Leaders: make([]leaderboardEntryResponse, 0, len(board.Leaders))}
	for _, leader := range board.Leaders {
		response.Leaders = append(response.Leaders, leaderboardEntryResponseOf(leader))
	}
	if board.Current != nil {
		current := leaderboardEntryResponseOf(*board.Current)
		response.Current = &current
	}
	writeJSON(w, r, http.StatusOK, response)
}

func leaderboardEntryResponseOf(leader profile.Leader) leaderboardEntryResponse {
	return leaderboardEntryResponse{
		Rank: leader.Rank, DisplayName: leader.DisplayName, Avatar: string(leader.Avatar),
		Rating: leader.Rating, CompletedScenarios: leader.CompletedScenarios,
		AverageScore: leader.AverageScore, CurrentPlayer: leader.CurrentPlayer,
	}
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
	clearCookie(w, h.sessionCookie)

	writeJSON(w, r, http.StatusOK, map[string]string{"status": "reset"})
}
