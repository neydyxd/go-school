package l08structs

// Stack: стек строк, нулевое значение которого сразу готово к работе:
//
//	var s Stack
//	s.Push("a")
//
// Никакого конструктора NewStack: в Go ценят типы, которые работают без инициализации,
// как bytes.Buffer или sync.Mutex. Поля придумай сам.
type Stack struct {
	items []string // nil-слайс уже годится для append
}

// Push кладёт v на вершину.
func (s *Stack) Push(v string) {
	s.items = append(s.items, v)
}

// Pop снимает элемент с вершины. Пустой стек: "", false.
func (s *Stack) Pop() (string, bool) {
	if len(s.items) == 0 {
		return "", false
	}
	last := len(s.items) - 1
	v := s.items[last]
	s.items[last] = "" // не держим ссылку на строку в хвосте массива
	s.items = s.items[:last]
	return v, true
}

// Len возвращает число элементов.
func (s *Stack) Len() int {
	return len(s.items)
}
