package l04httpclient

import (
	"context"
	"net/http"
	"strconv"
	"time"
)

// Policy: настройки повторов.
type Policy struct {
	MaxAttempts int           // всего попыток, включая первую; меньше 1 считается как 1
	BaseDelay   time.Duration // пауза перед второй попыткой, дальше удваивается
	MaxDelay    time.Duration // потолок любой паузы
	// Sleep ждёт d или отмены ctx (тогда возвращает ctx.Err()).
	// В тестах подменяется, чтобы не ждать по-настоящему.
	Sleep func(ctx context.Context, d time.Duration) error
}

// Do выполняет req клиентом c и повторяет его при временных сбоях.
//
//   - Повторять можно только идемпотентный запрос: методы GET, HEAD, OPTIONS,
//     PUT, DELETE, либо любой метод с заголовком Idempotency-Key.
//     Иначе ровно одна попытка, что бы ни случилось.
//   - Повод для повтора: ошибка транспорта (кроме случая, когда отменён
//     контекст запроса: тогда сразу вернуть ошибку) или статус 429, 502,
//     503, 504. Остальные ответы (2xx, 4xx, 500) вернуть сразу.
//   - Пауза перед попыткой номер n (n >= 2): BaseDelay * 2^(n-2), но не больше
//     MaxDelay. Если у ответа 429 или 503 есть Retry-After с целым числом
//     секунд, пауза равна ему (и тоже не больше MaxDelay).
//   - Ждать через p.Sleep(req.Context(), пауза); ошибку Sleep вернуть.
//   - Тело промежуточного ответа закрыть перед повтором. Последний ответ
//     (удачный или нет, когда попытки кончились) вернуть открытым.
//   - Запрос с телом можно повторить, только если есть req.GetBody
//     (http.NewRequest заполняет его для strings.Reader, bytes.Reader,
//     bytes.Buffer). Тело есть, а GetBody нет: одна попытка.
//     Повторно отправляя тот же *http.Request, транспорт сам перемотает
//     тело через GetBody; если же собираешь новый запрос (req.Clone),
//     тело для него бери из req.GetBody() явно.
func Do(c *http.Client, req *http.Request, p Policy) (*http.Response, error) {
	attempts := max(p.MaxAttempts, 1)
	if !retryable(req) {
		attempts = 1
	}
	ctx := req.Context()
	delay := p.BaseDelay
	for n := 1; ; n++ {
		r := req
		if n > 1 && req.GetBody != nil {
			body, err := req.GetBody()
			if err != nil {
				return nil, err
			}
			r = req.Clone(ctx)
			r.Body = body
		}
		resp, err := c.Do(r)
		if err != nil && ctx.Err() != nil {
			return nil, err
		}
		if n == attempts || (err == nil && !retryStatus(resp.StatusCode)) {
			return resp, err
		}

		wait := min(delay, p.MaxDelay)
		delay *= 2
		if resp != nil {
			if s, err := strconv.Atoi(resp.Header.Get("Retry-After")); err == nil && s >= 0 &&
				(resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode == http.StatusServiceUnavailable) {
				wait = min(time.Duration(s)*time.Second, p.MaxDelay)
			}
			resp.Body.Close()
		}
		if err := p.Sleep(ctx, wait); err != nil {
			return nil, err
		}
	}
}

func retryable(req *http.Request) bool {
	if req.Body != nil && req.Body != http.NoBody && req.GetBody == nil {
		return false
	}
	switch req.Method {
	case http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodPut, http.MethodDelete:
		return true
	}
	return req.Header.Get("Idempotency-Key") != ""
}

func retryStatus(code int) bool {
	switch code {
	case http.StatusTooManyRequests, http.StatusBadGateway, http.StatusServiceUnavailable, http.StatusGatewayTimeout:
		return true
	}
	return false
}
