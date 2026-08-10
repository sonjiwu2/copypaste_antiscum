package config_test

import (
	"slices"
	"testing"
	"time"

	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/config"
)

// baselineDatabaseURL — фиктивный адрес: соединение в тестах не открывается.
const baselineDatabaseURL = "postgres://user:secret@127.0.0.1:5432/antiscam?sslmode=disable"

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
			// Адрес базы обязателен в рабочем режиме хранилища, поэтому он
			// задаётся всем случаям: здесь проверяются настройки HTTP.
			t.Setenv("DATABASE_URL", baselineDatabaseURL)

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

func TestLoadGroq(t *testing.T) {
	t.Setenv("DATABASE_URL", baselineDatabaseURL)
	t.Setenv("GROQ_API_KEY", " test-groq-key ")
	t.Setenv("GROQ_MODEL", "openai/gpt-oss-20b")
	t.Setenv("GROQ_BASE_URL", "https://api.groq.com/openai/v1/")
	t.Setenv("GROQ_REQUEST_TIMEOUT", "25s")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}

	if cfg.Groq.APIKey != "test-groq-key" {
		t.Errorf("Groq APIKey не был очищен от пробелов")
	}
	if cfg.Groq.Model != "openai/gpt-oss-20b" {
		t.Errorf("Groq Model = %q", cfg.Groq.Model)
	}
	if cfg.Groq.BaseURL != "https://api.groq.com/openai/v1" {
		t.Errorf("Groq BaseURL = %q", cfg.Groq.BaseURL)
	}
	if cfg.Groq.RequestTimeout != 25*time.Second {
		t.Errorf("Groq RequestTimeout = %v", cfg.Groq.RequestTimeout)
	}
}

// Настройки базы данных проверяются отдельно: у них свои правила
// согласованности, которых нет у HTTP-таймаутов.
func TestLoadDatabase(t *testing.T) {
	testCases := []struct {
		name         string
		env          map[string]string
		wantErr      bool
		wantURL      string
		wantMaxConns int32
		wantMinConns int32
		wantQuery    time.Duration
	}{
		{
			name:         "значения по умолчанию",
			env:          nil,
			wantURL:      baselineDatabaseURL,
			wantMaxConns: 10,
			wantMinConns: 0,
			wantQuery:    5 * time.Second,
		},
		{
			name: "полное переопределение",
			env: map[string]string{
				"DATABASE_URL":           "postgres://user:secret@postgres:5432/antiscam?sslmode=disable",
				"DATABASE_MAX_CONNS":     "20",
				"DATABASE_MIN_CONNS":     "2",
				"DATABASE_QUERY_TIMEOUT": "3s",
			},
			wantURL:      "postgres://user:secret@postgres:5432/antiscam?sslmode=disable",
			wantMaxConns: 20,
			wantMinConns: 2,
			wantQuery:    3 * time.Second,
		},
		{
			name:    "нечисловой размер пула",
			env:     map[string]string{"DATABASE_MAX_CONNS": "много"},
			wantErr: true,
		},
		{
			name:    "нулевой максимум подключений",
			env:     map[string]string{"DATABASE_MAX_CONNS": "0"},
			wantErr: true,
		},
		{
			name:    "отрицательный минимум подключений",
			env:     map[string]string{"DATABASE_MIN_CONNS": "-1"},
			wantErr: true,
		},
		{
			// Минимум больше максимума не даст пулу подняться, поэтому
			// ошибка нужна на старте, а не при первом запросе.
			name:    "минимум превышает максимум",
			env:     map[string]string{"DATABASE_MAX_CONNS": "2", "DATABASE_MIN_CONNS": "5"},
			wantErr: true,
		},
		{
			name:    "неположительный таймаут подключения",
			env:     map[string]string{"DATABASE_CONNECT_TIMEOUT": "0s"},
			wantErr: true,
		},
		{
			name:    "нечитаемый таймаут запроса",
			env:     map[string]string{"DATABASE_QUERY_TIMEOUT": "мгновенно"},
			wantErr: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Setenv("DATABASE_URL", baselineDatabaseURL)

			for name, value := range testCase.env {
				t.Setenv(name, value)
			}

			cfg, err := config.Load()

			if testCase.wantErr {
				if err == nil {
					t.Fatalf("ожидалась ошибка, получена конфигурация %+v", cfg.Database)
				}

				return
			}

			if err != nil {
				t.Fatalf("неожиданная ошибка: %v", err)
			}

			if cfg.Database.URL != testCase.wantURL {
				t.Errorf("URL = %q, ожидался %q", cfg.Database.URL, testCase.wantURL)
			}

			if cfg.Database.MaxConns != testCase.wantMaxConns {
				t.Errorf("MaxConns = %d, ожидалось %d", cfg.Database.MaxConns, testCase.wantMaxConns)
			}

			if cfg.Database.MinConns != testCase.wantMinConns {
				t.Errorf("MinConns = %d, ожидалось %d", cfg.Database.MinConns, testCase.wantMinConns)
			}

			if cfg.Database.QueryTimeout != testCase.wantQuery {
				t.Errorf("QueryTimeout = %v, ожидалось %v", cfg.Database.QueryTimeout, testCase.wantQuery)
			}
		})
	}
}

