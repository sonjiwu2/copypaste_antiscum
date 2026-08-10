package weeklytest

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"golang.org/x/sync/singleflight"

	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/platform/clock"
	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/platform/identifier"
	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/profile"
)

type Service struct {
	repository Repository
	primary    Generator
	fallback   Generator
	clock      clock.Clock
	ids        identifier.Generator
	logger     *slog.Logger
	requests   singleflight.Group
}

func NewService(
	repository Repository,
	primary Generator,
	fallback Generator,
	currentTime clock.Clock,
	ids identifier.Generator,
	logger *slog.Logger,
) *Service {
	return &Service{
		repository: repository,
		primary:    primary,
		fallback:   fallback,
		clock:      currentTime,
		ids:        ids,
		logger:     logger,
	}
}

// Current возвращает активный тест либо создаёт новый после семидневной паузы.
func (s *Service) Current(ctx context.Context, owner profile.ID) (Test, error) {
	if owner == "" {
		return Test{}, profile.ErrEmptyID
	}

	now := s.clock.Now().UTC()
	found, err := s.repository.FindLatest(ctx, owner)
	if err == nil && !canCreateAfter(found, now) {
		return found, nil
	}
	if err != nil && !errors.Is(err, ErrNotFound) {
		return Test{}, fmt.Errorf("прочитать еженедельный тест: %w", err)
	}

	cycleStart := CycleStart(now)
	key := string(owner) + ":" + cycleStart.Format(time.DateOnly)
	created, err, _ := s.requests.Do(key, func() (any, error) {
		existing, findErr := s.repository.FindLatest(ctx, owner)
		if findErr == nil && !canCreateAfter(existing, now) {
			return existing, nil
		}
		if findErr != nil && !errors.Is(findErr, ErrNotFound) {
			return Test{}, fmt.Errorf("повторно прочитать еженедельный тест: %w", findErr)
		}

		generated := s.generate(ctx)
		normalized, normalizeErr := normalizeGenerated(generated)
		if normalizeErr != nil {
			s.logger.WarnContext(ctx, "ИИ вернул некорректный тест, используется резервный",
				slog.String("error", normalizeErr.Error()))
			normalized = s.mustFallback(ctx)
		}

		testID, idErr := s.ids.NewID()
		if idErr != nil {
			return Test{}, fmt.Errorf("создать идентификатор теста: %w", idErr)
		}

		for index := range normalized.Questions {
			normalized.Questions[index].ID = fmt.Sprintf("q%d", index+1)
		}
		createdAt := s.clock.Now().UTC()

		return s.repository.Create(ctx, Test{
			ID: ID(testID), Owner: owner, WeekStart: cycleStart,
			Title: normalized.Title, Intro: normalized.Intro,
			Source: normalized.Source, Model: normalized.Model,
			Questions: normalized.Questions, CreatedAt: createdAt,
		})
	})
	if err != nil {
		return Test{}, err
	}

	return created.(Test), nil
}

func canCreateAfter(found Test, now time.Time) bool {
	if len(found.Questions) != QuestionCount {
		return true
	}

	return found.Submission != nil && !now.Before(found.NextAvailableAt())
}

func (s *Service) generate(ctx context.Context) Generated {
	input := GenerationContext{}
	if s.primary != nil {
		generated, err := s.primary.Generate(ctx, input)
		if err == nil {
			return generated
		}

		s.logger.WarnContext(ctx, "ИИ-генератор недоступен, используется резервный тест",
			slog.String("error", err.Error()))
	}

	return s.mustFallback(ctx)
}

func (s *Service) mustFallback(ctx context.Context) Generated {
	generated, err := s.fallback.Generate(ctx, GenerationContext{})
	if err != nil {
		panic("резервный генератор еженедельного теста завершился ошибкой: " + err.Error())
	}

	return generated
}

// CheckAnswer проверяет один вариант, не раскрывая индекс правильного ответа.
// Итоговая награда всё равно рассчитывается повторно сервером при Submit.
func (s *Service) CheckAnswer(
	ctx context.Context,
	owner profile.ID,
	id ID,
	answer Answer,
) (AnswerCheck, error) {
	if owner == "" {
		return AnswerCheck{}, profile.ErrEmptyID
	}
	if answer.QuestionID == "" || answer.OptionIndex < 0 || answer.OptionIndex >= 4 {
		return AnswerCheck{}, ErrInvalidAnswers
	}

	found, err := s.repository.FindByID(ctx, id)
	if err != nil {
		return AnswerCheck{}, err
	}
	if found.Owner != owner {
		return AnswerCheck{}, ErrForbidden
	}
	if found.Submission != nil {
		return AnswerCheck{}, ErrAlreadyCompleted
	}

	for _, question := range found.Questions {
		if question.ID == answer.QuestionID {
			return AnswerCheck{Correct: answer.OptionIndex == question.CorrectIndex}, nil
		}
	}

	return AnswerCheck{}, ErrInvalidAnswers
}

