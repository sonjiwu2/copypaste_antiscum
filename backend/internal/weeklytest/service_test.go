package weeklytest_test

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/platform/clock"
	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/platform/identifier"
	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/profile"
	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/storage/memory"
	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/weeklytest"
)

func TestWeeklyExamRewardsOnlyPassingResultAndOpensAfterSevenDays(t *testing.T) {
	start := time.Date(2026, time.August, 9, 12, 0, 0, 0, time.UTC)
	currentTime := &clock.Fixed{Moment: start}
	service := newWeeklyTestService(currentTime)
	owner := profile.ID("anonymous-profile-test")

	created, err := service.Current(context.Background(), owner)
	if err != nil {
		t.Fatalf("Current: %v", err)
	}
	if len(created.Questions) != weeklytest.QuestionCount {
		t.Fatalf("вопросов = %d, ожидалось %d", len(created.Questions), weeklytest.QuestionCount)
	}

	completed, err := service.Submit(context.Background(), owner, created.ID, correctAnswers(created))
	if err != nil {
		t.Fatalf("Submit: %v", err)
	}
	if completed.Submission == nil || !completed.Submission.Passed ||
		completed.Submission.EarnedXP != weeklytest.RewardXP || completed.Submission.LivesLeft != weeklytest.MaxLives {
		t.Fatalf("неверный успешный результат: %+v", completed.Submission)
	}

	currentTime.Moment = start.Add(weeklytest.Cooldown - time.Second)
	locked, err := service.Current(context.Background(), owner)
	if err != nil {
		t.Fatalf("Current до cooldown: %v", err)
	}
	if locked.ID != created.ID {
		t.Fatal("до истечения 7 дней был создан новый экзамен")
	}

	currentTime.Moment = start.Add(weeklytest.Cooldown + time.Second)
	next, err := service.Current(context.Background(), owner)
	if err != nil {
		t.Fatalf("Current после cooldown: %v", err)
	}
	if next.ID == created.ID {
		t.Fatal("после истечения 7 дней новый экзамен не создан")
	}
}

func TestWeeklyExamThreeMistakesRemoveReward(t *testing.T) {
	start := time.Date(2026, time.August, 9, 12, 0, 0, 0, time.UTC)
	service := newWeeklyTestService(&clock.Fixed{Moment: start})
	owner := profile.ID("anonymous-profile-fail")

	created, err := service.Current(context.Background(), owner)
	if err != nil {
		t.Fatalf("Current: %v", err)
	}
	answers := correctAnswers(created)
	for index := 0; index < weeklytest.MaxLives; index++ {
		answers[index].OptionIndex = (answers[index].OptionIndex + 1) % 4
	}

	completed, err := service.Submit(context.Background(), owner, created.ID, answers)
	if err != nil {
		t.Fatalf("Submit: %v", err)
	}
	if completed.Submission == nil || completed.Submission.Passed ||
		completed.Submission.EarnedXP != 0 || completed.Submission.LivesLeft != 0 {
		t.Fatalf("три ошибки должны завершить экзамен без награды: %+v", completed.Submission)
	}
}

func TestWeeklyExamTimeoutRemovesReward(t *testing.T) {
	start := time.Date(2026, time.August, 9, 12, 0, 0, 0, time.UTC)
	currentTime := &clock.Fixed{Moment: start}
	service := newWeeklyTestService(currentTime)
	owner := profile.ID("anonymous-profile-timeout")

	created, err := service.Current(context.Background(), owner)
	if err != nil {
		t.Fatalf("Current: %v", err)
	}
	currentTime.Moment = start.Add(weeklytest.TestDuration + weeklytest.SubmissionGrace + time.Second)

	completed, err := service.Submit(context.Background(), owner, created.ID, correctAnswers(created))
	if err != nil {
		t.Fatalf("Submit: %v", err)
	}
	if completed.Submission == nil || completed.Submission.Passed ||
		!completed.Submission.TimedOut || completed.Submission.EarnedXP != 0 {
		t.Fatalf("просроченный экзамен должен завершиться без награды: %+v", completed.Submission)
	}
}

func newWeeklyTestService(currentTime *clock.Fixed) *weeklytest.Service {
	return weeklytest.NewService(
		memory.NewWeeklyTestRepository(),
		nil,
		weeklytest.FallbackGenerator{},
		currentTime,
		&identifier.Sequential{Prefix: "weekly-test"},
		slog.New(slog.NewTextHandler(io.Discard, nil)),
	)
}

func correctAnswers(test weeklytest.Test) []weeklytest.Answer {
	answers := make([]weeklytest.Answer, 0, len(test.Questions))
	for _, question := range test.Questions {
		answers = append(answers, weeklytest.Answer{
			QuestionID: question.ID, OptionIndex: question.CorrectIndex,
		})
	}

	return answers
}
