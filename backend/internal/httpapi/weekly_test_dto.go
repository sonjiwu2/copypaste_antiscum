package httpapi

import (
	"time"

	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/weeklytest"
)

type submitWeeklyTestRequest struct {
	Answers []weeklyTestAnswer `json:"answers"`
}

type checkWeeklyTestAnswerRequest struct {
	QuestionID  string `json:"questionId"`
	OptionIndex int    `json:"optionIndex"`
}

type checkWeeklyTestAnswerResponse struct {
	Correct bool `json:"correct"`
}

type weeklyTestAnswer struct {
	QuestionID  string `json:"questionId"`
	OptionIndex int    `json:"optionIndex"`
}

type weeklyTestResponse struct {
	TestID    string               `json:"testId"`
	WeekStart string               `json:"weekStart"`
	Title     string               `json:"title"`
	Intro     string               `json:"intro"`
	Source    string               `json:"source"`
	Completed bool                 `json:"completed"`
	ExpiresAt string               `json:"expiresAt"`
	NextAt    string               `json:"nextAvailableAt,omitempty"`
	Rules     weeklyTestRules      `json:"rules"`
	Questions []weeklyTestQuestion `json:"questions"`
	Result    *weeklyTestResult    `json:"result,omitempty"`
}

type weeklyTestRules struct {
	QuestionCount  int `json:"questionCount"`
	MaxLives       int `json:"maxLives"`
	PassingCorrect int `json:"passingCorrect"`
	RewardXP       int `json:"rewardXp"`
	Duration       int `json:"durationSeconds"`
}

type weeklyTestQuestion struct {
	ID         string   `json:"id"`
	Prompt     string   `json:"prompt"`
	Options    []string `json:"options"`
	RiskTag    string   `json:"riskTag"`
	Difficulty string   `json:"difficulty"`
}

type weeklyTestResult struct {
	Correct     int                `json:"correct"`
	Total       int                `json:"total"`
	Score       int                `json:"score"`
	Passed      bool               `json:"passed"`
	TimedOut    bool               `json:"timedOut"`
	LivesLeft   int                `json:"livesLeft"`
	EarnedXP    int                `json:"earnedXp"`
	CompletedAt string             `json:"completedAt"`
	Review      []weeklyTestReview `json:"review"`
}

type weeklyTestReview struct {
	QuestionID    string `json:"questionId"`
	SelectedIndex int    `json:"selectedIndex"`
	CorrectIndex  int    `json:"correctIndex"`
	Correct       bool   `json:"correct"`
	Explanation   string `json:"explanation"`
}

func weeklyTestResponseOf(source weeklytest.Test) weeklyTestResponse {
	response := weeklyTestResponse{
		TestID:    string(source.ID),
		WeekStart: source.WeekStart.Format(time.DateOnly),
		Title:     source.Title,
		Intro:     source.Intro,
		Source:    string(source.Source),
		Completed: source.Submission != nil,
		ExpiresAt: source.ExpiresAt().Format(time.RFC3339),
		Rules: weeklyTestRules{
			QuestionCount: weeklytest.QuestionCount, MaxLives: weeklytest.MaxLives,
			PassingCorrect: weeklytest.PassingCorrect, RewardXP: weeklytest.RewardXP,
			Duration: int(weeklytest.TestDuration.Seconds()),
		},
		Questions: make([]weeklyTestQuestion, 0, len(source.Questions)),
	}
	for _, question := range source.Questions {
		response.Questions = append(response.Questions, weeklyTestQuestion{
			ID:         question.ID,
			Prompt:     question.Prompt,
			Options:    append([]string(nil), question.Options...),
			RiskTag:    question.RiskTag,
			Difficulty: string(question.Difficulty),
		})
	}

	if source.Submission == nil {
		return response
	}

	selected := make(map[string]int, len(source.Submission.Answers))
	for _, answer := range source.Submission.Answers {
		selected[answer.QuestionID] = answer.OptionIndex
	}

	result := &weeklyTestResult{
		Correct:     source.Submission.Correct,
		Total:       source.Submission.Total,
		Score:       source.Submission.Score,
		Passed:      source.Submission.Passed,
		TimedOut:    source.Submission.TimedOut,
		LivesLeft:   source.Submission.LivesLeft,
		EarnedXP:    source.Submission.EarnedXP,
		CompletedAt: source.Submission.CompletedAt.Format(time.RFC3339),
		Review:      make([]weeklyTestReview, 0, len(source.Questions)),
	}
	response.NextAt = source.NextAvailableAt().Format(time.RFC3339)
	for _, question := range source.Questions {
		selectedIndex, answered := selected[question.ID]
		if !answered {
			selectedIndex = -1
		}
		result.Review = append(result.Review, weeklyTestReview{
			QuestionID:    question.ID,
			SelectedIndex: selectedIndex,
			CorrectIndex:  question.CorrectIndex,
			Correct:       selectedIndex == question.CorrectIndex,
			Explanation:   question.Explanation,
		})
	}
	response.Result = result

	return response
}
