package l07maps

import "sync"

// SafeCounter: счётчики по ключам, которыми можно пользоваться из многих горутин.
// Обычная мапа при одновременной записи роняет программу (fatal error: concurrent map writes),
// поэтому доступ к мапе защищай мьютексом mu.
type SafeCounter struct {
	mu     sync.Mutex
	counts map[string]int
}

// NewSafeCounter создаёт готовый к работе счётчик.
func NewSafeCounter() *SafeCounter {
	panic("TODO: реализуй NewSafeCounter")
}

// Inc увеличивает счётчик key на 1.
func (c *SafeCounter) Inc(key string) {
	panic("TODO: реализуй Inc")
}

// Get возвращает текущее значение счётчика key (0, если его не было).
func (c *SafeCounter) Get(key string) int {
	panic("TODO: реализуй Get")
}

// Snapshot возвращает копию всех счётчиков. Изменения копии не должны
// влиять на счётчик, а дальнейшие Inc не должны менять уже выданную копию.
func (c *SafeCounter) Snapshot() map[string]int {
	panic("TODO: реализуй Snapshot")
}
