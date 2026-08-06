package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/attempt"
	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/profile"
	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/progress"
	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/scenario"
)

// ProgressRepository читает факты прогресса из PostgreSQL.
//
// Репозиторий только выбирает сохранённые строки: расчёты живут в домене
// прогресса, а результат попытки не пересчитывается никогда.
type ProgressRepository struct {
	pool         *pgxpool.Pool
	queryTimeout time.Duration
}

// NewProgressRepository создаёт хранилище прогресса.
func NewProgressRepository(pool *pgxpool.Pool, queryTimeout time.Duration) *ProgressRepository {
	return &ProgressRepository{pool: pool, queryTimeout: queryTimeout}
}

// selectCompletedAttemptsQuery выбирает завершённые прохождения профиля.
// Порядок задан явно, чтобы ответ не зависел от плана запроса.
const selectCompletedAttemptsQuery = `
SELECT id, scenario_id, score, outcome, started_at, completed_at
  FROM attempts
 WHERE profile_id = $1
   AND status = 'completed'
 ORDER BY completed_at ASC, id ASC`

// selectCompletedDecisionsQuery выбирает решения завершённых прохождений.
//
// Тексты последствий не читаются: прогрессу нужны только критичность,
// метки риска, эффекты навыков и дельта результата.
const selectCompletedDecisionsQuery = `
SELECT d.attempt_id, d.criticality, d.risk_tags, d.skill_effects, d.score_delta, d.created_at
  FROM attempt_decisions AS d
  JOIN attempts AS a ON a.id = d.attempt_id
 WHERE a.profile_id = $1
   AND a.status = 'completed'
 ORDER BY d.attempt_id ASC, d.ordinal ASC`

const selectActiveAttemptsQuery = `
SELECT id, scenario_id, updated_at
  FROM attempts
 WHERE profile_id = $1
   AND status = 'in_progress'
 ORDER BY updated_at DESC, id DESC`

// Facts возвращает сохранённые данные профиля.
//
// Чтение идёт одной транзакцией с повторяемым чтением: сводка и история
// обязаны описывать одно и то же состояние, иначе параллельно завершённая
// попытка попала бы в один список и отсутствовала в другом.
func (r *ProgressRepository) Facts(ctx context.Context, owner profile.ID) (progress.Facts, error) {
	if err := ctx.Err(); err != nil {
		return progress.Facts{}, err
	}

	if owner == "" {
		return progress.Facts{}, profile.ErrEmptyID
	}

	queryCtx, cancel := WithTimeout(ctx, r.queryTimeout)
	defer cancel()

	tx, err := r.pool.BeginTx(queryCtx, pgx.TxOptions{
		IsoLevel:   pgx.RepeatableRead,
		AccessMode: pgx.ReadOnly,
	})
	if err != nil {
		return progress.Facts{}, fmt.Errorf("начать чтение прогресса: %w", MapError(err))
	}

	defer func() { _ = tx.Rollback(queryCtx) }()

	completed, err := r.completedAttempts(queryCtx, tx, owner)
	if err != nil {
		return progress.Facts{}, err
	}

	if err := r.attachDecisions(queryCtx, tx, owner, completed); err != nil {
		return progress.Facts{}, err
	}

	active, err := r.activeAttempts(queryCtx, tx, owner)
	if err != nil {
		return progress.Facts{}, err
	}

	facts := progress.Facts{
		Completed: make([]progress.CompletedAttempt, 0, len(completed)),
		Active:    active,
	}

	for _, item := range completed {
		facts.Completed = append(facts.Completed, *item)
	}

	return facts, nil
}