// Выбор хранилища определяет, переживают ли попытки перезапуск,
// поэтому режим работы с памятью включается только явно.
func TestLoadStorageDriver(t *testing.T) {
	testCases := []struct {
		name       string
		env        map[string]string
		wantErr    bool
		wantDriver config.StorageDriver
	}{
		{
			name:       "по умолчанию рабочее хранилище",
			env:        map[string]string{"DATABASE_URL": baselineDatabaseURL},
			wantDriver: config.StoragePostgres,
		},
		{
			name:       "память включается явно и не требует базы",
			env:        map[string]string{"STORAGE_DRIVER": "memory"},
			wantDriver: config.StorageMemory,
		},
		{
			// Пропущенный адрес базы не должен приводить к тихому запуску
			// без сохранения попыток.
			name:    "postgres без адреса базы",
			env:     map[string]string{"STORAGE_DRIVER": "postgres"},
			wantErr: true,
		},
		{
			name:    "неизвестное хранилище",
			env:     map[string]string{"STORAGE_DRIVER": "sqlite"},
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
					t.Fatalf("ожидалась ошибка, получен драйвер %q", cfg.StorageDriver)
				}

				return
			}

			if err != nil {
				t.Fatalf("неожиданная ошибка: %v", err)
			}

			if cfg.StorageDriver != testCase.wantDriver {
				t.Errorf("драйвер = %q, ожидался %q", cfg.StorageDriver, testCase.wantDriver)
			}
		})
	}
}

// Cookie профиля и межсайтовый доступ проверяются вместе: браузер принимает
// SameSite=None только вместе с Secure, а маска origin с credentials
// запрещена.
func TestLoadCookieAndCORS(t *testing.T) {
	testCases := []struct {
		name         string
		env          map[string]string
		wantErr      bool
		wantSameSite config.SameSite
		wantOrigins  []string
	}{
		{
			name:         "по умолчанию lax и без кросс-доменного доступа",
			env:          nil,
			wantSameSite: config.SameSiteLax,
			wantOrigins:  nil,
		},
		{
			name: "кросс-доменный фронтенд по HTTPS",
			env: map[string]string{
				"COOKIE_SAME_SITE":     "none",
				"COOKIE_SECURE":        "true",
				"CORS_ALLOWED_ORIGINS": "https://trainer.example, https://www.trainer.example",
			},
			wantSameSite: config.SameSiteNone,
			wantOrigins:  []string{"https://trainer.example", "https://www.trainer.example"},
		},
		{
			name:    "SameSite=None без Secure браузер отбросит",
			env:     map[string]string{"COOKIE_SAME_SITE": "none"},
			wantErr: true,
		},
		{
			name:    "неизвестная политика SameSite",
			env:     map[string]string{"COOKIE_SAME_SITE": "sometimes"},
			wantErr: true,
		},
		{
			name:    "маска origin с cookie профиля запрещена",
			env:     map[string]string{"CORS_ALLOWED_ORIGINS": "*"},
			wantErr: true,
		},
		{
			name:    "origin с путём вместо scheme://host",
			env:     map[string]string{"CORS_ALLOWED_ORIGINS": "https://trainer.example/app"},
			wantErr: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Setenv("DATABASE_URL", baselineDatabaseURL)

			for name, value := range testCase.env {
				t.Setenv(name, value)
			}

			cfg, err := config.Load()

			if testCase.wantErr {
				if err == nil {
					t.Fatalf("ожидалась ошибка, получена конфигурация %+v", cfg.Cookie)
				}

				return
			}

			if err != nil {
				t.Fatalf("неожиданная ошибка: %v", err)
			}

			if cfg.Cookie.SameSite != testCase.wantSameSite {
				t.Errorf("SameSite = %q, ожидался %q", cfg.Cookie.SameSite, testCase.wantSameSite)
			}

			if !slices.Equal(cfg.CORS.AllowedOrigins, testCase.wantOrigins) {
				t.Errorf("origin = %v, ожидались %v", cfg.CORS.AllowedOrigins, testCase.wantOrigins)
			}
		})
	}
}
