package l07linkedlists

import "container/list"

// LRU это кеш фиксированной ёмкости с вытеснением давно не используемых
// записей (least recently used). Get и Put делают ключ «самым свежим».
// Когда Put добавляет новый ключ в заполненный кеш, сначала вытесняется
// самый давно использованный ключ.
//
// Get и Put работают за O(1). Классическое устройство: map от ключа
// к элементу двусвязного списка (container/list), в списке порядок
// использования. Поиск самого старого ключа перебором не уложится
// на большом входе.
type LRU struct {
	capacity int
	order    *list.List            // спереди самые свежие, сзади кандидат на вытеснение
	items    map[int]*list.Element // ключ -> элемент списка, Value = *entry
}

type entry struct {
	key, value int // ключ нужен, чтобы при вытеснении удалить его из map
}

// NewLRU создаёт кеш ёмкости capacity >= 1.
func NewLRU(capacity int) *LRU {
	return &LRU{capacity: capacity, order: list.New(), items: make(map[int]*list.Element)}
}

// Get возвращает значение и true, если ключ есть, иначе 0, false.
func (c *LRU) Get(key int) (int, bool) {
	e, ok := c.items[key]
	if !ok {
		return 0, false
	}
	c.order.MoveToFront(e)
	return e.Value.(*entry).value, true
}

// Put записывает значение. Если ключ уже есть, обновляет его значение
// (это не вытесняет никого).
func (c *LRU) Put(key, value int) {
	if e, ok := c.items[key]; ok {
		e.Value.(*entry).value = value
		c.order.MoveToFront(e)
		return
	}
	if c.order.Len() == c.capacity {
		oldest := c.order.Back()
		c.order.Remove(oldest)
		delete(c.items, oldest.Value.(*entry).key)
	}
	c.items[key] = c.order.PushFront(&entry{key, value})
}

// Len возвращает число ключей в кеше.
func (c *LRU) Len() int {
	return c.order.Len()
}
