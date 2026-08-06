// Package config загружает настройки приложения из переменных окружения.
package config

import (
	"fmt"
	"os"
	"strconv"
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
}

// Cookie описывает cookie анонимного профиля.
type Cookie struct {
	Name   string
	MaxAge time.Duration

	// Secure по умолчанию выключен: локальное демо работает по HTTP.
	// Публичное развёртывание обязано включить его вместе с HTTPS.
	Secure bool
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
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
		ShutdownTimeout:   10 * time.Second,
		MaxRequestBytes:   64 * 1024,
		StorageDriver:     StoragePostgres,
		Database:          defaultDatabase(),
		Cookie:            defaultCookie(),
	}
}

// defaultCookie задаёт долгий срок жизни: профиль анонимный, и потеря cookie
// означает потерю всей истории прохождений.
func defaultCookie() Cookie {
	return Cookie{
		Name:   "ast_profile",
		MaxAge: 180 * 24 * time.Hour,
		Secure: false,
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

	if c.Cookie.Name == "" {
		return fmt.Errorf("COOKIE_NAME не может быть пустым")
	}

	return c.Database.validate()
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
