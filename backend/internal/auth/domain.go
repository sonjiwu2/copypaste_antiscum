// Package auth содержит регистрацию, вход и серверные сессии игроков.
package auth

import (
	"errors"
	"strings"
	"time"

	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/profile"
)

type AccountID string

var (
	ErrRequired           = errors.New("требуется авторизация")
	ErrInvalidCredentials = errors.New("неверная электронная почта или пароль")
	ErrEmailTaken         = errors.New("электронная почта уже используется")
	ErrProfileClaimed     = errors.New("профиль уже привязан к аккаунту")
	ErrInvalidEmail       = errors.New("укажите корректную электронную почту")
	ErrWeakPassword       = errors.New("пароль должен содержать 8–72 латинских символа, букву, цифру и спецсимвол")
)

type Account struct {
	ID        AccountID
	ProfileID profile.ID
	Email     string
	Identity  profile.Identity
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Credentials struct {
	Account
	PasswordHash string
}

type Session struct {
	TokenHash string
	AccountID AccountID
	CreatedAt time.Time
	LastSeen  time.Time
	ExpiresAt time.Time
}

type Current struct {
	Account
	ExpiresAt time.Time
}

type RegisterInput struct {
	Email    string
	Password string
	Identity profile.Identity
}

func NormalizeEmail(value string) (string, string, error) {
	value = strings.TrimSpace(value)
	if len(value) < 5 || len(value) > 254 || strings.Count(value, "@") != 1 {
		return "", "", ErrInvalidEmail
	}
	local, domain, _ := strings.Cut(value, "@")
	if len(local) == 0 || len(local) > 64 || strings.HasPrefix(local, ".") ||
		strings.HasSuffix(local, ".") || strings.Contains(local, "..") {
		return "", "", ErrInvalidEmail
	}
	for _, symbol := range local {
		if !isEmailLocalSymbol(symbol) {
			return "", "", ErrInvalidEmail
		}
	}

	labels := strings.Split(domain, ".")
	if len(domain) > 253 || len(labels) < 2 || len(labels[len(labels)-1]) < 2 {
		return "", "", ErrInvalidEmail
	}
	for _, label := range labels {
		if len(label) == 0 || len(label) > 63 || label[0] == '-' || label[len(label)-1] == '-' {
			return "", "", ErrInvalidEmail
		}
		for _, symbol := range label {
			if !isASCIILetter(symbol) && !isASCIIDigit(symbol) && symbol != '-' {
				return "", "", ErrInvalidEmail
			}
		}
	}
	return value, strings.ToLower(value), nil
}

func isEmailLocalSymbol(symbol rune) bool {
	return isASCIILetter(symbol) || isASCIIDigit(symbol) ||
		strings.ContainsRune(".!#$%&'*+/=?^_`{|}~-", symbol)
}

func isASCIILetter(symbol rune) bool {
	return symbol >= 'a' && symbol <= 'z' || symbol >= 'A' && symbol <= 'Z'
}

func isASCIIDigit(symbol rune) bool { return symbol >= '0' && symbol <= '9' }

func ValidatePassword(password string) error {
	if len(password) < 8 || len(password) > 72 {
		return ErrWeakPassword
	}

	var hasLetter, hasDigit, hasSpecial bool
	for _, symbol := range password {
		switch {
		case symbol >= 'a' && symbol <= 'z', symbol >= 'A' && symbol <= 'Z':
			hasLetter = true
		case symbol >= '0' && symbol <= '9':
			hasDigit = true
		case symbol >= 33 && symbol <= 126:
			hasSpecial = true
		default:
			return ErrWeakPassword
		}
	}
	if !hasLetter || !hasDigit || !hasSpecial {
		return ErrWeakPassword
	}
	return nil
}
