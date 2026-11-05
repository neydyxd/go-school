package l01docker

import (
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"
)

// Config: настройки сервиса. В контейнере они приходят из переменных окружения
// (environment в compose, env в Kubernetes), а не из файла внутри образа:
// один и тот же образ едет в staging и в прод с разными настройками.
type Config struct {
	Port            int
	DatabaseURL     string
	LogLevel        slog.Level
	ShutdownTimeout time.Duration
}

// LoadConfig читает настройки через getenv (в main это os.Getenv, в тестах
// подставная функция). Пустая строка считается «переменная не задана».
//
//   - PORT: целое 1..65535, по умолчанию 8080.
//   - DATABASE_URL: обязательна.
//   - LOG_LEVEL: debug, info, warn, error в любом регистре, по умолчанию info.
//     Ровно эти четыре слова: slog.Level.UnmarshalText принимает ещё
//     "warn+1" и "INFO-4", здесь это ошибка.
//   - SHUTDOWN_TIMEOUT: длительность в формате time.ParseDuration ("15s"),
//     строго больше нуля, по умолчанию 10s.
//
// Ошибки не обрываются на первой: верни их все разом (errors.Join), и в тексте
// каждой должно быть имя переменной. Тот, кто запускает контейнер, должен
// увидеть все проблемы с первого запуска, а не чинить их по одной.
// При ошибке Config может быть любым.
func LoadConfig(getenv func(string) string) (Config, error) {
	cfg := Config{Port: 8080, LogLevel: slog.LevelInfo, ShutdownTimeout: 10 * time.Second}
	var errs []error

	if v := getenv("PORT"); v != "" {
		p, err := strconv.Atoi(v)
		if err != nil || p < 1 || p > 65535 {
			errs = append(errs, fmt.Errorf("PORT=%q: нужно целое от 1 до 65535", v))
		}
		cfg.Port = p
	}

	cfg.DatabaseURL = getenv("DATABASE_URL")
	if cfg.DatabaseURL == "" {
		errs = append(errs, errors.New("DATABASE_URL: обязательная переменная не задана"))
	}

	if v := getenv("LOG_LEVEL"); v != "" {
		switch strings.ToLower(v) {
		case "debug":
			cfg.LogLevel = slog.LevelDebug
		case "info":
			cfg.LogLevel = slog.LevelInfo
		case "warn":
			cfg.LogLevel = slog.LevelWarn
		case "error":
			cfg.LogLevel = slog.LevelError
		default:
			errs = append(errs, fmt.Errorf("LOG_LEVEL=%q: допустимо debug, info, warn, error", v))
		}
	}

	if v := getenv("SHUTDOWN_TIMEOUT"); v != "" {
		d, err := time.ParseDuration(v)
		if err != nil || d <= 0 {
			errs = append(errs, fmt.Errorf("SHUTDOWN_TIMEOUT=%q: нужна положительная длительность, например 15s", v))
		}
		cfg.ShutdownTimeout = d
	}

	return cfg, errors.Join(errs...)
}
