// Package config загружает настройки приложения из переменных окружения.
package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

// StorageDriver — выбранное хранилище попыток.
type StorageDriver string

// Поддерживаемые хранилища.
const (
	// StoragePostgres — рабочее хранилище: попытки переживают перезапуск.
	StoragePostgres StorageDriver = "postgres"

	// StorageMemory — хранилище в памяти. Существует только для тестов и
	// демонстрации без базы, поэтому включается явно и никогда не выбирается
	// автоматически: тихий откат на память после сбоя базы означал бы
	// молчаливую потерю данных пользователя.
	StorageMemory StorageDriver = "memory"
)

// Valid сообщает, поддерживается ли хранилище.
func (d StorageDriver) Valid() bool {
	return d == StoragePostgres || d == StorageMemory
}

// SameSite — политика отправки cookie профиля в межсайтовых запросах.
type SameSite string

// Поддерживаемые политики.
const (
	// SameSiteLax — значение по умолчанию: фронтенд обслуживается тем же
	// сайтом, что и API (общий origin или общий registrable domain).
	SameSiteLax SameSite = "lax"

	// SameSiteStrict запрещает отправку cookie при любом внешнем переходе.
	SameSiteStrict SameSite = "strict"

	// SameSiteNone нужен только когда фронтенд живёт на чужом сайте и ходит
	// в API кросс-доменно. Браузер принимает такую cookie исключительно
	// вместе с Secure, поэтому конфигурация это требование проверяет.
	SameSiteNone SameSite = "none"
)

// Valid сообщает, поддерживается ли политика.
func (s SameSite) Valid() bool {
	return s == SameSiteLax || s == SameSiteStrict || s == SameSiteNone
}

// Config содержит настройки, необходимые для запуска HTTP-приложения.
type Config struct {
	HTTPAddr          string
	LogLevel          string
	ReadHeaderTimeout time.Duration
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
	ShutdownTimeout   time.Duration
	MaxRequestBytes   int64

	StorageDriver StorageDriver
	Database      Database
	Cookie        Cookie
	CORS          CORS
	Groq          Groq
	XAI           XAI
}

// Groq содержит серверные настройки бесплатного Groq Cloud API. Ключ
// остаётся только на backend и никогда не попадает в браузер.
type Groq struct {
	APIKey         string
	BaseURL        string
	Model          string
	RequestTimeout time.Duration
}

// XAI содержит серверные настройки Grok. Ключ никогда не передаётся
// фронтенду: браузер обращается только к нашему API.
type XAI struct {
	APIKey         string
	BaseURL        string
	Model          string
	RequestTimeout time.Duration
}

// Cookie описывает cookie анонимного профиля.
type Cookie struct {
	Name   string
	MaxAge time.Duration

	// Secure по умолчанию выключен: локальное демо работает по HTTP.
	// Публичное развёртывание обязано включить его вместе с HTTPS.
	Secure bool

	SameSite SameSite
}

// CORS описывает межсайтовый доступ фронтенда к API.
//
// Список origin пуст по умолчанию: браузерный доступ с чужого origin
// открывается только явной настройкой. Рекомендуемый путь — отдавать
// фронтенд и API с одного origin, тогда CORS не нужен вовсе.
type CORS struct {
	AllowedOrigins []string
}

// Database содержит настройки подключения к PostgreSQL.
//
// URL не имеет значения по умолчанию намеренно: тихий старт с чужой или
// локальной базой опаснее явной ошибки конфигурации.
type Database struct {
	URL            string
	MaxConns       int32
	MinConns       int32
	ConnectTimeout time.Duration
	QueryTimeout   time.Duration
	HealthTimeout  time.Duration
}

// Default возвращает конфигурацию со значениями по умолчанию.
func Default() Config {
	return Config{
		HTTPAddr:          ":8080",
		LogLevel:          "info",
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       60 * time.Second,
		ShutdownTimeout:   10 * time.Second,
		MaxRequestBytes:   64 * 1024,
		StorageDriver:     StoragePostgres,
		Database:          defaultDatabase(),
		Cookie:            defaultCookie(),
		Groq: Groq{
			BaseURL:        "https://api.groq.com/openai/v1",
			Model:          "openai/gpt-oss-20b",
			RequestTimeout: 40 * time.Second,
		},
		XAI: XAI{
			BaseURL:        "https://api.x.ai/v1",
			Model:          "grok-4.3",
			RequestTimeout: 40 * time.Second,
		},
	}
}

// defaultCookie задаёт долгий срок жизни: профиль анонимный, и потеря cookie
// означает потерю всей истории прохождений.
func defaultCookie() Cookie {
	return Cookie{
		Name:     "ast_profile",
		MaxAge:   180 * 24 * time.Hour,
		Secure:   false,
		SameSite: SameSiteLax,
	}
}

// defaultDatabase задаёт размеры пула и таймауты, но не адрес базы.
func defaultDatabase() Database {
	return Database{
		MaxConns:       10,
		MinConns:       0,
		ConnectTimeout: 5 * time.Second,
		QueryTimeout:   5 * time.Second,
		HealthTimeout:  2 * time.Second,
	}
}

