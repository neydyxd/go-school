package l05productionserver

import (
	"log/slog"
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
	panic("TODO: реализуй LoadConfig")
}
