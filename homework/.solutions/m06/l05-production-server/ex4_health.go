package l05productionserver

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"
)

// Check проверяет одну зависимость (БД, Redis, соседний сервис).
type Check func(ctx context.Context) error

// Readyz возвращает хендлер для /readyz: готов ли сервис принимать трафик.
//
//   - Все checks запускаются ПАРАЛЛЕЛЬНО (по горутине на проверку)
//     с общим контекстом context.WithTimeout(r.Context(), timeout).
//   - Хендлер отвечает не позже чем через timeout, даже если какая-то
//     проверка игнорирует контекст и висит дольше.
//   - Ответ JSON с Content-Type: application/json и Cache-Control: no-store:
//     {"status":"ok","checks":{"db":"ok","redis":"ok"}}             200
//     {"status":"fail","checks":{"db":"ok","redis":"timeout"}}       503
//     Значение проверки: "ok"; "timeout", если она не успела или вернула
//     ошибку с context.DeadlineExceeded; иначе err.Error().
//   - Без проверок: {"status":"ok","checks":{}} и 200.
//
// Горутины проверок не должны навсегда зависать на отправке результата
// после того, как хендлер ушёл по таймауту.
func Readyz(timeout time.Duration, checks map[string]Check) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), timeout)
		defer cancel()

		type result struct {
			name string
			err  error
		}
		results := make(chan result, len(checks)) // буфер: опоздавшие не зависнут
		for name, check := range checks {
			go func() { results <- result{name, check(ctx)} }()
		}

		status := map[string]string{}
		for name := range checks {
			status[name] = "timeout" // кто не успеет, так и останется
		}
		ok := true
	collect:
		for range checks {
			select {
			case res := <-results:
				switch {
				case res.err == nil:
					status[res.name] = "ok"
				case errors.Is(res.err, context.DeadlineExceeded):
					ok = false
				default:
					status[res.name] = res.err.Error()
					ok = false
				}
			case <-ctx.Done():
				ok = false
				break collect
			}
		}

		code, overall := http.StatusOK, "ok"
		if !ok {
			code, overall = http.StatusServiceUnavailable, "fail"
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
		w.WriteHeader(code)
		json.NewEncoder(w).Encode(map[string]any{"status": overall, "checks": status})
	})
}
