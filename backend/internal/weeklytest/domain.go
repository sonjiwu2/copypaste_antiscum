// Package weeklytest управляет персональным еженедельным тестом.
package weeklytest

import (
	"context"
	"errors"
	"time"

	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/profile"
)

const (
	QuestionCount  = 20
	MaxLives       = 3
	PassingCorrect = 18
	RewardXP       = 100

	TestDuration    = 10 * time.Minute
	SubmissionGrace = 5 * time.Second
	Cooldown        = 7 * 24 * time.Hour
)

type ID string

type Source string

const (
	SourceGrok     Source = "grok"
	SourceGroq     Source = "groq"
	SourceFallback Source = "fallback"
)

var (
	ErrNotFound         = errors.New("еженедельный тест не найден")
	ErrForbidden        = errors.New("еженедельный тест принадлежит другому профилю")
	ErrInvalidAnswers   = errors.New("ответы на еженедельный тест некорректны")
	ErrInvalidGenerated = errors.New("сгенерированный тест некорректен")
	ErrAlreadyCompleted = errors.New("еженедельный тест уже завершён")
)

type Difficulty string

const (
	DifficultyEasy   Difficulty = "easy"
	DifficultyMedium Difficulty = "medium"
	DifficultyHard   Difficulty = "hard"
)

func (d Difficulty) Valid() bool {
	return d == DifficultyEasy || d == DifficultyMedium || d == DifficultyHard
}

type Question struct {
	ID           string     `json:"id"`
	Prompt       string     `json:"prompt"`
	Options      []string   `json:"options"`
	CorrectIndex int        `json:"correctIndex"`
	Explanation  string     `json:"explanation"`
	RiskTag      string     `json:"riskTag"`
	Difficulty   Difficulty `json:"difficulty"`
}

type Answer struct {
	QuestionID  string `json:"questionId"`
	OptionIndex int    `json:"optionIndex"`
}

type Submission struct {
	Answers     []Answer  `json:"answers"`
	Correct     int       `json:"correct"`
	Total       int       `json:"total"`
	Score       int       `json:"score"`
	Passed      bool      `json:"passed"`
	TimedOut    bool      `json:"timedOut"`
	LivesLeft   int       `json:"livesLeft"`
	EarnedXP    int       `json:"earnedXp"`
	CompletedAt time.Time `json:"completedAt"`
}

type Test struct {
	ID         ID
	Owner      profile.ID
	WeekStart  time.Time
	Title      string
	Intro      string
	Source     Source
	Model      string
	Questions  []Question
	CreatedAt  time.Time
	Submission *Submission
}

type Generated struct {
	Title     string
	Intro     string
	Source    Source
	Model     string
	Questions []Question
}

type AnswerCheck struct {
	Correct bool
}

func (t Test) ExpiresAt() time.Time {
	return t.CreatedAt.Add(TestDuration)
}

func (t Test) NextAvailableAt() time.Time {
	if t.Submission == nil {
		return time.Time{}
	}

	return t.Submission.CompletedAt.Add(Cooldown)
}

// GenerationContext намеренно не содержит данные профиля или историю:
// сторонний API получает только общую инструкцию создать учебный тест.
type GenerationContext struct{}

type Generator interface {
	Generate(ctx context.Context, input GenerationContext) (Generated, error)
}

type Repository interface {
	FindByWeek(ctx context.Context, owner profile.ID, weekStart time.Time) (Test, error)
	FindLatest(ctx context.Context, owner profile.ID) (Test, error)
	FindByID(ctx context.Context, id ID) (Test, error)
	Create(ctx context.Context, created Test) (Test, error)
	SaveSubmission(ctx context.Context, id ID, owner profile.ID, submission Submission) (Test, error)
}
