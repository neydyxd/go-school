package l07maps

// LRU: кеш фиксированной ёмкости, который при переполнении выкидывает
// элемент, к которому дольше всего не обращались (least recently used).
// Get и Put должны работать за O(1): мапа ключ -> узел двусвязного списка
// (container/list) плюс сам список в порядке использования.
//
// Поля структуры придумай сам.
type LRU struct{}

// NewLRU создаёт кеш на capacity элементов. capacity <= 0: кеш ничего не хранит.
func NewLRU(capacity int) *LRU {
	panic("TODO: реализуй NewLRU")
}

// Get возвращает значение и true, если ключ есть. Обращение делает ключ «свежим».
func (c *LRU) Get(key string) (int, bool) {
	panic("TODO: реализуй Get")
}

// Put добавляет или обновляет значение и делает ключ «свежим».
// Если элементов стало больше capacity, удаляет самый давно использованный.
func (c *LRU) Put(key string, value int) {
	panic("TODO: реализуй Put")
}

// Len возвращает число элементов в кеше.
func (c *LRU) Len() int {
	panic("TODO: реализуй Len")
}
