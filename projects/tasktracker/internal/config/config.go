// Package config читает настройки из окружения (12-factor) и валидирует их при старте.
package config

import (
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"strconv"
	"time"
)

type Config struct {
	Port            int
	DatabaseURL     string
	JWTSecret       string
	AccessTTL       time.Duration
	RefreshTTL      time.Duration
	LogLevel        slog.Level
	ShutdownTimeout time.Duration
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
	dur := func(key, def string) time.Duration {
		d, err := time.ParseDuration(str(key, def))
		if err != nil || d <= 0 {
			errs = append(errs, fmt.Errorf("%s: want положительную длительность вроде %s", key, def))
		}
		return d
	}

	p, err := strconv.Atoi(str("PORT", "8080"))
	if err != nil || p < 1 || p > 65535 {
		errs = append(errs, errors.New("PORT: want число 1-65535"))
	}
	c.Port = p

	c.DatabaseURL = getenv("DATABASE_URL")
	if u, err := url.Parse(c.DatabaseURL); err != nil || (u.Scheme != "postgres" && u.Scheme != "postgresql") {
		errs = append(errs, errors.New("DATABASE_URL: обязателен, формат postgres://user:pass@host:port/db"))
	}

	c.JWTSecret = getenv("JWT_SECRET")
	if len(c.JWTSecret) < 32 {
		errs = append(errs, errors.New("JWT_SECRET: обязателен, не короче 32 байт (openssl rand -base64 48)"))
	}

	c.AccessTTL = dur("ACCESS_TTL", "15m")
	c.RefreshTTL = dur("REFRESH_TTL", "720h")
	c.ShutdownTimeout = dur("SHUTDOWN_TIMEOUT", "10s")

	if err := c.LogLevel.UnmarshalText([]byte(str("LOG_LEVEL", "info"))); err != nil {
		errs = append(errs, errors.New("LOG_LEVEL: want debug, info, warn или error"))
	}

	if err := errors.Join(errs...); err != nil {
		return Config{}, err
	}
	return c, nil
}

// LogValue прячет секреты, когда конфиг попадает в лог.
func (c Config) LogValue() slog.Value {
	db := "[REDACTED]"
	if u, err := url.Parse(c.DatabaseURL); err == nil {
		db = u.Redacted()
	}
	return slog.GroupValue(
		slog.Int("port", c.Port),
		slog.String("database_url", db),
		slog.String("jwt_secret", "[REDACTED]"),
		slog.Duration("access_ttl", c.AccessTTL),
		slog.Duration("refresh_ttl", c.RefreshTTL),
		slog.String("log_level", c.LogLevel.String()),
	)
}
