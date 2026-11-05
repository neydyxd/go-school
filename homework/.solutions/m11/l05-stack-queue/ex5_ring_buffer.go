package l05stackqueue

// Ring это очередь FIFO фиксированной ёмкости на кольцевом буфере:
// один слайс длины capacity, индекс головы и число элементов. Память
// выделяется один раз в NewRing, дальше Push и Pop работают за O(1)
// без сдвигов и аллокаций.
type Ring struct {
	buf  []int
	head int // индекс первого элемента
	n    int // сколько элементов лежит
}

// NewRing создаёт буфер ёмкости capacity. При capacity <= 0 буфер
// создаётся, но ничего не вмещает: любой Push вернёт false.
func NewRing(capacity int) *Ring {
	return &Ring{buf: make([]int, max(capacity, 0))}
}

// Push добавляет v в конец. Если буфер полон, ничего не меняет и
// возвращает false.
func (r *Ring) Push(v int) bool {
	if r.n == len(r.buf) {
		return false
	}
	r.buf[(r.head+r.n)%len(r.buf)] = v
	r.n++
	return true
}

// Pop снимает и возвращает элемент из начала. На пустом буфере 0, false.
func (r *Ring) Pop() (int, bool) {
	if r.n == 0 {
		return 0, false
	}
	v := r.buf[r.head]
	r.head = (r.head + 1) % len(r.buf)
	r.n--
	return v, true
}

// Len возвращает число элементов, Cap ёмкость буфера.
func (r *Ring) Len() int {
	return r.n
}

func (r *Ring) Cap() int {
	return len(r.buf)
}
