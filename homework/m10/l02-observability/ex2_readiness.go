package l02observability

import (
	"context"
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
	panic("TODO: реализуй Readiness")
}