// Submit проверяет ответы на сервере и сохраняет первый результат.
// Повторная отправка идемпотентна: возвращает уже записанный результат.
func (s *Service) Submit(ctx context.Context, owner profile.ID, id ID, answers []Answer) (Test, error) {
	if owner == "" {
		return Test{}, profile.ErrEmptyID
	}

	found, err := s.repository.FindByID(ctx, id)
	if err != nil {
		return Test{}, err
	}
	if found.Owner != owner {
		return Test{}, ErrForbidden
	}
	if found.Submission != nil {
		return found, nil
	}

	ordered, correct, err := validateAnswers(found.Questions, answers)
	if err != nil {
		return Test{}, err
	}

	now := s.clock.Now().UTC()
	total := len(found.Questions)
	mistakes := total - correct
	livesLeft := MaxLives - mistakes
	if livesLeft < 0 {
		livesLeft = 0
	}
	timedOut := now.After(found.ExpiresAt().Add(SubmissionGrace))
	passed := correct >= PassingCorrect && !timedOut
	earnedXP := 0
	if passed {
		earnedXP = RewardXP
	}
	submission := Submission{
		Answers:     ordered,
		Correct:     correct,
		Total:       total,
		Score:       (correct * 100) / total,
		Passed:      passed,
		TimedOut:    timedOut,
		LivesLeft:   livesLeft,
		EarnedXP:    earnedXP,
		CompletedAt: now,
	}

	return s.repository.SaveSubmission(ctx, id, owner, submission)
}

func CycleStart(moment time.Time) time.Time {
	utc := moment.UTC()
	return time.Date(utc.Year(), utc.Month(), utc.Day(), 0, 0, 0, 0, time.UTC)
}

func normalizeGenerated(source Generated) (Generated, error) {
	source.Title = strings.TrimSpace(source.Title)
	source.Intro = strings.TrimSpace(source.Intro)
	if source.Title == "" || source.Intro == "" || len(source.Questions) != QuestionCount {
		return Generated{}, ErrInvalidGenerated
	}

	for index := range source.Questions {
		question := &source.Questions[index]
		question.Prompt = strings.TrimSpace(question.Prompt)
		question.Explanation = strings.TrimSpace(question.Explanation)
		question.RiskTag = strings.TrimSpace(question.RiskTag)
		if question.Prompt == "" || question.Explanation == "" || question.RiskTag == "" ||
			len(question.Options) != 4 || question.CorrectIndex < 0 || question.CorrectIndex >= 4 ||
			!question.Difficulty.Valid() {
			return Generated{}, fmt.Errorf("%w: вопрос %d", ErrInvalidGenerated, index+1)
		}

		seen := make(map[string]struct{}, 4)
		for optionIndex := range question.Options {
			question.Options[optionIndex] = strings.TrimSpace(question.Options[optionIndex])
			normalized := strings.ToLower(question.Options[optionIndex])
			if normalized == "" {
				return Generated{}, fmt.Errorf("%w: пустой вариант в вопросе %d", ErrInvalidGenerated, index+1)
			}
			if _, exists := seen[normalized]; exists {
				return Generated{}, fmt.Errorf("%w: повтор варианта в вопросе %d", ErrInvalidGenerated, index+1)
			}
			seen[normalized] = struct{}{}
		}
	}

	if source.Source != SourceGrok && source.Source != SourceGroq && source.Source != SourceFallback {
		return Generated{}, ErrInvalidGenerated
	}

	return source, nil
}

func validateAnswers(questions []Question, answers []Answer) ([]Answer, int, error) {
	if len(answers) > len(questions) {
		return nil, 0, ErrInvalidAnswers
	}

	provided := make(map[string]int, len(answers))
	for _, answer := range answers {
		if answer.QuestionID == "" || answer.OptionIndex < 0 || answer.OptionIndex >= 4 {
			return nil, 0, ErrInvalidAnswers
		}
		if _, duplicate := provided[answer.QuestionID]; duplicate {
			return nil, 0, ErrInvalidAnswers
		}
		provided[answer.QuestionID] = answer.OptionIndex
	}

	ordered := make([]Answer, 0, len(answers))
	correct := 0
	for _, question := range questions {
		optionIndex, exists := provided[question.ID]
		if !exists {
			continue
		}
		ordered = append(ordered, Answer{QuestionID: question.ID, OptionIndex: optionIndex})
		if optionIndex == question.CorrectIndex {
			correct++
		}
	}

	return ordered, correct, nil
}
