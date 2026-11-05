package l03internalsperf

// Ring это кольцевой буфер фиксированной ёмкости, устроенный как буфер
// канала в runtime/chan.go. Поля названы так же, как в hchan, тест
// заглядывает в buf, поэтому поля не переименовывай.
type Ring[T any] struct {
	buf   []T // слоты, len(buf) это ёмкость (dataqsiz в hchan)
	sendx int // куда писать следующий элемент
	recvx int // откуда читать следующий элемент
	count int // сколько элементов сейчас лежит (qcount в hchan)
}

// NewRing: make сам паникует на отрицательной длине, но с невнятным
// сообщением, поэтому проверяем явно.
func NewRing[T any](capacity int) *Ring[T] {
	if capacity < 0 {
		panic("NewRing: отрицательная ёмкость")
	}
	return &Ring[T]{buf: make([]T, capacity)}
}

// Push повторяет ветку «есть место в буфере» из chansend.
func (r *Ring[T]) Push(v T) bool {
	if r.count == len(r.buf) {
		return false
	}
	r.buf[r.sendx] = v
	r.sendx++
	if r.sendx == len(r.buf) {
		r.sendx = 0
	}
	r.count++
	return true
}

// Pop повторяет ветку «в буфере есть данные» из chanrecv.
func (r *Ring[T]) Pop() (T, bool) {
	var zero T
	if r.count == 0 {
		return zero, false
	}
	v := r.buf[r.recvx]
	r.buf[r.recvx] = zero
	r.recvx++
	if r.recvx == len(r.buf) {
		r.recvx = 0
	}
	r.count--
	return v, true
}

// Len возвращает число элементов (как len(ch)).
func (r *Ring[T]) Len() int { return r.count }

// Cap возвращает ёмкость (как cap(ch)).
func (r *Ring[T]) Cap() int { return len(r.buf) }
