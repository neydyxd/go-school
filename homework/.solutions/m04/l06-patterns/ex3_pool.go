package l06patterns

import "sync"

// Pool: worker pool. Обрабатывает jobs функцией fn силами workers горутин и
// возвращает результаты в порядке jobs (results[i] = fn(jobs[i])).
//   - Одновременно выполняется не больше workers вызовов fn, и при достаточном
//     числе задач ровно workers.
//   - workers < 1 считать равным 1.
//   - Пустой jobs: пустой (не nil) слайс.
//   - Pool возвращается, когда все воркеры завершились.
func Pool[T, R any](jobs []T, workers int, fn func(T) R) []R {
	workers = max(workers, 1)
	results := make([]R, len(jobs))
	idx := make(chan int)
	var wg sync.WaitGroup
	for range workers {
		wg.Go(func() {
			for i := range idx {
				results[i] = fn(jobs[i])
			}
		})
	}
	for i := range jobs {
		idx <- i
	}
	close(idx)
	wg.Wait()
	return results
}
