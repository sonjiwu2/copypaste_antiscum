// Package identifier генерирует идентификаторы доменных сущностей.
package identifier

import (
	"crypto/rand"
	"encoding/base32"
	"fmt"
	"strings"
	"sync/atomic"
)

// randomByteLength даёт 128 бит энтропии — достаточно, чтобы идентификатор
// попытки нельзя было подобрать перебором при отсутствии аутентификации.
const randomByteLength = 16

// Generator создаёт уникальные идентификаторы.
type Generator interface {
	NewID() (string, error)
}

// Random генерирует идентификаторы на основе crypto/rand.
type Random struct{}

// NewID возвращает случайный идентификатор в нижнем регистре без разделителей.
func (Random) NewID() (string, error) {
	buffer := make([]byte, randomByteLength)
	if _, err := rand.Read(buffer); err != nil {
		return "", fmt.Errorf("сгенерировать случайный идентификатор: %w", err)
	}

	encoded := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(buffer)

	return strings.ToLower(encoded), nil
}

// Sequential выдаёт предсказуемые идентификаторы вида "attempt-1" для тестов.
type Sequential struct {
	Prefix  string
	counter atomic.Uint64
}

// NewID возвращает следующий идентификатор последовательности.
func (s *Sequential) NewID() (string, error) {
	next := s.counter.Add(1)

	prefix := s.Prefix
	if prefix == "" {
		prefix = "id"
	}

	return fmt.Sprintf("%s-%d", prefix, next), nil
}
