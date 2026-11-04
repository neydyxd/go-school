package l05context

import (
	"context"
	"errors"
	"sync"
)

// ErrNoFuncs возвращается, если First вызвали без функций.
var ErrNoFuncs = errors.New("l05context: нет функций")

// First запускает все fns конкурентно, каждой передаёт контекст, производный
// от ctx, и возвращает результат первой функции, завершившейся без ошибки.
//
//   - Как только есть успешный результат, контекст остальных отменяется.
//   - Если все функции вернули ошибку, вернуть "" и errors.Join всех ошибок
//     (порядок не важен). Отмена ctx снаружи сюда тоже попадает: функции
//     вернут ctx.Err(), и он окажется в общей ошибке.
//   - Пустой fns: вернуть "", ErrNoFuncs.
//   - First возвращается только после того, как все запущенные функции
//     завершились: после неё не остаётся ни одной горутины.
func First(ctx context.Context, fns ...func(context.Context) (string, error)) (string, error) {
	if len(fns) == 0 {
		return "", ErrNoFuncs
	}
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	type result struct {
		val string
		err error
	}
	results := make(chan result, len(fns))
	var wg sync.WaitGroup
	for _, fn := range fns {
		wg.Go(func() {
			v, err := fn(ctx)
			results <- result{v, err}
		})
	}

	var (
		winner string
		won    bool
		errs   []error
	)
	for range fns {
		r := <-results
		if r.err != nil {
			errs = append(errs, r.err)
			continue
		}
		if !won {
			winner, won = r.val, true
			cancel()
		}
	}
	wg.Wait()
	if won {
		return winner, nil
	}
	return "", errors.Join(errs...)
}
