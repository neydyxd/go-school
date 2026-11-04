// Package linkcheck: ядро конкурентного чекера ссылок (мини-проект модуля 04).
package linkcheck

import (
	"context"
	"io"
	"net/http"
	"sync"
	"time"
)

// Result: итог проверки одной ссылки.
type Result struct {
	URL    string
	Status int   // код HTTP-ответа; 0, если ответа не было
	Err    error // ошибка запроса или чтения тела; nil, если ответ прочитан целиком
}

// OK сообщает, жива ли ссылка: ошибки нет и код 2xx или 3xx.
func (r Result) OK() bool {
	return r.Err == nil && r.Status >= 200 && r.Status < 400
}

// Options: настройки проверки.
type Options struct {
	Concurrency int           // максимум одновременных запросов; < 1 значит 1
	Timeout     time.Duration // таймаут на одну ссылку, включая чтение тела; 0: без таймаута
	Client      *http.Client  // nil: http.DefaultClient
}

// Check проверяет urls GET-запросами и возвращает результаты в том же
// порядке: results[i] относится к urls[i], len(results) == len(urls).
//
//   - Одновременно идёт не больше opts.Concurrency запросов.
//   - Каждая ссылка получает свой таймаут opts.Timeout (context.WithTimeout).
//     Он покрывает и чтение тела: тело нужно дочитать (io.Copy в io.Discard)
//     и закрыть, тогда соединение вернётся в пул. Если тело не дочиталось,
//     Status = код ответа, Err = ошибка чтения.
//   - Кривой URL (http.NewRequestWithContext вернул ошибку): Status 0, Err.
//   - Отмена ctx: идущие запросы прерываются, новые не начинаются; у всех
//     непроверенных ссылок Err такой, что errors.Is(Err, context.Canceled)
//     (или DeadlineExceeded, если у ctx истёк дедлайн). Check возвращается
//     сразу, а не после всех ссылок.
//   - Check возвращается только после завершения всех своих горутин.
func Check(ctx context.Context, urls []string, opts Options) []Result {
	client := opts.Client
	if client == nil {
		client = http.DefaultClient
	}
	results := make([]Result, len(urls))
	sem := make(chan struct{}, max(opts.Concurrency, 1))
	var wg sync.WaitGroup
	for i, u := range urls {
		results[i].URL = u
		select {
		case sem <- struct{}{}:
		case <-ctx.Done():
			results[i].Err = ctx.Err()
			continue
		}
		wg.Go(func() {
			defer func() { <-sem }()
			results[i].Status, results[i].Err = checkOne(ctx, client, u, opts.Timeout)
		})
	}
	wg.Wait()
	return results
}

func checkOne(ctx context.Context, client *http.Client, url string, timeout time.Duration) (int, error) {
	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return 0, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	_, err = io.Copy(io.Discard, resp.Body)
	return resp.StatusCode, err
}
