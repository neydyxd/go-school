package l02livecoding

import (
	"sync"
	"time"
)

// Cache: потокобезопасный кеш со временем жизни записей.
type Cache[K comparable, V any] struct {
	mu  sync.RWMutex
	ttl time.Duration
	m   map[K]entry[V]
}

type entry[V any] struct {
	v   V
	exp time.Time // нулевое: не протухает
}

func (e entry[V]) alive(now time.Time) bool { return e.exp.IsZero() || now.Before(e.exp) }

// NewCache создаёт пустой кеш.
func NewCache[K comparable, V any](ttl time.Duration) *Cache[K, V] {
	return &Cache[K, V]{ttl: ttl, m: map[K]entry[V]{}}
}

// Set кладёт значение; перезапись продлевает срок жизни.
func (c *Cache[K, V]) Set(k K, v V) {
	e := entry[V]{v: v}
	if c.ttl > 0 {
		e.exp = time.Now().Add(c.ttl)
	}
	c.mu.Lock()
	c.m[k] = e
	c.mu.Unlock()
}

// Get возвращает живое значение.
func (c *Cache[K, V]) Get(k K) (V, bool) {
	c.mu.RLock()
	e, ok := c.m[k]
	c.mu.RUnlock()
	if !ok || !e.alive(time.Now()) {
		var zero V
		return zero, false
	}
	return e.v, true
}

// Delete удаляет запись.
func (c *Cache[K, V]) Delete(k K) {
	c.mu.Lock()
	delete(c.m, k)
	c.mu.Unlock()
}

// Len возвращает число живых записей.
// ponytail: O(n) под RLock; если Len горячий, считать живые при Cleanup по таймеру
func (c *Cache[K, V]) Len() int {
	now := time.Now()
	c.mu.RLock()
	defer c.mu.RUnlock()
	n := 0
	for _, e := range c.m {
		if e.alive(now) {
			n++
		}
	}
	return n
}

// Cleanup удаляет протухшие записи и возвращает их число.
func (c *Cache[K, V]) Cleanup() int {
	now := time.Now()
	c.mu.Lock()
	defer c.mu.Unlock()
	n := 0
	for k, e := range c.m {
		if !e.alive(now) {
			delete(c.m, k) // удалять при range по мапе разрешено
			n++
		}
	}
	return n
}
