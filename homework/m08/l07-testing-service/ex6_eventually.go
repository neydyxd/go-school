package l07testingservice

import "time"

// Eventually ждёт, пока cond() вернёт true, вместо time.Sleep «с запасом»,
// из-за которого тесты то медленные, то флакают.
//
//   - Первая проверка cond сразу, без ожидания; дальше раз в tick.
//   - cond вернул true: выйти без ошибок и больше cond не вызывать.
//   - Прошло timeout, а cond так и не стал true: ровно один t.Errorf с
//     сообщением, где есть timeout, и выход. После выхода cond не вызывается.
//
// cond вызывается из той же горутины, что и Eventually: так проще и без гонок.
func Eventually(t T, cond func() bool, timeout, tick time.Duration) {
	panic("TODO: реализуй Eventually")
}
