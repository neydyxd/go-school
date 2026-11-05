package l04httpclient

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
)

// First параллельно делает GET по всем urls (по горутине на url) и
// возвращает тело первого ответа со статусом 200.
//
//   - Как только есть победитель, остальные запросы отменяются
//     (общий context.WithCancel поверх ctx).
//   - First возвращается только после того, как ВСЕ запущенные горутины
//     завершились и все тела ответов закрыты: никаких утечек горутин.
//   - Ни один url не дал 200: вернуть ошибку (например, errors.Join
//     ошибок по каждому url). Пустой urls: тоже ошибка.
//   - ctx отменили снаружи раньше, чем нашёлся победитель: вернуть
//     ошибку, для которой errors.Is(err, ctx.Err()) истинно.
//
// Подумай про буфер канала результатов: если победитель уже найден,
// проигравшие горутины не должны навсегда зависнуть на отправке.
func First(ctx context.Context, c *http.Client, urls []string) (string, error) {
	if len(urls) == 0 {
		return "", errors.New("first: пустой список url")
	}
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	type result struct {
		body string
		err  error
	}
	results := make(chan result, len(urls)) // буфер: проигравшие не блокируются
	for _, u := range urls {
		go func() {
			body, err := get(ctx, c, u)
			results <- result{body, err}
		}()
	}

	var (
		winner string
		found  bool
		errs   []error
	)
	for range urls {
		r := <-results
		switch {
		case r.err == nil && !found:
			winner, found = r.body, true
			cancel()
		case r.err != nil && !found:
			errs = append(errs, r.err)
		}
	}
	if found {
		return winner, nil
	}
	return "", errors.Join(errs...)
}

func get(ctx context.Context, c *http.Client, url string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	resp, err := c.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("%s: status %d", url, resp.StatusCode)
	}
	b, err := io.ReadAll(resp.Body)
	return string(b), err
}
