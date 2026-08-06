package profile_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/platform/clock"
	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/platform/identifier"
	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/profile"
	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/storage/memory"
)

var moment = time.Date(2026, time.August, 3, 12, 0, 0, 0, time.UTC)

func newService(repository profile.Repository) *profile.Service {
	return profile.NewService(
		repository,
		&clock.Fixed{Moment: moment, Step: time.Minute},
		&identifier.Sequential{Prefix: "anonymous-profile"},
	)
}

func TestResolveIssuesProfileWhenAbsent(t *testing.T) {
	repository := memory.NewProfileRepository()
	service := newService(repository)

	resolved, err := service.Resolve(context.Background(), "")
	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}

	if !resolved.Issued {
		t.Error("профиль должен быть выдан как новый")
	}

	if resolved.ID == "" {
		t.Error("идентификатор профиля пуст")
	}

	if repository.Count() != 1 {
		t.Errorf("профилей в хранилище = %d, ожидался 1", repository.Count())
	}
}

// Возвращающийся пользователь сохраняет свой профиль: иначе история
// прохождений терялась бы при каждом запросе.
func TestResolveKeepsExistingProfile(t *testing.T) {
	repository := memory.NewProfileRepository()
	service := newService(repository)

	first, err := service.Resolve(context.Background(), "")
	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}

	second, err := service.Resolve(context.Background(), first.ID)
	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}

	if second.ID != first.ID {
		t.Errorf("профиль = %q, ожидался %q", second.ID, first.ID)
	}

	if second.Issued {
		t.Error("существующий профиль не должен выдаваться заново")
	}

	if repository.Count() != 1 {
		t.Errorf("профилей в хранилище = %d, ожидался 1", repository.Count())
	}
}

// Присланное клиентом значение не должно попадать в базу как есть:
// иначе профиль можно было бы выбрать себе самому.
func TestResolveRejectsForeignIdentifier(t *testing.T) {
	testCases := []struct {
		name      string
		presented profile.ID
	}{
		{name: "слишком короткий", presented: "short"},
		{name: "недопустимые символы", presented: "ПрофильСПробелами и знаками!"},
		{name: "слишком длинный", presented: profile.ID(longID(profile.MaxIDLength + 1))},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			repository := memory.NewProfileRepository()
			service := newService(repository)

			resolved, err := service.Resolve(context.Background(), testCase.presented)
			if err != nil {
				t.Fatalf("неожиданная ошибка: %v", err)
			}

			if !resolved.Issued {
				t.Error("некорректный идентификатор должен заменяться новым профилем")
			}

			if resolved.ID == testCase.presented {
				t.Error("сервер принял присланный идентификатор")
			}
		})
	}
}

func TestResolveRespectsCanceledContext(t *testing.T) {
	service := newService(memory.NewProfileRepository())

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if _, err := service.Resolve(ctx, ""); !errors.Is(err, context.Canceled) {
		t.Errorf("ошибка = %v, ожидалась context.Canceled", err)
	}
}

func TestIDValid(t *testing.T) {
	testCases := []struct {
		name string
		id   profile.ID
		want bool
	}{
		{name: "выданный сервером", id: "anonymous-profile-1", want: true},
		{name: "base32 из crypto/rand", id: "mfrggzdfmztwq2lknnwg23tp", want: true},
		{name: "пустой", id: "", want: false},
		{name: "короткий", id: "abc", want: false},
		{name: "с пробелом", id: "profile with space", want: false},
		{name: "в верхнем регистре", id: "ANONYMOUSPROFILE1234", want: false},
		{name: "длинный", id: profile.ID(longID(profile.MaxIDLength + 1)), want: false},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			if got := testCase.id.Valid(); got != testCase.want {
				t.Errorf("Valid() = %v, ожидалось %v", got, testCase.want)
			}
		})
	}
}

func longID(length int) string {
	value := make([]byte, length)
	for i := range value {
		value[i] = 'a'
	}

	return string(value)
}
