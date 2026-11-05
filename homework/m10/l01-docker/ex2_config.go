package l01docker

import (
	"log/slog"
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
	panic("TODO: реализуй LoadConfig")
}
