package l06patterns

import (
	"context"
	"sync"
)

// Square: стадия конвейера. Читает числа из in, отправляет их квадраты в
// выходной канал. Выходной канал закрывается, когда закрыт in или отменён ctx.
// При отмене горутина стадии выходит, даже если её никто не читает и даже
// если in так и не закрыли.
func Square(ctx context.Context, in <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for {
			select {
			case n, ok := <-in:
				if !ok {
					return
				}
				select {
				case out <- n * n:
				case <-ctx.Done():
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}()
	return out
}

// Merge: fan-in. Сливает все каналы cs в один выходной.
//   - Значение из любого входа должно уходить на выход сразу, как появилось:
//     медленный первый вход не задерживает быстрый второй.
//   - Выход закрывается, когда закрыты все входы (без входов: сразу).
//   - При отмене ctx все горутины Merge выходят (даже если входы не закрыты),
//     выход закрывается.
func Merge(ctx context.Context, cs ...<-chan int) <-chan int {
	out := make(chan int)
	var wg sync.WaitGroup
	for _, c := range cs {
		wg.Go(func() {
			for {
				select {
				case v, ok := <-c:
					if !ok {
						return
					}
					select {
					case out <- v:
					case <-ctx.Done():
						return
					}
				case <-ctx.Done():
					return
				}
			}
		})
	}
	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}
