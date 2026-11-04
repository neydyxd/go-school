// Package linkcheck: ядро конкурентного чекера ссылок (мини-проект модуля 04).
package linkcheck

import (
	"context"
	"net/http"
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
	panic("TODO: реализуй OK")
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
	panic("TODO: реализуй Check")
}
