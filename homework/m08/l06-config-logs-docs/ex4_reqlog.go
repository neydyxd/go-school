package l06configlogsdocs

import (
	"context"
	"log/slog"

	"github.com/gin-gonic/gin"
)

type loggerKey struct{}

// WithLogger кладёт логгер в контекст.
func WithLogger(ctx context.Context, l *slog.Logger) context.Context {
	panic("TODO: реализуй WithLogger")
}

// FromContext достаёт логгер из контекста. Если его там нет, возвращает
// slog.Default(), но никогда не nil: вызывающий код не должен проверять.
func FromContext(ctx context.Context) *slog.Logger {
	panic("TODO: реализуй FromContext")
}

// RequestLogger это middleware логирования запросов.
//
//  1. request_id: значение заголовка X-Request-ID, если оно состоит только из
//     латиницы, цифр и дефиса и не длиннее 64 символов. Иначе (нет, мусор, перевод
//     строки) сгенерируй новый: 16 байт из crypto/rand в hex (32 символа).
//     Верни его клиенту заголовком ответа X-Request-ID.
//  2. Создай логгер base.With("request_id", id) и положи его в контекст запроса
//     (c.Request = c.Request.WithContext(WithLogger(...))), чтобы обработчики
//     писали логи через FromContext(c.Request.Context()) с тем же request_id.
//  3. После c.Next() запиши одну строку с сообщением "http request" и атрибутами
//     method, path, status, duration_ms (int64). path это c.Request.URL.Path, без query:
//     в query бывают токены. Заголовки (Authorization, Cookie) не логируй вообще.
//     Уровень: Error при status >= 500, Warn при 400-499, иначе Info.
func RequestLogger(base *slog.Logger) gin.HandlerFunc {
	panic("TODO: реализуй RequestLogger")
}
