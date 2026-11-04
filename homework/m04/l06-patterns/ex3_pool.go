package l06patterns

// Pool: worker pool. Обрабатывает jobs функцией fn силами workers горутин и
// возвращает результаты в порядке jobs (results[i] = fn(jobs[i])).
//   - Одновременно выполняется не больше workers вызовов fn, и при достаточном
//     числе задач ровно workers.
//   - workers < 1 считать равным 1.
//   - Пустой jobs: пустой (не nil) слайс.
//   - Pool возвращается, когда все воркеры завершились.
func Pool[T, R any](jobs []T, workers int, fn func(T) R) []R {
	panic("TODO: реализуй Pool")
}
