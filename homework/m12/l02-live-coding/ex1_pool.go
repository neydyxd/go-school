package l02livecoding

import "context"

// RunPool: worker pool на каналах.
//   - Запускает workers воркеров (workers < 1 считать равным 1). Каждый читает задачи
//     из jobs и отправляет handle(job) в возвращаемый канал.
//   - Одновременно выполняется не больше workers вызовов handle.
//   - Канал результатов закрывается, когда jobs закрыт и все воркеры закончили.
//   - Порядок результатов не важен.
//   - При отмене ctx воркеры перестают брать задачи и выходят, даже если результат
//     никто не читает (не висят на отправке), после чего канал результатов закрывается.
func RunPool[J, R any](ctx context.Context, jobs <-chan J, workers int, handle func(J) R) <-chan R {
	panic("TODO: реализуй RunPool")
}
