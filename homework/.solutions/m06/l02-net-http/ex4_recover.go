package l02nethttp

import "net/http"

// Recover возвращает middleware, который перехватывает панику в next.
//
//   - Паники нет: ответ next проходит как есть, log не вызывается.
//   - Паника с любым значением v: вызвать log(v) и ответить клиенту
//     500 с телом "internal server error\n" (ровно как http.Error).
//     Детали паники клиенту не отдаём: там бывают пути, SQL, секреты.
//   - Паника со значением http.ErrAbortHandler: это не ошибка, а штатный
//     способ хендлера оборвать ответ. Её не логировать и пробросить дальше
//     (panic(v) повторно), чтобы net/http тихо закрыл соединение.
func Recover(log func(v any)) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				v := recover()
				if v == nil {
					return
				}
				if v == http.ErrAbortHandler {
					panic(v)
				}
				log(v)
				http.Error(w, "internal server error", http.StatusInternalServerError)
			}()
			next.ServeHTTP(w, r)
		})
	}
}
