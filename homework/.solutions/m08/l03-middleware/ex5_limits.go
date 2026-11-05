package l03middleware

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// Формат тела ответа здесь упрощён ради краткости ({"error": "..."}). В настоящем сервисе
// ответы идут в едином формате курса из урока про binding: {"error": {"code", "message", "fields"}}.
//
// Timeout возвращает middleware, которое даёт запросу дедлайн d.
//
// Обработчик в отдельной горутине не запускай: gin.Context не потокобезопасен.
// Вместо этого положи в c.Request контекст с таймаутом (и обязательно отмени его
// после c.Next()), а обработчики и база сами следят за ctx.Done().
// Если после c.Next() дедлайн истёк, а ответ ещё не записан (c.Writer.Written()),
// ответь 504 {"error": "timeout"}.
func Timeout(d time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), d)
		defer cancel()
		c.Request = c.Request.WithContext(ctx)

		c.Next()

		if errors.Is(ctx.Err(), context.DeadlineExceeded) && !c.Writer.Written() {
			c.AbortWithStatusJSON(http.StatusGatewayTimeout, gin.H{"error": "timeout"})
		}
	}
}

// BodyLimit возвращает middleware, которое ограничивает тело запроса n байтами.
//
//   - Content-Length известен и больше n: сразу 413 {"error": "body too large"},
//     обработчик не вызывается.
//   - Иначе обернуть тело в http.MaxBytesReader: если длина неизвестна (chunked)
//     или клиент соврал, чтение тела в обработчике вернёт *http.MaxBytesError.
func BodyLimit(n int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.ContentLength > n {
			c.AbortWithStatusJSON(http.StatusRequestEntityTooLarge, gin.H{"error": "body too large"})
			return
		}
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, n)
		c.Next()
	}
}
