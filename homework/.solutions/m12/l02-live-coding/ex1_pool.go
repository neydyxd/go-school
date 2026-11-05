package l02livecoding

import (
	"context"
	"sync"
)

// RunPool: worker pool на каналах.
func RunPool[J, R any](ctx context.Context, jobs <-chan J, workers int, handle func(J) R) <-chan R {
	workers = max(workers, 1)
	out := make(chan R)
	var wg sync.WaitGroup
	for range workers {
		wg.Go(func() {
			for {
				var j J
				var ok bool
				select {
				case j, ok = <-jobs:
					if !ok {
						return
					}
				case <-ctx.Done():
					return
				}
				select {
				case out <- handle(j):
				case <-ctx.Done():
					return
				}
			}
		})
	}
	go func() {
		wg.Wait()
		close(out) // закрывает тот, кто знает, что писателей больше нет
	}()
	return out
}
