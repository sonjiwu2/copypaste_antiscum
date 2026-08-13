package auth

import (
	"errors"
	"fmt"
	"testing"
)

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		valid    bool
	}{
		{name: "латиница цифра и спецсимвол", password: "Antiscam1!", valid: true},
		{name: "кириллица", password: "Пароль123!", valid: false},
		{name: "без спецсимвола", password: "Antiscam1", valid: false},
		{name: "без цифры", password: "Antiscam!", valid: false},
		{name: "без буквы", password: "12345678!", valid: false},
		{name: "с пробелом", password: "Antiscam 1!", valid: false},
		{name: "слишком короткий", password: "Test1!", valid: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := ValidatePassword(test.password)
			if test.valid && err != nil {
				t.Fatalf("ValidatePassword: %v", err)
			}
			if !test.valid && !errors.Is(err, ErrWeakPassword) {
				t.Fatalf("ожидалась ErrWeakPassword, получено %v", err)
			}
		})
	}
}

func TestPasswordHasherRoundTrip(t *testing.T) {
	hasher := PasswordHasher{Iterations: 100}
	encoded, err := hasher.Hash("надежный-пароль")
	if err != nil {
		t.Fatalf("Hash: %v", err)
	}
	if encoded == "надежный-пароль" {
		t.Fatal("пароль сохранён открытым текстом")
	}
	if !hasher.Verify(encoded, "надежный-пароль") {
		t.Fatal("правильный пароль не прошёл проверку")
	}
	if hasher.Verify(encoded, "другой-пароль") {
		t.Fatal("неправильный пароль прошёл проверку")
	}
}

func TestPBKDF2SHA256KnownVector(t *testing.T) {
	actual := pbkdf2SHA256([]byte("password"), []byte("salt"), 2, 32)
	want := "ae4d0c95af6b46d32d0adff928f06dd02a303f8ef3c251dfd6e2d85a95474c43"
	if got := fmt.Sprintf("%x", actual); got != want {
		t.Fatalf("PBKDF2 = %s, ожидалось %s", got, want)
	}
}
