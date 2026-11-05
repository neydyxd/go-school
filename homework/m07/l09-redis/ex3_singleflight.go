package l09redis

import "context"

// GetOnce это Get, защищённый от cache stampede (эффект толпы): когда
// популярный ключ протух, сотни запросов одновременно промахиваются
// и все идут в базу. С GetOnce в базу идёт ровно один, остальные ждут его результат.
//
//  1. GET ключа: есть в кеше, вернуть.
//  2. Промах: c.group.Do(ключ, ...) из golang.org/x/sync/singleflight.
//     Внутри Do СНОВА проверь кеш, прежде чем звать c.load: горутина могла
//     промахнуться, пока предыдущий «полёт» ещё шёл, а войти в Do уже после
//     того, как он положил значение в Redis. Проще всего внутри Do вызвать c.Get.
//
// Результат Do имеет тип any, приведи его обратно к Product.
func (c *ProductCache) GetOnce(ctx context.Context, id int64) (Product, error) {
	panic("TODO: GetOnce")
}
