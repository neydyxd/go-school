package l02nethttp

import (
	"context"
	"net/http"
)

// Middleware оборачивает хендлер и возвращает новый.
type Middleware func(http.Handler) http.Handler

// Chain оборачивает h в middleware так, что первый в списке выполняется
// первым (он самый внешний): Chain(h, a, b, c) == a(b(c(h))).
// Без middleware возвращает h как есть.
func Chain(h http.Handler, mws ...Middleware) http.Handler {
	panic("TODO: реализуй Chain")
}

// RequestID возвращает middleware, который присваивает запросу ID.
//
//   - Если в запросе есть заголовок X-Request-Id и он валиден (1-64 символа,
//     только латинские буквы, цифры и дефис), используется он: так ID
//     сквозной через несколько сервисов.
//   - Иначе ID берётся из gen(). Невалидный входящий ID не доверяем:
//     это пользовательский ввод, он попадёт в логи.
//   - ID кладётся в контекст запроса (достаётся через RequestIDFrom)
//     и ставится в заголовок ответа X-Request-Id ДО вызова next.
//
// Ключ контекста делай своим неэкспортируемым типом, а не строкой.
func RequestID(gen func() string) Middleware {
	panic("TODO: реализуй RequestID")
}

// RequestIDFrom достаёт ID из контекста; если его там нет, возвращает "".
func RequestIDFrom(ctx context.Context) string {
	panic("TODO: реализуй RequestIDFrom")
}
