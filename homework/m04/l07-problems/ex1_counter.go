package l07problems

// Counter: потокобезопасный счётчик событий по ключам.
// Нулевое значение готово к работе: var c Counter; c.Inc("x").
// Counter нельзя копировать после первого использования (в нём будет мьютекс).
type Counter struct {
	// TODO: поля
}

// Inc увеличивает счётчик key на 1. Безопасен из многих горутин.
func (c *Counter) Inc(key string) {
	panic("TODO: реализуй Inc")
}

// Get возвращает текущее значение счётчика key (0, если не было).
func (c *Counter) Get(key string) int {
	panic("TODO: реализуй Get")
}

// Snapshot возвращает КОПИЮ всех счётчиков: изменения копии не влияют на
// Counter, а последующие Inc не меняют уже выданную копию.
// Для пустого счётчика вернуть пустую не-nil мапу.
func (c *Counter) Snapshot() map[string]int {
	panic("TODO: реализуй Snapshot")
}