// completedAttempts читает завершённые попытки, сохраняя порядок выборки.
func (r *ProgressRepository) completedAttempts(
	ctx context.Context,
	tx pgx.Tx,
	owner profile.ID,
) ([]*progress.CompletedAttempt, error) {
	rows, err := tx.Query(ctx, selectCompletedAttemptsQuery, string(owner))
	if err != nil {
		return nil, fmt.Errorf("прочитать завершённые попытки: %w", MapError(err))
	}

	defer rows.Close()

	completed := make([]*progress.CompletedAttempt, 0)

	for rows.Next() {
		var (
			id          string
			scenarioID  string
			score       int32
			outcome     *string
			startedAt   time.Time
			completedAt time.Time
		)

		if err := rows.Scan(&id, &scenarioID, &score, &outcome, &startedAt, &completedAt); err != nil {
			return nil, fmt.Errorf("прочитать завершённую попытку: %w", MapError(err))
		}

		completed = append(completed, &progress.CompletedAttempt{
			AttemptID:   attempt.ID(id),
			ScenarioID:  scenario.ID(scenarioID),
			Score:       int(score),
			Outcome:     scenario.OutcomeType(stringValue(outcome)),
			StartedAt:   startedAt,
			CompletedAt: completedAt,
			Decisions:   []progress.DecisionFact{},
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("прочитать завершённые попытки: %w", MapError(err))
	}

	return completed, nil
}

// attachDecisions раскладывает решения по попыткам одним запросом.
// Отдельный запрос на каждую попытку дал бы N+1 обращений к базе.
func (r *ProgressRepository) attachDecisions(
	ctx context.Context,
	tx pgx.Tx,
	owner profile.ID,
	completed []*progress.CompletedAttempt,
) error {
	if len(completed) == 0 {
		return nil
	}

	byID := make(map[attempt.ID]*progress.CompletedAttempt, len(completed))
	for _, item := range completed {
		byID[item.AttemptID] = item
	}

	rows, err := tx.Query(ctx, selectCompletedDecisionsQuery, string(owner))
	if err != nil {
		return fmt.Errorf("прочитать решения прогресса: %w", MapError(err))
	}

	defer rows.Close()

	for rows.Next() {
		var (
			attemptID    string
			criticality  string
			riskTags     []string
			skillEffects []byte
			scoreDelta   int32
			createdAt    time.Time
		)

		err := rows.Scan(&attemptID, &criticality, &riskTags, &skillEffects, &scoreDelta, &createdAt)
		if err != nil {
			return fmt.Errorf("прочитать решение прогресса: %w", MapError(err))
		}

		owner, found := byID[attempt.ID(attemptID)]
		if !found {
			continue
		}

		effects, err := unmarshalSkillEffects(skillEffects)
		if err != nil {
			return err
		}

		tags := make([]scenario.RiskTag, 0, len(riskTags))
		for _, tag := range riskTags {
			tags = append(tags, scenario.RiskTag(tag))
		}

		owner.Decisions = append(owner.Decisions, progress.DecisionFact{
			Criticality:  scenario.Criticality(criticality),
			RiskTags:     tags,
			SkillEffects: effects,
			ScoreDelta:   int(scoreDelta),
			CreatedAt:    createdAt,
		})
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("прочитать решения прогресса: %w", MapError(err))
	}

	return nil
}

func (r *ProgressRepository) activeAttempts(
	ctx context.Context,
	tx pgx.Tx,
	owner profile.ID,
) ([]progress.ActiveAttempt, error) {
	rows, err := tx.Query(ctx, selectActiveAttemptsQuery, string(owner))
	if err != nil {
		return nil, fmt.Errorf("прочитать активные попытки: %w", MapError(err))
	}

	defer rows.Close()

	active := make([]progress.ActiveAttempt, 0)

	for rows.Next() {
		var (
			id         string
			scenarioID string
			updatedAt  time.Time
		)

		if err := rows.Scan(&id, &scenarioID, &updatedAt); err != nil {
			return nil, fmt.Errorf("прочитать активную попытку: %w", MapError(err))
		}

		active = append(active, progress.ActiveAttempt{
			AttemptID:  attempt.ID(id),
			ScenarioID: scenario.ID(scenarioID),
			UpdatedAt:  updatedAt,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("прочитать активные попытки: %w", MapError(err))
	}

	return active, nil
}