// Load читает конфигурацию из окружения, дополняя её значениями по умолчанию.
func Load() (Config, error) {
	cfg := Default()

	if addr, ok := os.LookupEnv("HTTP_ADDR"); ok {
		cfg.HTTPAddr = addr
	}

	if level, ok := os.LookupEnv("LOG_LEVEL"); ok {
		cfg.LogLevel = level
	}

	durations := map[string]*time.Duration{
		"HTTP_READ_HEADER_TIMEOUT": &cfg.ReadHeaderTimeout,
		"HTTP_READ_TIMEOUT":        &cfg.ReadTimeout,
		"HTTP_WRITE_TIMEOUT":       &cfg.WriteTimeout,
		"HTTP_IDLE_TIMEOUT":        &cfg.IdleTimeout,
		"HTTP_SHUTDOWN_TIMEOUT":    &cfg.ShutdownTimeout,
		"DATABASE_CONNECT_TIMEOUT": &cfg.Database.ConnectTimeout,
		"DATABASE_QUERY_TIMEOUT":   &cfg.Database.QueryTimeout,
		"DATABASE_HEALTH_TIMEOUT":  &cfg.Database.HealthTimeout,
		"COOKIE_MAX_AGE":           &cfg.Cookie.MaxAge,
		"GROQ_REQUEST_TIMEOUT":     &cfg.Groq.RequestTimeout,
		"XAI_REQUEST_TIMEOUT":      &cfg.XAI.RequestTimeout,
	}

	for name, target := range durations {
		if err := readDuration(name, target); err != nil {
			return Config{}, err
		}
	}

	if err := readBytes("HTTP_MAX_REQUEST_BYTES", &cfg.MaxRequestBytes); err != nil {
		return Config{}, err
	}

	if name, ok := os.LookupEnv("COOKIE_NAME"); ok {
		cfg.Cookie.Name = name
	}

	if err := readBool("COOKIE_SECURE", &cfg.Cookie.Secure); err != nil {
		return Config{}, err
	}

	if sameSite, ok := os.LookupEnv("COOKIE_SAME_SITE"); ok {
		cfg.Cookie.SameSite = SameSite(strings.ToLower(strings.TrimSpace(sameSite)))
	}

	if origins, ok := os.LookupEnv("CORS_ALLOWED_ORIGINS"); ok {
		cfg.CORS.AllowedOrigins = splitList(origins)
	}

	if key, ok := os.LookupEnv("GROQ_API_KEY"); ok {
		cfg.Groq.APIKey = strings.TrimSpace(key)
	}
	if baseURL, ok := os.LookupEnv("GROQ_BASE_URL"); ok {
		cfg.Groq.BaseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	}
	if model, ok := os.LookupEnv("GROQ_MODEL"); ok {
		cfg.Groq.Model = strings.TrimSpace(model)
	}

	if key, ok := os.LookupEnv("XAI_API_KEY"); ok {
		cfg.XAI.APIKey = strings.TrimSpace(key)
	}
	if baseURL, ok := os.LookupEnv("XAI_BASE_URL"); ok {
		cfg.XAI.BaseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	}
	if model, ok := os.LookupEnv("XAI_MODEL"); ok {
		cfg.XAI.Model = strings.TrimSpace(model)
	}

	if driver, ok := os.LookupEnv("STORAGE_DRIVER"); ok {
		cfg.StorageDriver = StorageDriver(driver)
	}

	if url, ok := os.LookupEnv("DATABASE_URL"); ok {
		cfg.Database.URL = url
	}

	counts := map[string]*int32{
		"DATABASE_MAX_CONNS": &cfg.Database.MaxConns,
		"DATABASE_MIN_CONNS": &cfg.Database.MinConns,
	}

	for name, target := range counts {
		if err := readInt32(name, target); err != nil {
			return Config{}, err
		}
	}

	if err := cfg.validate(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func (c Config) validate() error {
	if c.HTTPAddr == "" {
		return fmt.Errorf("HTTP_ADDR не может быть пустым")
	}

	if c.MaxRequestBytes <= 0 {
		return fmt.Errorf("HTTP_MAX_REQUEST_BYTES должен быть положительным")
	}

	if !c.StorageDriver.Valid() {
		return fmt.Errorf("STORAGE_DRIVER = %q, поддерживаются %q и %q",
			c.StorageDriver, StoragePostgres, StorageMemory)
	}

	// Адрес базы обязателен именно в postgres-режиме: приложение не имеет
	// права подняться и начать терять попытки из-за пропущенной настройки.
	if c.StorageDriver == StoragePostgres && c.Database.URL == "" {
		return fmt.Errorf("DATABASE_URL обязателен при STORAGE_DRIVER=%q", StoragePostgres)
	}

	if err := c.Cookie.validate(); err != nil {
		return err
	}

	if err := c.CORS.validate(); err != nil {
		return err
	}

	if err := c.Groq.validate(); err != nil {
		return err
	}

	if err := c.XAI.validate(); err != nil {
		return err
	}

	return c.Database.validate()
}

func (g Groq) validate() error {
	if g.Model == "" {
		return fmt.Errorf("GROQ_MODEL не может быть пустым")
	}

	parsed, err := url.Parse(g.BaseURL)
	if err != nil {
		return fmt.Errorf("разобрать GROQ_BASE_URL: %w", err)
	}
	if parsed.Scheme == "" || parsed.Host == "" || parsed.RawQuery != "" || parsed.Fragment != "" {
		return fmt.Errorf("GROQ_BASE_URL должен иметь вид scheme://host[/path]")
	}

	return nil
}

func (x XAI) validate() error {
	if x.Model == "" {
		return fmt.Errorf("XAI_MODEL не может быть пустым")
	}

	parsed, err := url.Parse(x.BaseURL)
	if err != nil {
		return fmt.Errorf("разобрать XAI_BASE_URL: %w", err)
	}
	if parsed.Scheme == "" || parsed.Host == "" || parsed.RawQuery != "" || parsed.Fragment != "" {
		return fmt.Errorf("XAI_BASE_URL должен иметь вид scheme://host[/path]")
	}

	return nil
}

// validate проверяет настройки cookie профиля.
func (c Cookie) validate() error {
	if c.Name == "" {
		return fmt.Errorf("COOKIE_NAME не может быть пустым")
	}

	if !c.SameSite.Valid() {
		return fmt.Errorf("COOKIE_SAME_SITE = %q, поддерживаются %q, %q и %q",
			c.SameSite, SameSiteLax, SameSiteStrict, SameSiteNone)
	}

	// Браузер молча отбрасывает SameSite=None без Secure, и пользователь
	// терял бы профиль при каждом запросе. Ошибка конфигурации честнее.
	if c.SameSite == SameSiteNone && !c.Secure {
		return fmt.Errorf("COOKIE_SAME_SITE=%q требует COOKIE_SECURE=true", SameSiteNone)
	}

	return nil
}

// validate проверяет список разрешённых origin.
//
// Маска "*" запрещена намеренно: ответы API привязаны к cookie профиля, а
// credentialed-запрос с произвольного сайта означал бы чтение чужой истории
// любым сторонним ресурсом.
func (c CORS) validate() error {
	for _, origin := range c.AllowedOrigins {
		if origin == "*" {
			return fmt.Errorf("CORS_ALLOWED_ORIGINS не поддерживает %q: origin перечисляются явно", origin)
		}

		parsed, err := url.Parse(origin)
		if err != nil {
			return fmt.Errorf("разобрать CORS_ALLOWED_ORIGINS %q: %w", origin, err)
		}

		if parsed.Scheme == "" || parsed.Host == "" || parsed.Path != "" {
			return fmt.Errorf("CORS_ALLOWED_ORIGINS %q должен иметь вид scheme://host[:port]", origin)
		}
	}

	return nil
}

// splitList разбирает список значений, разделённых запятыми.
func splitList(raw string) []string {
	parts := strings.Split(raw, ",")
	values := make([]string, 0, len(parts))

	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed == "" {
			continue
		}

		values = append(values, trimmed)
	}

	return values
}

// validate проверяет настройки пула.
func (d Database) validate() error {
	if d.MaxConns <= 0 {
		return fmt.Errorf("DATABASE_MAX_CONNS должен быть положительным, получено %d", d.MaxConns)
	}

	if d.MinConns < 0 {
		return fmt.Errorf("DATABASE_MIN_CONNS не может быть отрицательным, получено %d", d.MinConns)
	}

	if d.MinConns > d.MaxConns {
		return fmt.Errorf("DATABASE_MIN_CONNS (%d) не может превышать DATABASE_MAX_CONNS (%d)",
			d.MinConns, d.MaxConns)
	}

	return nil
}

func readDuration(name string, target *time.Duration) error {
	raw, ok := os.LookupEnv(name)
	if !ok {
		return nil
	}

	parsed, err := time.ParseDuration(raw)
	if err != nil {
		return fmt.Errorf("разобрать %s: %w", name, err)
	}

	if parsed <= 0 {
		return fmt.Errorf("%s должен быть положительным, получено %q", name, raw)
	}

	*target = parsed

	return nil
}

func readBytes(name string, target *int64) error {
	raw, ok := os.LookupEnv(name)
	if !ok {
		return nil
	}

	parsed, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return fmt.Errorf("разобрать %s: %w", name, err)
	}

	*target = parsed

	return nil
}

func readBool(name string, target *bool) error {
	raw, ok := os.LookupEnv(name)
	if !ok {
		return nil
	}

	parsed, err := strconv.ParseBool(raw)
	if err != nil {
		return fmt.Errorf("разобрать %s: %w", name, err)
	}

	*target = parsed

	return nil
}

// readInt32 читает размер пула. Ширина 32 бита выбрана под pgxpool.Config,
// поэтому слишком большое значение отвергается здесь, а не молча обрезается.
func readInt32(name string, target *int32) error {
	raw, ok := os.LookupEnv(name)
	if !ok {
		return nil
	}

	parsed, err := strconv.ParseInt(raw, 10, 32)
	if err != nil {
		return fmt.Errorf("разобрать %s: %w", name, err)
	}

	*target = int32(parsed)

	return nil
}
