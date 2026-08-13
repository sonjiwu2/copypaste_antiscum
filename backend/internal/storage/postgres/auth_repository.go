package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/auth"
	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/profile"
)

type AuthRepository struct {
	pool         *pgxpool.Pool
	queryTimeout time.Duration
}

func NewAuthRepository(pool *pgxpool.Pool, queryTimeout time.Duration) *AuthRepository {
	return &AuthRepository{pool: pool, queryTimeout: queryTimeout}
}

func (r *AuthRepository) Register(ctx context.Context, account auth.Account,
	emailNormalized, passwordHash string) error {
	queryCtx, cancel := WithTimeout(ctx, r.queryTimeout)
	defer cancel()
	tx, err := r.pool.BeginTx(queryCtx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("начать регистрацию: %w", MapError(err))
	}
	defer func() { _ = tx.Rollback(queryCtx) }()
	if _, err := tx.Exec(queryCtx, `
INSERT INTO accounts (id, profile_id, email, email_normalized, password_hash, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7)`, string(account.ID), string(account.ProfileID),
		account.Email, emailNormalized, passwordHash, account.CreatedAt, account.UpdatedAt); err != nil {
		return fmt.Errorf("создать аккаунт: %w", MapError(err))
	}
	result, err := tx.Exec(queryCtx, `UPDATE profiles SET display_name=$2, avatar=$3, updated_at=$4 WHERE id=$1`,
		string(account.ProfileID), account.Identity.DisplayName, string(account.Identity.Avatar), account.UpdatedAt)
	if err != nil {
		return fmt.Errorf("обновить профиль аккаунта: %w", MapError(err))
	}
	if result.RowsAffected() != 1 {
		return profile.ErrInvalidID
	}
	if err := tx.Commit(queryCtx); err != nil {
		return fmt.Errorf("завершить регистрацию: %w", MapError(err))
	}
	return nil
}

func (r *AuthRepository) CredentialsByEmail(ctx context.Context, normalized string) (auth.Credentials, error) {
	queryCtx, cancel := WithTimeout(ctx, r.queryTimeout)
	defer cancel()
	var credentials auth.Credentials
	var accountID, profileID, avatar string
	err := r.pool.QueryRow(queryCtx, `
SELECT a.id, a.profile_id, a.email, a.password_hash, p.display_name, p.avatar,
       a.created_at, a.updated_at
  FROM accounts a JOIN profiles p ON p.id=a.profile_id
 WHERE a.email_normalized=$1`, normalized).Scan(&accountID, &profileID, &credentials.Email,
		&credentials.PasswordHash, &credentials.Identity.DisplayName, &avatar,
		&credentials.CreatedAt, &credentials.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return auth.Credentials{}, auth.ErrInvalidCredentials
	}
	if err != nil {
		return auth.Credentials{}, fmt.Errorf("найти аккаунт: %w", MapError(err))
	}
	credentials.ID = auth.AccountID(accountID)
	credentials.ProfileID = profile.ID(profileID)
	credentials.Identity.Avatar = profile.Avatar(avatar)
	return credentials, nil
}

func (r *AuthRepository) CreateSession(ctx context.Context, session auth.Session) error {
	queryCtx, cancel := WithTimeout(ctx, r.queryTimeout)
	defer cancel()
	_, err := r.pool.Exec(queryCtx, `
INSERT INTO auth_sessions (token_hash, account_id, created_at, last_seen_at, expires_at)
VALUES ($1, $2, $3, $4, $5)`, session.TokenHash, string(session.AccountID),
		session.CreatedAt, session.LastSeen, session.ExpiresAt)
	if err != nil {
		return fmt.Errorf("создать сессию: %w", MapError(err))
	}
	return nil
}

func (r *AuthRepository) CurrentBySession(ctx context.Context, tokenHash string, now time.Time) (auth.Current, error) {
	queryCtx, cancel := WithTimeout(ctx, r.queryTimeout)
	defer cancel()
	var current auth.Current
	var accountID, profileID, avatar string
	err := r.pool.QueryRow(queryCtx, `
SELECT a.id, a.profile_id, a.email, p.display_name, p.avatar,
       a.created_at, a.updated_at, s.expires_at
  FROM auth_sessions s
  JOIN accounts a ON a.id=s.account_id
  JOIN profiles p ON p.id=a.profile_id
 WHERE s.token_hash=$1 AND s.expires_at>$2`, tokenHash, now).Scan(&accountID, &profileID,
		&current.Email, &current.Identity.DisplayName, &avatar,
		&current.CreatedAt, &current.UpdatedAt, &current.ExpiresAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return auth.Current{}, auth.ErrRequired
	}
	if err != nil {
		return auth.Current{}, fmt.Errorf("прочитать сессию: %w", MapError(err))
	}
	current.ID = auth.AccountID(accountID)
	current.ProfileID = profile.ID(profileID)
	current.Identity.Avatar = profile.Avatar(avatar)
	_, _ = r.pool.Exec(queryCtx, `UPDATE auth_sessions SET last_seen_at=$2 WHERE token_hash=$1`, tokenHash, now)
	return current, nil
}

func (r *AuthRepository) DeleteSession(ctx context.Context, tokenHash string) error {
	queryCtx, cancel := WithTimeout(ctx, r.queryTimeout)
	defer cancel()
	if _, err := r.pool.Exec(queryCtx, `DELETE FROM auth_sessions WHERE token_hash=$1`, tokenHash); err != nil {
		return fmt.Errorf("удалить сессию: %w", MapError(err))
	}
	return nil
}
