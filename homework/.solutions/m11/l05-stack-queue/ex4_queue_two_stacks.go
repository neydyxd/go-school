package l05stackqueue

// Queue это очередь FIFO, собранная из двух стеков на слайсах: в один
// только кладём, из другого только снимаем. Нулевое значение готово
// к работе: var q Queue[string]. Тест проверяет устройство: в структуре
// ровно два поля-слайса.
//
// Push, Pop и Peek работают за амортизированное O(1). Перекладывать все
// элементы туда и обратно на каждый Pop нельзя: это O(n) на операцию
// и не уложится на большом входе.
type Queue[T any] struct {
	in  []T // сюда кладём
	out []T // отсюда снимаем, вершина это начало очереди
}

// Push добавляет v в конец очереди.
func (q *Queue[T]) Push(v T) {
	q.in = append(q.in, v)
}

// shift перекладывает in в out, только когда out пуст. Каждый элемент
// перекладывается один раз за жизнь, отсюда амортизированное O(1).
func (q *Queue[T]) shift() {
	if len(q.out) > 0 {
		return
	}
	for len(q.in) > 0 {
		last := len(q.in) - 1
		q.out = append(q.out, q.in[last])
		var zero T
		q.in[last] = zero // не держим ссылку для сборщика мусора
		q.in = q.in[:last]
	}
}

// Pop снимает и возвращает элемент из начала очереди.
// На пустой очереди возвращает нулевое значение T и false.
func (q *Queue[T]) Pop() (T, bool) {
	q.shift()
	var zero T
	if len(q.out) == 0 {
		return zero, false
	}
	last := len(q.out) - 1
	v := q.out[last]
	q.out[last] = zero
	q.out = q.out[:last]
	return v, true
}

// Peek возвращает элемент из начала очереди, не снимая его.
// На пустой очереди возвращает нулевое значение T и false.
func (q *Queue[T]) Peek() (T, bool) {
	q.shift()
	if len(q.out) == 0 {
		var zero T
		return zero, false
	}
	return q.out[len(q.out)-1], true
}

// Len возвращает число элементов в очереди.
func (q *Queue[T]) Len() int {
	return len(q.in) + len(q.out)
}
