package auth

import (
	"errors"
	"testing"
)

func TestNormalizeEmail(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		normalized string
		valid      bool
	}{
		{name: "обычная почта", input: " Dmitriy.Novikov+game@Example.COM ", normalized: "dmitriy.novikov+game@example.com", valid: true},
		{name: "кириллица", input: "дмитрий@example.com", valid: false},
		{name: "нет доменной зоны", input: "dmitriy@localhost", valid: false},
		{name: "двойная точка", input: "dmitriy..novikov@example.com", valid: false},
		{name: "дефис на краю домена", input: "dmitriy@-example.com", valid: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, normalized, err := NormalizeEmail(test.input)
			if test.valid {
				if err != nil || normalized != test.normalized {
					t.Fatalf("NormalizeEmail = %q, %v", normalized, err)
				}
				return
			}
			if !errors.Is(err, ErrInvalidEmail) {
				t.Fatalf("ожидалась ErrInvalidEmail, получено %v", err)
			}
		})
	}
}
