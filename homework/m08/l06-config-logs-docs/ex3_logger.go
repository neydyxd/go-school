package l06configlogsdocs

import (
	"io"
	"log/slog"
)

// sensitiveKeys это ключи атрибутов, чьи значения нельзя писать в лог.
var sensitiveKeys = map[string]bool{
	"password": true, "token": true, "access_token": true, "refresh_token": true,
	"authorization": true, "cookie": true, "secret": true, "jwt_secret": true,
}

// NewLogger возвращает JSON-логгер в w с минимальным уровнем level.
//
// Через HandlerOptions.ReplaceAttr замени значение любого атрибута, ключ которого
// (без учёта регистра) есть в sensitiveKeys, на строку "[REDACTED]". Это должно
// работать везде: в атрибутах вызова, в logger.With(...) и внутри slog.Group.
// Остальные атрибуты, включая time, level и msg, не трогай.
func NewLogger(w io.Writer, level slog.Level) *slog.Logger {
	panic("TODO: реализуй NewLogger")
}
