// Package config загружает настройки приложения из переменных окружения.
package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

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
	}

	for name, target := range durations {
		if err := readDuration(name, target); err != nil {
			return Config{}, err
		}
	}

	if err := readBytes("HTTP_MAX_REQUEST_BYTES", &cfg.MaxRequestBytes); err != nil {
		return Config{}, err
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
