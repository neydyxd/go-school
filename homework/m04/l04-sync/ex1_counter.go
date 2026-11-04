package l04sync

// Counter считает события по ключам и безопасен для одновременного
// использования из многих горутин. Защити мапу мьютексом.
type Counter struct {
	// TODO: поля
}

// NewCounter создаёт пустой счётчик.
func NewCounter() *Counter {
	panic("TODO: реализуй NewCounter")
}

// Inc увеличивает счётчик ключа key на 1.
func (c *Counter) Inc(key string) {
	panic("TODO: реализуй Inc")
}

// Get возвращает текущее значение ключа (0, если ключа не было).
func (c *Counter) Get(key string) int {
	panic("TODO: реализуй Get")
}

// Snapshot возвращает КОПИЮ всех счётчиков. Вызывающий может читать
// и менять её сколько угодно, пока другие горутины вызывают Inc:
// ни гонки, ни влияния на сам Counter быть не должно.
func (c *Counter) Snapshot() map[string]int {
	panic("TODO: реализуй Snapshot")
}
