package l03middleware

import (
	"context"

	"github.com/gin-gonic/gin"
)

// HeaderRequestID это заголовок с идентификатором запроса (и во входящем запросе, и в ответе).
const HeaderRequestID = "X-Request-ID"

// RequestID возвращает middleware, которое присваивает запросу идентификатор.
//
//   - Входящий X-Request-ID берётся как есть, если он допустимый: от 1 до 64 символов,
//     только латинские буквы, цифры, '-' и '_'. Так ID от балансировщика сквозной.
//   - Иначе генерируется новый: 16 случайных байт из crypto/rand в hex (32 символа).
//     Мусор из заголовка (пробелы, переводы строк, 1000 символов) в логи и ответ не пускаем.
//   - ID кладётся в контекст запроса (c.Request = c.Request.WithContext(...)) под ключом
//     приватного типа, чтобы его видели сервис и репозиторий, и в заголовок ответа.
func RequestID() gin.HandlerFunc {
	panic("TODO: реализуй RequestID")
}

// RequestIDFrom достаёт ID из контекста. Нет ID: пустая строка.
func RequestIDFrom(ctx context.Context) string {
	panic("TODO: реализуй RequestIDFrom")
}
