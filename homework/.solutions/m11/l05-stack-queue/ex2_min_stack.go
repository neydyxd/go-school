package l05stackqueue

// MinStack это стек целых чисел, который за O(1) отвечает, какой элемент
// в нём сейчас минимальный. Нулевое значение готово к работе:
// var s MinStack; s.Push(3).
//
// Все методы работают за O(1) (Push за амортизированное O(1)).
// Поиск минимума перебором стека не уложится на большом входе.
type MinStack struct {
	vals []int
	mins []int // mins[i] это минимум среди vals[0..i]
}

// Push кладёт v на вершину.
func (s *MinStack) Push(v int) {
	m := v
	if len(s.mins) > 0 {
		m = min(m, s.mins[len(s.mins)-1])
	}
	s.vals = append(s.vals, v)
	s.mins = append(s.mins, m)
}

// Pop снимает и возвращает вершину. На пустом стеке возвращает 0, false.
func (s *MinStack) Pop() (int, bool) {
	if len(s.vals) == 0 {
		return 0, false
	}
	v := s.vals[len(s.vals)-1]
	s.vals = s.vals[:len(s.vals)-1]
	s.mins = s.mins[:len(s.mins)-1]
	return v, true
}

// Top возвращает вершину, не снимая её. На пустом стеке 0, false.
func (s *MinStack) Top() (int, bool) {
	if len(s.vals) == 0 {
		return 0, false
	}
	return s.vals[len(s.vals)-1], true
}

// Min возвращает минимальный элемент стека. На пустом стеке 0, false.
func (s *MinStack) Min() (int, bool) {
	if len(s.mins) == 0 {
		return 0, false
	}
	return s.mins[len(s.mins)-1], true
}

// Len возвращает число элементов.
func (s *MinStack) Len() int {
	return len(s.vals)
}
