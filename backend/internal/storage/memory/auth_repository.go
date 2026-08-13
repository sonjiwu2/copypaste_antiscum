package memory

import (
	"context"
	"sync"
	"time"

	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/auth"
	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/profile"
)

type AuthRepository struct {
	mu        sync.RWMutex
	profiles  *ProfileRepository
	accounts  map[auth.AccountID]auth.Credentials
	emails    map[string]auth.AccountID
	byProfile map[profile.ID]auth.AccountID
	sessions  map[string]auth.Session
}

func NewAuthRepository(profiles *ProfileRepository) *AuthRepository {
	return &AuthRepository{profiles: profiles, accounts: make(map[auth.AccountID]auth.Credentials),
		emails: make(map[string]auth.AccountID), byProfile: make(map[profile.ID]auth.AccountID),
		sessions: make(map[string]auth.Session)}
}

func (r *AuthRepository) Register(ctx context.Context, account auth.Account,
	emailNormalized, passwordHash string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.emails[emailNormalized]; exists {
		return auth.ErrEmailTaken
	}
	if _, exists := r.byProfile[account.ProfileID]; exists {
		return auth.ErrProfileClaimed
	}
	r.profiles.mu.Lock()
	stored, exists := r.profiles.profiles[account.ProfileID]
	if !exists {
		r.profiles.mu.Unlock()
		return profile.ErrInvalidID
	}
	stored.Identity = account.Identity
	stored.UpdatedAt = account.UpdatedAt
	r.profiles.profiles[account.ProfileID] = stored
	r.profiles.mu.Unlock()
	r.accounts[account.ID] = auth.Credentials{Account: account, PasswordHash: passwordHash}
	r.emails[emailNormalized] = account.ID
	r.byProfile[account.ProfileID] = account.ID
	return nil
}

func (r *AuthRepository) CredentialsByEmail(ctx context.Context, normalized string) (auth.Credentials, error) {
	if err := ctx.Err(); err != nil {
		return auth.Credentials{}, err
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	id, exists := r.emails[normalized]
	if !exists {
		return auth.Credentials{}, auth.ErrInvalidCredentials
	}
	return r.accounts[id], nil
}

func (r *AuthRepository) CreateSession(ctx context.Context, session auth.Session) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.sessions[session.TokenHash] = session
	return nil
}

func (r *AuthRepository) CurrentBySession(ctx context.Context, tokenHash string, now time.Time) (auth.Current, error) {
	if err := ctx.Err(); err != nil {
		return auth.Current{}, err
	}
	r.mu.RLock()
	session, exists := r.sessions[tokenHash]
	credentials := r.accounts[session.AccountID]
	r.mu.RUnlock()
	if !exists || !session.ExpiresAt.After(now) {
		return auth.Current{}, auth.ErrRequired
	}
	r.profiles.mu.RLock()
	stored := r.profiles.profiles[credentials.ProfileID]
	r.profiles.mu.RUnlock()
	credentials.Identity = stored.Identity
	return auth.Current{Account: credentials.Account, ExpiresAt: session.ExpiresAt}, nil
}

func (r *AuthRepository) DeleteSession(ctx context.Context, tokenHash string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.sessions, tokenHash)
	return nil
}
