package l05productionserver

import (
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"
)

// Config: настройки сервиса из переменных окружения (12-factor).
type Config struct {
	Addr            string // ":" + PORT, например ":8080"
	LogLevel        slog.Level
	ShutdownTimeout time.Duration
	DatabaseURL     string
}

// LoadConfig читает настройки через getenv (в main это os.Getenv,
// в тестах подменяется картой).
//
//   - PORT: по умолчанию 8080; целое 1..65535. Addr = ":" + PORT.
//   - LOG_LEVEL: по умолчанию info; допустимы debug, info, warn, error
//     в любом регистре. Остальное: ошибка.
//   - SHUTDOWN_TIMEOUT: по умолчанию 10s; формат time.ParseDuration ("15s",
//     "1m"); должен быть больше нуля.
//   - DATABASE_URL: обязательна, пустая строка это ошибка.
//
// Пустая переменная значит «не задана». Ошибки не обрывают разбор:
// собери ВСЕ через errors.Join, чтобы человек починил конфиг за один заход.
// Текст каждой ошибки содержит имя переменной ("PORT: ...").
// При ошибке Config может быть любым.
func LoadConfig(getenv func(string) string) (Config, error) {
	get := func(key, def string) string {
		if v := getenv(key); v != "" {
			return v
		}
		return def
	}
	var cfg Config
	var errs []error

	port := get("PORT", "8080")
	if n, err := strconv.Atoi(port); err != nil || n < 1 || n > 65535 {
		errs = append(errs, fmt.Errorf("PORT: %q не порт 1..65535", port))
	}
	cfg.Addr = ":" + port

	switch lvl := strings.ToLower(get("LOG_LEVEL", "info")); lvl {
	case "debug":
		cfg.LogLevel = slog.LevelDebug
	case "info":
		cfg.LogLevel = slog.LevelInfo
	case "warn":
		cfg.LogLevel = slog.LevelWarn
	case "error":
		cfg.LogLevel = slog.LevelError
	default:
		errs = append(errs, fmt.Errorf("LOG_LEVEL: неизвестный уровень %q", lvl))
	}

	st := get("SHUTDOWN_TIMEOUT", "10s")
	d, err := time.ParseDuration(st)
	if err != nil || d <= 0 {
		errs = append(errs, fmt.Errorf("SHUTDOWN_TIMEOUT: %q не положительная длительность", st))
	}
	cfg.ShutdownTimeout = d

	cfg.DatabaseURL = getenv("DATABASE_URL")
	if cfg.DatabaseURL == "" {
		errs = append(errs, errors.New("DATABASE_URL: обязательна"))
	}
	return cfg, errors.Join(errs...)
}
