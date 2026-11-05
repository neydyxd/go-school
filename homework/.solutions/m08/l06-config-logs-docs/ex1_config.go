package l06configlogsdocs

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// Config это настройки сервиса. Источники по убыванию приоритета: флаги, переменные окружения, значения по умолчанию.
type Config struct {
	Env             string        // APP_ENV: "dev" или "prod", по умолчанию "dev"
	Port            int           // PORT или флаг -port: 1-65535, по умолчанию 8080
	DatabaseURL     string        // DATABASE_URL: обязателен, схема postgres:// или postgresql://
	JWTSecret       string        // JWT_SECRET: обязателен, не короче 32 байт
	LogLevel        slog.Level    // LOG_LEVEL или флаг -log-level: debug, info, warn, error (регистр любой), по умолчанию info
	ShutdownTimeout time.Duration // SHUTDOWN_TIMEOUT: строка для time.ParseDuration, больше нуля, по умолчанию 10s
}

var levels = map[string]slog.Level{
	"debug": slog.LevelDebug, "info": slog.LevelInfo, "warn": slog.LevelWarn, "error": slog.LevelError,
}

// Load собирает Config из аргументов командной строки args (без имени программы)
// и окружения getenv (в проде os.Getenv, в тестах фейк).
//
// Флаги: -port и -log-level, разбирать через собственный flag.NewFlagSet с
// flag.ContinueOnError (глобальный flag.CommandLine трогать нельзя, вывод в io.Discard).
// Флаг, если задан, важнее переменной окружения. Неизвестный флаг: ошибка.
//
// Валидация при старте: проверь ВСЕ поля и верни все проблемы сразу через errors.Join,
// а не только первую. Текст каждой ошибки начинается с имени переменной (PORT, DATABASE_URL...).
// Значения секретов (JWT_SECRET, пароль из DATABASE_URL) в тексте ошибок появляться не должны.
func Load(args []string, getenv func(string) string) (Config, error) {
	fs := flag.NewFlagSet("api", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	portFlag := fs.String("port", "", "порт HTTP")
	levelFlag := fs.String("log-level", "", "уровень логов")
	if err := fs.Parse(args); err != nil {
		return Config{}, fmt.Errorf("флаги: %w", err)
	}

	pick := func(flagVal, env, def string) string {
		if flagVal != "" {
			return flagVal
		}
		if v := getenv(env); v != "" {
			return v
		}
		return def
	}

	var cfg Config
	var errs []error

	cfg.Env = pick("", "APP_ENV", "dev")
	if cfg.Env != "dev" && cfg.Env != "prod" {
		errs = append(errs, fmt.Errorf("APP_ENV: want dev или prod, got %q", cfg.Env))
	}

	port := pick(*portFlag, "PORT", "8080")
	p, err := strconv.Atoi(port)
	if err != nil || p < 1 || p > 65535 {
		errs = append(errs, fmt.Errorf("PORT: want число 1-65535, got %q", port))
	}
	cfg.Port = p

	cfg.DatabaseURL = getenv("DATABASE_URL")
	if cfg.DatabaseURL == "" {
		errs = append(errs, errors.New("DATABASE_URL: обязательная переменная не задана"))
	} else if u, err := url.Parse(cfg.DatabaseURL); err != nil || (u.Scheme != "postgres" && u.Scheme != "postgresql") || u.Host == "" {
		// сам адрес не печатаем: в нём пароль
		errs = append(errs, errors.New("DATABASE_URL: want postgres://user:pass@host:port/db"))
	}

	cfg.JWTSecret = getenv("JWT_SECRET")
	switch {
	case cfg.JWTSecret == "":
		errs = append(errs, errors.New("JWT_SECRET: обязательная переменная не задана"))
	case len(cfg.JWTSecret) < 32:
		errs = append(errs, fmt.Errorf("JWT_SECRET: want не короче 32 байт, got %d", len(cfg.JWTSecret)))
	}

	lvl := pick(*levelFlag, "LOG_LEVEL", "info")
	l, ok := levels[strings.ToLower(lvl)]
	if !ok {
		errs = append(errs, fmt.Errorf("LOG_LEVEL: want debug, info, warn или error, got %q", lvl))
	}
	cfg.LogLevel = l

	st := pick("", "SHUTDOWN_TIMEOUT", "10s")
	d, err := time.ParseDuration(st)
	if err != nil || d <= 0 {
		errs = append(errs, fmt.Errorf("SHUTDOWN_TIMEOUT: want положительную длительность вроде 10s, got %q", st))
	}
	cfg.ShutdownTimeout = d

	if err := errors.Join(errs...); err != nil {
		return Config{}, err
	}
	return cfg, nil
}
