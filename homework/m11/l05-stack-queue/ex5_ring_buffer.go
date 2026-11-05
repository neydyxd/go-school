package l05stackqueue

// Ring это очередь FIFO фиксированной ёмкости на кольцевом буфере:
// один слайс длины capacity, индекс головы и число элементов. Память
// выделяется один раз в NewRing, дальше Push и Pop работают за O(1)
// без сдвигов и аллокаций.
type Ring struct {
	// TODO: поля
}

// NewRing создаёт буфер ёмкости capacity. При capacity <= 0 буфер
// создаётся, но ничего не вмещает: любой Push вернёт false.
func NewRing(capacity int) *Ring {
	panic("TODO: реализуй NewRing")
}

// Push добавляет v в конец. Если буфер полон, ничего не меняет и
// возвращает false.
func (r *Ring) Push(v int) bool {
	panic("TODO: реализуй Push")
}

// Pop снимает и возвращает элемент из начала. На пустом буфере 0, false.
func (r *Ring) Pop() (int, bool) {
	panic("TODO: реализуй Pop")
}

// Len возвращает число элементов, Cap ёмкость буфера.
func (r *Ring) Len() int {
	panic("TODO: реализуй Len")
}

func (r *Ring) Cap() int {
	panic("TODO: реализуй Cap")
}
