package l06configlogsdocs

import (
	"log/slog"
	"net/url"
)

// LogValue делает Config безопасным для логов (интерфейс slog.LogValuer):
// slog сам вызывает его, когда конфиг передают атрибутом, например logger.Info("start", "config", cfg).
//
// Верни группу (slog.GroupValue) с ключами env, port, database_url, jwt_secret,
// log_level, shutdown_timeout. В database_url пароль заменён (url.URL.Redacted
// подставляет "xxxxx"), jwt_secret всегда строка "[REDACTED]", а не сам секрет.
// Если DATABASE_URL не разбирается, тоже "[REDACTED]".
func (c Config) LogValue() slog.Value {
	db := "[REDACTED]"
	if u, err := url.Parse(c.DatabaseURL); err == nil {
		db = u.Redacted()
	}
	return slog.GroupValue(
		slog.String("env", c.Env),
		slog.Int("port", c.Port),
		slog.String("database_url", db),
		slog.String("jwt_secret", "[REDACTED]"),
		slog.String("log_level", c.LogLevel.String()),
		slog.Duration("shutdown_timeout", c.ShutdownTimeout),
	)
}

// String защищает от fmt.Println(cfg) и %v/%+v: без него fmt напечатает все поля, включая секрет.
// Верни c.LogValue().String().
func (c Config) String() string {
	return c.LogValue().String()
}
