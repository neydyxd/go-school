package l06configlogsdocs

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"regexp"
	"time"

	"github.com/gin-gonic/gin"
)

type loggerKey struct{}

// WithLogger кладёт логгер в контекст.
func WithLogger(ctx context.Context, l *slog.Logger) context.Context {
	return context.WithValue(ctx, loggerKey{}, l)
}

// FromContext достаёт логгер из контекста. Если его там нет, возвращает
// slog.Default(), но никогда не nil: вызывающий код не должен проверять.
func FromContext(ctx context.Context) *slog.Logger {
	if l, ok := ctx.Value(loggerKey{}).(*slog.Logger); ok {
		return l
	}
	return slog.Default()
}

var validRequestID = regexp.MustCompile(`^[A-Za-z0-9-]{1,64}$`)

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
	return func(c *gin.Context) {
		start := time.Now()
		id := c.GetHeader("X-Request-ID")
		if !validRequestID.MatchString(id) {
			b := make([]byte, 16)
			_, _ = rand.Read(b)
			id = hex.EncodeToString(b)
		}
		c.Header("X-Request-ID", id)
		l := base.With("request_id", id)
		c.Request = c.Request.WithContext(WithLogger(c.Request.Context(), l))

		c.Next()

		status := c.Writer.Status()
		level := slog.LevelInfo
		switch {
		case status >= 500:
			level = slog.LevelError
		case status >= 400:
			level = slog.LevelWarn
		}
		l.LogAttrs(c.Request.Context(), level, "http request",
			slog.String("method", c.Request.Method),
			slog.String("path", c.Request.URL.Path),
			slog.Int("status", status),
			slog.Int64("duration_ms", time.Since(start).Milliseconds()),
		)
	}
}
