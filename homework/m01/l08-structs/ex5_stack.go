package l08structs

// Stack: стек строк, нулевое значение которого сразу готово к работе:
//
//	var s Stack
//	s.Push("a")
//
// Никакого конструктора NewStack: в Go ценят типы, которые работают без инициализации,
// как bytes.Buffer или sync.Mutex. Поля придумай сам.
type Stack struct{}

// Push кладёт v на вершину.
func (s *Stack) Push(v string) {
	panic("TODO: реализуй Push")
}

// Pop снимает элемент с вершины. Пустой стек: "", false.
func (s *Stack) Pop() (string, bool) {
	panic("TODO: реализуй Pop")
}

// Len возвращает число элементов.
func (s *Stack) Len() int {
	panic("TODO: реализуй Len")
}
