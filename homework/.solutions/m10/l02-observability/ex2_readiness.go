package l02observability

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

// Check проверяет одну зависимость (база, Redis, соседний сервис).
// Должна уважать ctx, но код ниже не обязан ей доверять.
type Check func(ctx context.Context) error

// Readiness возвращает обработчик /readyz: сервис готов принимать трафик,
// только если готовы все зависимости.
//
//   - Все проверки запускаются параллельно, каждая получает контекст запроса
//     с таймаутом timeout.
//   - Обработчик отвечает не позже чем через timeout, даже если какая-то
//     проверка игнорирует ctx и висит: её результат считается ошибкой
//     "timeout", а сама горутина пусть доработает в фоне.
//   - Ответ: Content-Type application/json, тело
//     {"status":"ok","checks":{"db":"ok","redis":"ok"}} и код 200, если все
//     проверки вернули nil; иначе "status":"fail", код 503, а у упавших
//     проверок вместо "ok" текст ошибки (для зависшей: "timeout").
//   - Нет проверок: готов, 200 и "checks":{}.
func Readiness(checks map[string]Check, timeout time.Duration) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), timeout)
		defer cancel()

		type result struct {
			name string
			err  error
		}
		// Буфер на все проверки: опоздавшие горутины запишут и завершатся,
		// даже когда их уже никто не ждёт.
		results := make(chan result, len(checks))
		for name, check := range checks {
			go func() { results <- result{name, check(ctx)} }()
		}

		out := map[string]string{}
		for name := range checks {
			out[name] = "timeout"
		}
		ok := true
	collect:
		for range checks {
			select {
			case res := <-results:
				if res.err != nil {
					out[res.name] = res.err.Error()
					ok = false
				} else {
					out[res.name] = "ok"
				}
			case <-ctx.Done():
				ok = false
				break collect
			}
		}

		status, code := "ok", http.StatusOK
		if !ok {
			status, code = "fail", http.StatusServiceUnavailable
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(code)
		_ = json.NewEncoder(w).Encode(map[string]any{"status": status, "checks": out})
	})
}
