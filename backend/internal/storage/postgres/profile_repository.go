package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/profile"
)

// ProfileRepository хранит анонимные профили в PostgreSQL.
type ProfileRepository struct {
	pool         *pgxpool.Pool
	queryTimeout time.Duration
}

// Reset удаляет все пользовательские данные одной анонимной cookie в одной
// транзакции. Решения попыток удаляются каскадом вместе с attempts.
func (r *ProfileRepository) Reset(ctx context.Context, id profile.ID) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if id == "" {
		return profile.ErrEmptyID
	}

	queryCtx, cancel := WithTimeout(ctx, r.queryTimeout)
	defer cancel()

	tx, err := r.pool.BeginTx(queryCtx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("начать сброс профиля: %w", MapError(err))
	}
	defer func() { _ = tx.Rollback(queryCtx) }()

	for _, query := range []string{
		"DELETE FROM weekly_tests WHERE profile_id = $1",
		"DELETE FROM attempts WHERE profile_id = $1",
		"DELETE FROM profiles WHERE id = $1 AND id <> 'profile_default_demo'",
	} {
		if _, err := tx.Exec(queryCtx, query, string(id)); err != nil {
			return fmt.Errorf("очистить данные профиля: %w", MapError(err))
		}
	}

	if err := tx.Commit(queryCtx); err != nil {
		return fmt.Errorf("завершить сброс профиля: %w", MapError(err))
	}

	return nil
}

// NewProfileRepository создаёт хранилище профилей.
func NewProfileRepository(pool *pgxpool.Pool, queryTimeout time.Duration) *ProfileRepository {
	return &ProfileRepository{pool: pool, queryTimeout: queryTimeout}
}

// ensureProfileQuery создаёт профиль или отмечает его новое посещение.
//
// created_at не переписывается: он показывает, когда пользователь пришёл
// впервые, и это единственная историческая метка анонимного профиля.
const ensureProfileQuery = `
INSERT INTO profiles (id, created_at, updated_at)
VALUES ($1, $2, $3)
ON CONFLICT (id) DO UPDATE
   SET updated_at = EXCLUDED.updated_at`

// Ensure создаёт профиль, если его ещё нет.
func (r *ProfileRepository) Ensure(ctx context.Context, created profile.Profile) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	if created.ID == "" {
		return profile.ErrEmptyID
	}

	queryCtx, cancel := WithTimeout(ctx, r.queryTimeout)
	defer cancel()

	_, err := r.pool.Exec(queryCtx, ensureProfileQuery,
		string(created.ID), created.CreatedAt, created.UpdatedAt)
	if err != nil {
		return fmt.Errorf("сохранить профиль: %w", MapError(err))
	}

	return nil
}

func (r *ProfileRepository) UpdateIdentity(
	ctx context.Context,
	id profile.ID,
	identity profile.Identity,
	updatedAt time.Time,
) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	queryCtx, cancel := WithTimeout(ctx, r.queryTimeout)
	defer cancel()
	result, err := r.pool.Exec(queryCtx, `
UPDATE profiles
   SET display_name = $2, avatar = $3, updated_at = $4
 WHERE id = $1`, string(id), identity.DisplayName, string(identity.Avatar), updatedAt)
	if err != nil {
		return fmt.Errorf("обновить публичный профиль: %w", MapError(err))
	}
	if result.RowsAffected() == 0 {
		return profile.ErrInvalidID
	}
	return nil
}

const leaderboardQuery = `
WITH best_attempts AS (
    SELECT profile_id, scenario_id, max(score)::integer AS best_score,
           max(completed_at) AS last_completed_at
      FROM attempts
     WHERE status = 'completed'
     GROUP BY profile_id, scenario_id
), aggregated AS (
    SELECT p.id, p.display_name, p.avatar,
           count(b.scenario_id)::integer AS completed_scenarios,
           coalesce(sum(b.best_score), 0)::integer AS rating,
           coalesce(round(avg(b.best_score)), 0)::integer AS average_score,
           max(b.last_completed_at) AS last_completed_at
      FROM profiles p
      JOIN accounts account ON account.profile_id = p.id
      JOIN best_attempts b ON b.profile_id = p.id
     WHERE p.id <> 'profile_default_demo'
     GROUP BY p.id, p.display_name, p.avatar
), ranked AS (
    SELECT *, row_number() OVER (
        ORDER BY rating DESC, completed_scenarios DESC, average_score DESC,
                 last_completed_at ASC, id ASC
    )::integer AS rank
      FROM aggregated
)
SELECT id, display_name, avatar, completed_scenarios, rating, average_score, rank
  FROM ranked
 WHERE rank <= $2 OR id = $1
 ORDER BY rank ASC`

func (r *ProfileRepository) Leaderboard(
	ctx context.Context,
	current profile.ID,
	limit int,
) (profile.Leaderboard, error) {
	if err := ctx.Err(); err != nil {
		return profile.Leaderboard{}, err
	}
	queryCtx, cancel := WithTimeout(ctx, r.queryTimeout)
	defer cancel()
	rows, err := r.pool.Query(queryCtx, leaderboardQuery, string(current), limit)
	if err != nil {
		return profile.Leaderboard{}, fmt.Errorf("прочитать рейтинг: %w", MapError(err))
	}
	defer rows.Close()
	board := profile.Leaderboard{Leaders: []profile.Leader{}}
	for rows.Next() {
		var id, name, avatar string
		var completed, rating, average, rank int32
		if err := rows.Scan(&id, &name, &avatar, &completed, &rating, &average, &rank); err != nil {
			return profile.Leaderboard{}, fmt.Errorf("прочитать строку рейтинга: %w", MapError(err))
		}
		entry := profile.Leader{
			Rank: int(rank), DisplayName: name, Avatar: profile.Avatar(avatar),
			Rating: int(rating), CompletedScenarios: int(completed),
			AverageScore: int(average), CurrentPlayer: profile.ID(id) == current,
		}
		if entry.Rank <= limit {
			board.Leaders = append(board.Leaders, entry)
		}
		if entry.CurrentPlayer {
			copy := entry
			board.Current = &copy
		}
	}
	if err := rows.Err(); err != nil {
		return profile.Leaderboard{}, fmt.Errorf("прочитать рейтинг: %w", MapError(err))
	}
	return board, nil
}
