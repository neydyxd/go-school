package l04sync

import (
	"maps"
	"sync"
)

// Counter считает события по ключам и безопасен для одновременного
// использования из многих горутин. Защити мапу мьютексом.
type Counter struct {
	mu sync.Mutex
	m  map[string]int
}

// NewCounter создаёт пустой счётчик.
func NewCounter() *Counter {
	return &Counter{m: make(map[string]int)}
}

// Inc увеличивает счётчик ключа key на 1.
func (c *Counter) Inc(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.m[key]++
}

// Get возвращает текущее значение ключа (0, если ключа не было).
func (c *Counter) Get(key string) int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.m[key]
}

// Snapshot возвращает КОПИЮ всех счётчиков. Вызывающий может читать
// и менять её сколько угодно, пока другие горутины вызывают Inc:
// ни гонки, ни влияния на сам Counter быть не должно.
func (c *Counter) Snapshot() map[string]int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return maps.Clone(c.m)
}
