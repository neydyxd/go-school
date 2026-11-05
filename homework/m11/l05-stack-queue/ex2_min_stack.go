package l05stackqueue

// MinStack это стек целых чисел, который за O(1) отвечает, какой элемент
// в нём сейчас минимальный. Нулевое значение готово к работе:
// var s MinStack; s.Push(3).
//
// Все методы работают за O(1) (Push за амортизированное O(1)).
// Поиск минимума перебором стека не уложится на большом входе.
type MinStack struct {
	// TODO: поля
}

// Push кладёт v на вершину.
func (s *MinStack) Push(v int) {
	panic("TODO: реализуй Push")
}

// Pop снимает и возвращает вершину. На пустом стеке возвращает 0, false.
func (s *MinStack) Pop() (int, bool) {
	panic("TODO: реализуй Pop")
}

// Top возвращает вершину, не снимая её. На пустом стеке 0, false.
func (s *MinStack) Top() (int, bool) {
	panic("TODO: реализуй Top")
}

// Min возвращает минимальный элемент стека. На пустом стеке 0, false.
func (s *MinStack) Min() (int, bool) {
	panic("TODO: реализуй Min")
}

// Len возвращает число элементов.
func (s *MinStack) Len() int {
	panic("TODO: реализуй Len")
}
