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
	for i := len(mws) - 1; i >= 0; i-- {
		h = mws[i](h)
	}
	return h
}

type requestIDKey struct{}

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
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id := r.Header.Get("X-Request-Id")
			if !validID(id) {
				id = gen()
			}
			w.Header().Set("X-Request-Id", id)
			ctx := context.WithValue(r.Context(), requestIDKey{}, id)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func validID(s string) bool {
	if len(s) == 0 || len(s) > 64 {
		return false
	}
	for _, c := range []byte(s) {
		ok := c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c == '-'
		if !ok {
			return false
		}
	}
	return true
}

// RequestIDFrom достаёт ID из контекста; если его там нет, возвращает "".
func RequestIDFrom(ctx context.Context) string {
	id, _ := ctx.Value(requestIDKey{}).(string)
	return id
}
