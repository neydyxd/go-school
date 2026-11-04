package l07problems

import (
	"maps"
	"sync"
)

// Counter: потокобезопасный счётчик событий по ключам.
// Нулевое значение готово к работе: var c Counter; c.Inc("x").
// Counter нельзя копировать после первого использования (в нём мьютекс).
type Counter struct {
	mu sync.Mutex
	m  map[string]int
}

// Inc увеличивает счётчик key на 1. Безопасен из многих горутин.
func (c *Counter) Inc(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.m == nil {
		c.m = make(map[string]int)
	}
	c.m[key]++
}

// Get возвращает текущее значение счётчика key (0, если не было).
func (c *Counter) Get(key string) int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.m[key]
}

// Snapshot возвращает КОПИЮ всех счётчиков: изменения копии не влияют на
// Counter, а последующие Inc не меняют уже выданную копию.
// Для пустого счётчика вернуть пустую не-nil мапу.
func (c *Counter) Snapshot() map[string]int {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := make(map[string]int, len(c.m))
	maps.Copy(out, c.m)
	return out
}
