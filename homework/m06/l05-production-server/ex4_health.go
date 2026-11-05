package l05productionserver

import (
	"context"
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
	panic("TODO: реализуй Readyz")
}
