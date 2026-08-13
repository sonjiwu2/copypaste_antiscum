// Package profile содержит анонимный профиль пользователя.
//
// Профиль нужен, чтобы попытки и прогресс кому-то принадлежали. Регистрации,
// паролей и персональных данных в MVP нет: идентификатор выдаёт сервер, и
// пользователь остаётся анонимным.
package profile

import (
	"errors"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
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

	ErrInvalidName = errors.New("имя игрока должно содержать от 2 до 24 символов")

	ErrInvalidAvatar = errors.New("аватар игрока некорректен")
)

type Avatar string

const (
	AvatarProfile Avatar = "profile"
	AvatarLeader1 Avatar = "leader-1"
	AvatarLeader2 Avatar = "leader-2"
	AvatarLeader3 Avatar = "leader-3"
)

func (avatar Avatar) Valid() bool {
	switch avatar {
	case AvatarProfile, AvatarLeader1, AvatarLeader2, AvatarLeader3:
		return true
	default:
		return false
	}
}

type Identity struct {
	DisplayName string
	Avatar      Avatar
}

func NormalizeIdentity(identity Identity) (Identity, error) {
	for _, symbol := range identity.DisplayName {
		if unicode.IsControl(symbol) {
			return Identity{}, ErrInvalidName
		}
	}
	identity.DisplayName = strings.Join(strings.Fields(identity.DisplayName), " ")
	length := utf8.RuneCountInString(identity.DisplayName)
	if length < 2 || length > 24 {
		return Identity{}, ErrInvalidName
	}
	if !identity.Avatar.Valid() {
		return Identity{}, ErrInvalidAvatar
	}
	return identity, nil
}

type Leader struct {
	Rank               int
	DisplayName        string
	Avatar             Avatar
	Rating             int
	CompletedScenarios int
	AverageScore       int
	CurrentPlayer      bool
}

type Leaderboard struct {
	Leaders []Leader
	Current *Leader
}

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
	Identity  Identity
	CreatedAt time.Time
	UpdatedAt time.Time
}
