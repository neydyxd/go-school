package l07maps

import (
	"maps"
	"sync"
)

// SafeCounter: счётчики по ключам, которыми можно пользоваться из многих горутин.
// Обычная мапа при одновременной записи роняет программу (fatal error: concurrent map writes),
// поэтому доступ к мапе защищай мьютексом mu.
type SafeCounter struct {
	mu     sync.Mutex
	counts map[string]int
}

// NewSafeCounter создаёт готовый к работе счётчик.
func NewSafeCounter() *SafeCounter {
	return &SafeCounter{counts: make(map[string]int)}
}

// Inc увеличивает счётчик key на 1.
func (c *SafeCounter) Inc(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.counts[key]++
}

// Get возвращает текущее значение счётчика key (0, если его не было).
func (c *SafeCounter) Get(key string) int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.counts[key]
}

// Snapshot возвращает копию всех счётчиков. Изменения копии не должны
// влиять на счётчик, а дальнейшие Inc не должны менять уже выданную копию.
func (c *SafeCounter) Snapshot() map[string]int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return maps.Clone(c.counts) // вернуть c.counts нельзя: мапа это ссылка на ту же таблицу
}
