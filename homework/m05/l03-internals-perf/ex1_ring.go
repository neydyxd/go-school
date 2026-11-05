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

// NewRing создаёт буфер на capacity элементов. Отрицательная ёмкость это
// паника (как у make(chan T, -1)). Ёмкость 0 допустима: в такой буфер
// ничего не положить.
func NewRing[T any](capacity int) *Ring[T] {
	panic("TODO: NewRing")
}

// Push кладёт v в хвост. Если буфер полон, ничего не меняет и возвращает
// false. Индексы идут по кругу: после последнего слота снова слот 0, как
// c.sendx в chansend (там без деления по модулю, просто сброс в 0).
func (r *Ring[T]) Push(v T) bool {
	panic("TODO: Push")
}

// Pop забирает элемент из головы (FIFO). Если пусто, возвращает нулевое
// значение и false. Освободившийся слот обязательно обнули, как делает
// chanrecv (typedmemclr): иначе буфер держит ссылку на уже выданный объект,
// и сборщик мусора не сможет его освободить.
func (r *Ring[T]) Pop() (T, bool) {
	panic("TODO: Pop")
}

// Len возвращает число элементов (как len(ch)).
func (r *Ring[T]) Len() int {
	panic("TODO: Len")
}

// Cap возвращает ёмкость (как cap(ch)).
func (r *Ring[T]) Cap() int {
	panic("TODO: Cap")
}
