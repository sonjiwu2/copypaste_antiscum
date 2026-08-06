// Package profile содержит анонимный профиль пользователя.
//
// Профиль нужен, чтобы попытки и прогресс кому-то принадлежали. Регистрации,
// паролей и персональных данных в MVP нет: идентификатор выдаёт сервер, и
// пользователь остаётся анонимным.
package profile

import (
	"errors"
	"time"
)

// ID — идентификатор анонимного профиля.
//
// Значение выдаёт сервер и хранит в cookie. Оно является секретом доступа
// к истории пользователя, поэтому подбор должен быть невозможен.
type ID string

// Границы длины идентификатора.
//
// Верхняя граница защищает от мусора в базе: cookie может прийти от кого
// угодно, а колонка не должна принимать произвольно длинные строки.
const (
	MinIDLength = 16
	MaxIDLength = 64
)

// Доменные ошибки профиля.
var (
	// ErrEmptyID — идентификатор профиля обязателен.
	ErrEmptyID = errors.New("идентификатор профиля обязателен")

	// ErrInvalidID — идентификатор не мог быть выдан этим сервером.
	ErrInvalidID = errors.New("идентификатор профиля некорректен")
)

// Valid сообщает, мог ли такой идентификатор быть выдан сервером.
//
// Проверка намеренно строгая: непохожее значение из cookie заменяется новым
// профилем, а не создаёт запись в базе.
func (id ID) Valid() bool {
	if len(id) < MinIDLength || len(id) > MaxIDLength {
		return false
	}

	for _, symbol := range id {
		if !isIDSymbol(symbol) {
			return false
		}
	}

	return true
}

// isIDSymbol допускает алфавит base32 в нижнем регистре и разделители,
// которыми пользуются генераторы идентификаторов.
func isIDSymbol(symbol rune) bool {
	switch {
	case symbol >= 'a' && symbol <= 'z':
		return true
	case symbol >= '0' && symbol <= '9':
		return true
	case symbol == '-' || symbol == '_':
		return true
	default:
		return false
	}
}

// Profile — анонимный владелец попыток и прогресса.
type Profile struct {
	ID        ID
	CreatedAt time.Time
	UpdatedAt time.Time
}
