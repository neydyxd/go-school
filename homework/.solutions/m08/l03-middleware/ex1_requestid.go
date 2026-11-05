package l03middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"regexp"

	"github.com/gin-gonic/gin"
)

// HeaderRequestID это заголовок с идентификатором запроса (и во входящем запросе, и в ответе).
const HeaderRequestID = "X-Request-ID"

type requestIDKey struct{}

var validID = regexp.MustCompile(`^[A-Za-z0-9_-]{1,64}$`)

// RequestID возвращает middleware, которое присваивает запросу идентификатор.
//
//   - Входящий X-Request-ID берётся как есть, если он допустимый: от 1 до 64 символов,
//     только латинские буквы, цифры, '-' и '_'. Так ID от балансировщика сквозной.
//   - Иначе генерируется новый: 16 случайных байт из crypto/rand в hex (32 символа).
//     Мусор из заголовка (пробелы, переводы строк, 1000 символов) в логи и ответ не пускаем.
//   - ID кладётся в контекст запроса (c.Request = c.Request.WithContext(...)) под ключом
//     приватного типа, чтобы его видели сервис и репозиторий, и в заголовок ответа.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader(HeaderRequestID)
		if !validID.MatchString(id) {
			b := make([]byte, 16)
			_, _ = rand.Read(b) // crypto/rand.Read не возвращает ошибку с Go 1.24
			id = hex.EncodeToString(b)
		}
		c.Request = c.Request.WithContext(context.WithValue(c.Request.Context(), requestIDKey{}, id))
		c.Header(HeaderRequestID, id)
		c.Next()
	}
}

// RequestIDFrom достаёт ID из контекста. Нет ID: пустая строка.
func RequestIDFrom(ctx context.Context) string {
	id, _ := ctx.Value(requestIDKey{}).(string)
	return id
}
