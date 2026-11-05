package l06configlogsdocs

import (
	"log/slog"
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
	panic("TODO: реализуй Load")
}
