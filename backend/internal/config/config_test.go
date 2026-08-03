package config_test

import (
	"testing"
	"time"

	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/config"
)

func TestLoad(t *testing.T) {
	testCases := []struct {
		name        string
		env         map[string]string
		wantErr     bool
		wantAddr    string
		wantTimeout time.Duration
	}{
		{
			name:        "значения по умолчанию без переменных окружения",
			env:         nil,
			wantAddr:    ":8080",
			wantTimeout: 15 * time.Second,
		},
		{
			name:        "переопределение адреса и таймаута",
			env:         map[string]string{"HTTP_ADDR": ":9090", "HTTP_READ_TIMEOUT": "30s"},
			wantAddr:    ":9090",
			wantTimeout: 30 * time.Second,
		},
		{
			name:    "нечитаемая длительность",
			env:     map[string]string{"HTTP_READ_TIMEOUT": "тридцать секунд"},
			wantErr: true,
		},
		{
			name:    "отрицательная длительность",
			env:     map[string]string{"HTTP_READ_TIMEOUT": "-1s"},
			wantErr: true,
		},
		{
			name:    "пустой адрес",
			env:     map[string]string{"HTTP_ADDR": ""},
			wantErr: true,
		},
		{
			name:    "неположительный лимит тела запроса",
			env:     map[string]string{"HTTP_MAX_REQUEST_BYTES": "0"},
			wantErr: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			for name, value := range testCase.env {
				t.Setenv(name, value)
			}

			cfg, err := config.Load()

			if testCase.wantErr {
				if err == nil {
					t.Fatalf("ожидалась ошибка, получена конфигурация %+v", cfg)
				}

				return
			}

			if err != nil {
				t.Fatalf("неожиданная ошибка: %v", err)
			}

			if cfg.HTTPAddr != testCase.wantAddr {
				t.Errorf("HTTPAddr = %q, ожидалось %q", cfg.HTTPAddr, testCase.wantAddr)
			}

			if cfg.ReadTimeout != testCase.wantTimeout {
				t.Errorf("ReadTimeout = %v, ожидалось %v", cfg.ReadTimeout, testCase.wantTimeout)
			}
		})
	}
}
