package l07problems

import (
	"context"
	"errors"
)

// ErrAllFailed: ни одна реплика не ответила успешно.
var ErrAllFailed = errors.New("все реплики вернули ошибку")

// Replica ищет query и возвращает ответ.
type Replica func(ctx context.Context, query string) (string, error)

// Search отправляет query во все реплики одновременно и возвращает первый
// успешный ответ (классика из доклада Роба Пайка про паттерны конкурентности).
//   - Возвращается СРАЗУ, как пришёл первый успех, не дожидаясь остальных.
//   - Остальным репликам отменяет контекст.
//   - Все ответили ошибкой (или реплик нет): вернуть ошибку, для которой
//     errors.Is(err, ErrAllFailed) и errors.Is(err, каждая ошибка реплики).
//   - Родительский ctx отменён раньше: сразу вернуть ctx.Err().
//   - Утечек нет: даже если реплика игнорирует отмену и ответит через час,
//     её горутина после этого должна завершиться, а не повиснуть навсегда.
func Search(ctx context.Context, query string, replicas ...Replica) (string, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	type result struct {
		val string
		err error
	}
	// Буфер на всех: опоздавшие реплики положат ответ и выйдут, даже если
	// читать его уже некому. Это и есть защита от утечки.
	results := make(chan result, len(replicas))
	for _, r := range replicas {
		go func() {
			v, err := r(ctx, query)
			results <- result{v, err}
		}()
	}

	errs := []error{ErrAllFailed}
	for range replicas {
		select {
		case r := <-results:
			if r.err == nil {
				return r.val, nil
			}
			errs = append(errs, r.err)
		case <-ctx.Done():
			return "", ctx.Err()
		}
	}
	return "", errors.Join(errs...)
}
