package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/platform/clock"
	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/platform/identifier"
	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/profile"
)

type Repository interface {
	Register(ctx context.Context, account Account, emailNormalized, passwordHash string) error
	CredentialsByEmail(ctx context.Context, emailNormalized string) (Credentials, error)
	CreateSession(ctx context.Context, session Session) error
	CurrentBySession(ctx context.Context, tokenHash string, now time.Time) (Current, error)
	DeleteSession(ctx context.Context, tokenHash string) error
}

type Service struct {
	repository Repository
	profiles   *profile.Service
	clock      clock.Clock
	ids        identifier.Generator
	hasher     PasswordHasher
	lifetime   time.Duration
}

func NewService(repository Repository, profiles *profile.Service, systemClock clock.Clock,
	ids identifier.Generator, hasher PasswordHasher, lifetime time.Duration) *Service {
	return &Service{repository: repository, profiles: profiles, clock: systemClock,
		ids: ids, hasher: hasher, lifetime: lifetime}
}

func (s *Service) Register(ctx context.Context, profileID profile.ID, input RegisterInput) (Current, string, error) {
	email, normalized, err := NormalizeEmail(input.Email)
	if err != nil {
		return Current{}, "", err
	}
	if err := ValidatePassword(input.Password); err != nil {
		return Current{}, "", err
	}
	identity, err := profile.NormalizeIdentity(input.Identity)
	if err != nil {
		return Current{}, "", err
	}
	accountID, err := s.ids.NewID()
	if err != nil {
		return Current{}, "", fmt.Errorf("создать аккаунт: %w", err)
	}
	passwordHash, err := s.hasher.Hash(input.Password)
	if err != nil {
		return Current{}, "", err
	}
	now := s.clock.Now()
	account := Account{ID: AccountID(accountID), ProfileID: profileID, Email: email,
		Identity: identity, CreatedAt: now, UpdatedAt: now}
	if err := s.repository.Register(ctx, account, normalized, passwordHash); err != nil {
		return Current{}, "", fmt.Errorf("зарегистрировать аккаунт: %w", err)
	}
	return s.newSession(ctx, account)
}

func (s *Service) Login(ctx context.Context, email, password string) (Current, string, error) {
	_, normalized, err := NormalizeEmail(email)
	if err != nil || ValidatePassword(password) != nil {
		return Current{}, "", ErrInvalidCredentials
	}
	credentials, err := s.repository.CredentialsByEmail(ctx, normalized)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			return Current{}, "", ErrInvalidCredentials
		}
		return Current{}, "", fmt.Errorf("найти аккаунт: %w", err)
	}
	if !s.hasher.Verify(credentials.PasswordHash, password) {
		return Current{}, "", ErrInvalidCredentials
	}
	return s.newSession(ctx, credentials.Account)
}

func (s *Service) newSession(ctx context.Context, account Account) (Current, string, error) {
	rawToken, err := s.ids.NewID()
	if err != nil {
		return Current{}, "", fmt.Errorf("создать сессию: %w", err)
	}
	now := s.clock.Now()
	expires := now.Add(s.lifetime)
	if err := s.repository.CreateSession(ctx, Session{TokenHash: HashToken(rawToken),
		AccountID: account.ID, CreatedAt: now, LastSeen: now, ExpiresAt: expires}); err != nil {
		return Current{}, "", fmt.Errorf("сохранить сессию: %w", err)
	}
	return Current{Account: account, ExpiresAt: expires}, rawToken, nil
}

func (s *Service) Resolve(ctx context.Context, rawToken string) (Current, error) {
	if rawToken == "" {
		return Current{}, ErrRequired
	}
	current, err := s.repository.CurrentBySession(ctx, HashToken(rawToken), s.clock.Now())
	if err != nil {
		if errors.Is(err, ErrRequired) {
			return Current{}, ErrRequired
		}
		return Current{}, fmt.Errorf("прочитать сессию: %w", err)
	}
	return current, nil
}

func (s *Service) Logout(ctx context.Context, rawToken string) error {
	if rawToken == "" {
		return nil
	}
	if err := s.repository.DeleteSession(ctx, HashToken(rawToken)); err != nil {
		return fmt.Errorf("завершить сессию: %w", err)
	}
	return nil
}

func HashToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}
