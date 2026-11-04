package l07maps

import "container/list"

// LRU: кеш фиксированной ёмкости, который при переполнении выкидывает
// элемент, к которому дольше всего не обращались (least recently used).
// Get и Put должны работать за O(1): мапа ключ -> узел двусвязного списка
// (container/list) плюс сам список в порядке использования.
//
// Поля структуры придумай сам.
type LRU struct {
	capacity int
	order    *list.List               // спереди самые свежие
	items    map[string]*list.Element // значение узла: *entry
}

type entry struct {
	key   string
	value int
}

// NewLRU создаёт кеш на capacity элементов. capacity <= 0: кеш ничего не хранит.
func NewLRU(capacity int) *LRU {
	return &LRU{capacity: capacity, order: list.New(), items: make(map[string]*list.Element)}
}

// Get возвращает значение и true, если ключ есть. Обращение делает ключ «свежим».
func (c *LRU) Get(key string) (int, bool) {
	el, ok := c.items[key]
	if !ok {
		return 0, false
	}
	c.order.MoveToFront(el)
	return el.Value.(*entry).value, true
}

// Put добавляет или обновляет значение и делает ключ «свежим».
// Если элементов стало больше capacity, удаляет самый давно использованный.
func (c *LRU) Put(key string, value int) {
	if c.capacity <= 0 {
		return
	}
	if el, ok := c.items[key]; ok {
		el.Value.(*entry).value = value
		c.order.MoveToFront(el)
		return
	}
	c.items[key] = c.order.PushFront(&entry{key, value})
	if c.order.Len() > c.capacity {
		oldest := c.order.Back()
		c.order.Remove(oldest)
		delete(c.items, oldest.Value.(*entry).key)
	}
}

// Len возвращает число элементов в кеше.
func (c *LRU) Len() int {
	return c.order.Len()
}
