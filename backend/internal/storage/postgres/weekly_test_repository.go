package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/profile"
	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/weeklytest"
)

type WeeklyTestRepository struct {
	pool         *pgxpool.Pool
	queryTimeout time.Duration
}

func NewWeeklyTestRepository(pool *pgxpool.Pool, queryTimeout time.Duration) *WeeklyTestRepository {
	return &WeeklyTestRepository{pool: pool, queryTimeout: queryTimeout}
}

const selectWeeklyTestFields = `
SELECT id, profile_id, week_start, title, intro, source, model, questions,
       answers, correct_count, total_count, score, passed, timed_out, lives_left,
       earned_xp, completed_at, created_at
  FROM weekly_tests`

const insertWeeklyTestQuery = `
INSERT INTO weekly_tests (
    id, profile_id, week_start, title, intro, source, model, questions, created_at
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
ON CONFLICT (profile_id, week_start) DO NOTHING`

const saveWeeklyTestSubmissionQuery = `
UPDATE weekly_tests
   SET answers = $3,
       correct_count = $4,
       total_count = $5,
       score = $6,
       passed = $7,
       timed_out = $8,
       lives_left = $9,
       earned_xp = $10,
       completed_at = $11
 WHERE id = $1
   AND profile_id = $2
   AND completed_at IS NULL`

func (r *WeeklyTestRepository) FindByWeek(
	ctx context.Context,
	owner profile.ID,
	weekStart time.Time,
) (weeklytest.Test, error) {
	return r.find(ctx, selectWeeklyTestFields+" WHERE profile_id = $1 AND week_start = $2", owner, weekStart)
}

func (r *WeeklyTestRepository) FindLatest(
	ctx context.Context,
	owner profile.ID,
) (weeklytest.Test, error) {
	return r.find(ctx,
		selectWeeklyTestFields+" WHERE profile_id = $1 ORDER BY created_at DESC, id DESC LIMIT 1",
		owner,
	)
}

func (r *WeeklyTestRepository) FindByID(ctx context.Context, id weeklytest.ID) (weeklytest.Test, error) {
	return r.find(ctx, selectWeeklyTestFields+" WHERE id = $1", id)
}

func (r *WeeklyTestRepository) Create(
	ctx context.Context,
	created weeklytest.Test,
) (weeklytest.Test, error) {
	questions, err := json.Marshal(created.Questions)
	if err != nil {
		return weeklytest.Test{}, fmt.Errorf("сериализовать вопросы еженедельного теста: %w", err)
	}

	queryCtx, cancel := WithTimeout(ctx, r.queryTimeout)
	defer cancel()

	_, err = r.pool.Exec(queryCtx, insertWeeklyTestQuery,
		string(created.ID), string(created.Owner), created.WeekStart, created.Title, created.Intro,
		string(created.Source), created.Model, questions, created.CreatedAt)
	if err != nil {
		return weeklytest.Test{}, fmt.Errorf("сохранить еженедельный тест: %w", err)
	}

	// ON CONFLICT мог оставить тест, созданный конкурентным запросом.
	return r.findWithContext(queryCtx,
		selectWeeklyTestFields+" WHERE profile_id = $1 AND week_start = $2", created.Owner, created.WeekStart)
}

func (r *WeeklyTestRepository) SaveSubmission(
	ctx context.Context,
	id weeklytest.ID,
	owner profile.ID,
	submission weeklytest.Submission,
) (weeklytest.Test, error) {
	answers, err := json.Marshal(submission.Answers)
	if err != nil {
		return weeklytest.Test{}, fmt.Errorf("сериализовать ответы еженедельного теста: %w", err)
	}

	queryCtx, cancel := WithTimeout(ctx, r.queryTimeout)
	defer cancel()

	result, err := r.pool.Exec(queryCtx, saveWeeklyTestSubmissionQuery,
		string(id), string(owner), answers, submission.Correct, submission.Total,
		submission.Score, submission.Passed, submission.TimedOut, submission.LivesLeft,
		submission.EarnedXP, submission.CompletedAt)
	if err != nil {
		return weeklytest.Test{}, fmt.Errorf("сохранить результат еженедельного теста: %w", err)
	}

	found, err := r.findWithContext(queryCtx, selectWeeklyTestFields+" WHERE id = $1", id)
	if err != nil {
		return weeklytest.Test{}, err
	}
	if found.Owner != owner {
		return weeklytest.Test{}, weeklytest.ErrForbidden
	}
	// Ноль строк допустим только при идемпотентной повторной отправке.
	if result.RowsAffected() == 0 && found.Submission == nil {
		return weeklytest.Test{}, fmt.Errorf("результат еженедельного теста не был сохранён")
	}

	return found, nil
}

func (r *WeeklyTestRepository) find(ctx context.Context, query string, args ...any) (weeklytest.Test, error) {
	queryCtx, cancel := WithTimeout(ctx, r.queryTimeout)
	defer cancel()

	return r.findWithContext(queryCtx, query, args...)
}

func (r *WeeklyTestRepository) findWithContext(
	ctx context.Context,
	query string,
	args ...any,
) (weeklytest.Test, error) {
	var (
		found         weeklytest.Test
		owner         string
		source        string
		questionsJSON []byte
		answersJSON   []byte
		correct       *int
		total         *int
		score         *int
		passed        *bool
		timedOut      *bool
		livesLeft     *int
		earnedXP      *int
		completedAt   *time.Time
	)

	err := r.pool.QueryRow(ctx, query, args...).Scan(
		&found.ID, &owner, &found.WeekStart, &found.Title, &found.Intro, &source, &found.Model,
		&questionsJSON, &answersJSON, &correct, &total, &score, &passed, &timedOut, &livesLeft,
		&earnedXP, &completedAt, &found.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return weeklytest.Test{}, weeklytest.ErrNotFound
	}
	if err != nil {
		return weeklytest.Test{}, fmt.Errorf("прочитать еженедельный тест: %w", err)
	}

	found.Owner = profile.ID(owner)
	found.Source = weeklytest.Source(source)
	if err := json.Unmarshal(questionsJSON, &found.Questions); err != nil {
		return weeklytest.Test{}, fmt.Errorf("разобрать вопросы еженедельного теста: %w", err)
	}

	if completedAt != nil {
		if correct == nil || total == nil || score == nil || passed == nil || timedOut == nil ||
			livesLeft == nil || earnedXP == nil {
			return weeklytest.Test{}, fmt.Errorf("сохранён неполный результат еженедельного теста")
		}

		submission := weeklytest.Submission{
			Correct: *correct, Total: *total, Score: *score, Passed: *passed, TimedOut: *timedOut,
			LivesLeft: *livesLeft, EarnedXP: *earnedXP, CompletedAt: *completedAt,
		}
		if err := json.Unmarshal(answersJSON, &submission.Answers); err != nil {
			return weeklytest.Test{}, fmt.Errorf("разобрать ответы еженедельного теста: %w", err)
		}
		found.Submission = &submission
	}

	return found, nil
}
