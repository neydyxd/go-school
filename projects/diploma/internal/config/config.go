// Package config читает настройки из окружения (12-factor) и валидирует их при старте.
// Добавляешь переменную для своей темы: поле в Config, разбор в Load, строка в LogValue,
// строка в .env.example и в docker-compose.yml.
package config

import (
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"strconv"
	"time"
)

// Config хранит все настройки сервиса. Источник один: переменные окружения.
type Config struct {
	Port            int
	DatabaseURL     string
	RedisURL        string // пусто: Redis не используется и не проверяется в /readyz
	LogLevel        slog.Level
	ShutdownTimeout time.Duration // сколько ждать активные запросы при остановке
	DrainDelay      time.Duration // пауза между «/readyz = 503» и Shutdown, чтобы балансировщик успел убрать под
}

// Load собирает конфиг из getenv (в проде os.Getenv) и возвращает все ошибки валидации сразу.
func Load(getenv func(string) string) (Config, error) {
	var (
		c    Config
		errs []error
	)
	str := func(key, def string) string {
		if v := getenv(key); v != "" {
			return v
		}
		return def
	}
	dur := func(key, def string, allowZero bool) time.Duration {
		d, err := time.ParseDuration(str(key, def))
		if err != nil || d < 0 || (d == 0 && !allowZero) {
			errs = append(errs, fmt.Errorf("%s: want длительность вроде %s", key, def))
		}
		return d
	}

	p, err := strconv.Atoi(str("PORT", "8080"))
	if err != nil || p < 1 || p > 65535 {
		errs = append(errs, errors.New("PORT: want число 1-65535"))
	}
	c.Port = p

	c.DatabaseURL = getenv("DATABASE_URL")
	if u, err := url.Parse(c.DatabaseURL); err != nil || (u.Scheme != "postgres" && u.Scheme != "postgresql") || u.Host == "" {
		errs = append(errs, errors.New("DATABASE_URL: обязателен, формат postgres://user:pass@host:port/db"))
	}

	c.RedisURL = getenv("REDIS_URL")
	if c.RedisURL != "" {
		if u, err := url.Parse(c.RedisURL); err != nil || (u.Scheme != "redis" && u.Scheme != "rediss") || u.Host == "" {
			errs = append(errs, errors.New("REDIS_URL: формат redis://host:port/0 или пусто"))
		}
	}

	if err := c.LogLevel.UnmarshalText([]byte(str("LOG_LEVEL", "info"))); err != nil {
		errs = append(errs, errors.New("LOG_LEVEL: want debug, info, warn или error"))
	}
	c.ShutdownTimeout = dur("SHUTDOWN_TIMEOUT", "10s", false) // 0 оборвал бы все активные запросы
	c.DrainDelay = dur("DRAIN_DELAY", "0s", true)

	if err := errors.Join(errs...); err != nil {
		return Config{}, err
	}
	return c, nil
}

// LogValue прячет пароли из строк подключения, когда конфиг попадает в лог.
func (c Config) LogValue() slog.Value {
	redact := func(raw string) string {
		if u, err := url.Parse(raw); err == nil {
			return u.Redacted()
		}
		return "[REDACTED]"
	}
	return slog.GroupValue(
		slog.Int("port", c.Port),
		slog.String("database_url", redact(c.DatabaseURL)),
		slog.String("redis_url", redact(c.RedisURL)),
		slog.String("log_level", c.LogLevel.String()),
		slog.String("shutdown_timeout", c.ShutdownTimeout.String()), // "10s", а не наносекунды
		slog.String("drain_delay", c.DrainDelay.String()),
	)
}
