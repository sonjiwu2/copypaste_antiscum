package httpapi

import (
	"encoding/json"
	"net/http"

	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/weeklytest"
)

type weeklyTestHandler struct {
	tests *weeklytest.Service
}

func (h *weeklyTestHandler) current(w http.ResponseWriter, r *http.Request) {
	found, err := h.tests.Current(r.Context(), ProfileIDFrom(r.Context()))
	if err != nil {
		writeDomainError(w, r, err)

		return
	}

	writeJSON(w, r, http.StatusOK, weeklyTestResponseOf(found))
}

func (h *weeklyTestHandler) submit(w http.ResponseWriter, r *http.Request) {
	var request submitWeeklyTestRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeDecodeError(w, r, err)

		return
	}

	answers := make([]weeklytest.Answer, 0, len(request.Answers))
	for _, answer := range request.Answers {
		answers = append(answers, weeklytest.Answer{
			QuestionID:  answer.QuestionID,
			OptionIndex: answer.OptionIndex,
		})
	}

	found, err := h.tests.Submit(
		r.Context(),
		ProfileIDFrom(r.Context()),
		weeklytest.ID(r.PathValue("testId")),
		answers,
	)
	if err != nil {
		writeDomainError(w, r, err)

		return
	}

	writeJSON(w, r, http.StatusOK, weeklyTestResponseOf(found))
}

func (h *weeklyTestHandler) checkAnswer(w http.ResponseWriter, r *http.Request) {
	var request checkWeeklyTestAnswerRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeDecodeError(w, r, err)

		return
	}

	checked, err := h.tests.CheckAnswer(
		r.Context(),
		ProfileIDFrom(r.Context()),
		weeklytest.ID(r.PathValue("testId")),
		weeklytest.Answer{QuestionID: request.QuestionID, OptionIndex: request.OptionIndex},
	)
	if err != nil {
		writeDomainError(w, r, err)

		return
	}

	writeJSON(w, r, http.StatusOK, checkWeeklyTestAnswerResponse{Correct: checked.Correct})
}
