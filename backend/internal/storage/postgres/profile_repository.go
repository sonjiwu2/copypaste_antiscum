package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/profile"
)

// ProfileRepository хранит анонимные профили в PostgreSQL.
type ProfileRepository struct {
	pool         *pgxpool.Pool
	queryTimeout time.Duration
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
