package l02livecoding

import "time"

// Cache: потокобезопасный кеш со временем жизни записей.
// Запись видна, пока с момента последнего Set прошло меньше ttl; ровно через ttl
// её уже нет. ttl <= 0 значит «записи не протухают».
type Cache[K comparable, V any] struct {
	// добавь нужные поля
}

// NewCache создаёт пустой кеш.
func NewCache[K comparable, V any](ttl time.Duration) *Cache[K, V] {
	panic("TODO: реализуй NewCache")
}

// Set кладёт значение; перезапись продлевает срок жизни.
func (c *Cache[K, V]) Set(k K, v V) {
	panic("TODO: реализуй Set")
}

// Get возвращает значение и true; для отсутствующей или протухшей записи
// нулевое значение и false.
func (c *Cache[K, V]) Get(k K) (V, bool) {
	panic("TODO: реализуй Get")
}

// Delete удаляет запись (отсутствующая запись не ошибка).
func (c *Cache[K, V]) Delete(k K) {
	panic("TODO: реализуй Delete")
}

// Len возвращает число живых (не протухших) записей.
func (c *Cache[K, V]) Len() int {
	panic("TODO: реализуй Len")
}

// Cleanup удаляет протухшие записи из памяти и возвращает, сколько удалил.
func (c *Cache[K, V]) Cleanup() int {
	panic("TODO: реализуй Cleanup")
}
