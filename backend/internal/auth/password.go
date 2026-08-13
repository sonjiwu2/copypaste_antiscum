package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

const (
	defaultPBKDF2Iterations = 210_000
	passwordSaltBytes       = 16
	passwordKeyBytes        = 32
)

// PasswordHasher скрывает детали хранения пароля от сервиса и позволяет
// тестам использовать меньше итераций без ослабления production-настройки.
type PasswordHasher struct{ Iterations int }

func NewPasswordHasher() PasswordHasher {
	return PasswordHasher{Iterations: defaultPBKDF2Iterations}
}

func (h PasswordHasher) Hash(password string) (string, error) {
	iterations := h.Iterations
	if iterations <= 0 {
		iterations = defaultPBKDF2Iterations
	}
	salt := make([]byte, passwordSaltBytes)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("создать соль пароля: %w", err)
	}
	key := pbkdf2SHA256([]byte(password), salt, iterations, passwordKeyBytes)
	return fmt.Sprintf("pbkdf2-sha256$%d$%s$%s", iterations,
		base64.RawStdEncoding.EncodeToString(salt), base64.RawStdEncoding.EncodeToString(key)), nil
}

func (h PasswordHasher) Verify(encoded, password string) bool {
	parts := strings.Split(encoded, "$")
	if len(parts) != 4 || parts[0] != "pbkdf2-sha256" {
		return false
	}
	iterations, err := strconv.Atoi(parts[1])
	if err != nil || iterations < 1 || iterations > 1_000_000 {
		return false
	}
	salt, saltErr := base64.RawStdEncoding.DecodeString(parts[2])
	expected, keyErr := base64.RawStdEncoding.DecodeString(parts[3])
	if saltErr != nil || keyErr != nil || len(salt) < 8 || len(expected) != passwordKeyBytes {
		return false
	}
	actual := pbkdf2SHA256([]byte(password), salt, iterations, len(expected))
	return subtle.ConstantTimeCompare(actual, expected) == 1
}

func pbkdf2SHA256(password, salt []byte, iterations, keyLength int) []byte {
	if iterations < 1 || keyLength < 1 {
		panic(errors.New("некорректные параметры PBKDF2"))
	}
	blocks := (keyLength + sha256.Size - 1) / sha256.Size
	derived := make([]byte, 0, blocks*sha256.Size)
	for block := 1; block <= blocks; block++ {
		mac := hmac.New(sha256.New, password)
		_, _ = mac.Write(salt)
		_, _ = mac.Write([]byte{byte(block >> 24), byte(block >> 16), byte(block >> 8), byte(block)})
		u := mac.Sum(nil)
		result := append([]byte(nil), u...)
		for iteration := 1; iteration < iterations; iteration++ {
			mac = hmac.New(sha256.New, password)
			_, _ = mac.Write(u)
			u = mac.Sum(nil)
			for index := range result {
				result[index] ^= u[index]
			}
		}
		derived = append(derived, result...)
	}
	return derived[:keyLength]
}
