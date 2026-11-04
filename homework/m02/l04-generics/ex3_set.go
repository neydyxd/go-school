package l04generics

// Set: множество значений любого сравнимого типа.
// Нулевое значение Set не готово к работе, создавай через NewSet.
type Set[T comparable] struct {
	m map[T]struct{}
}

// NewSet создаёт множество из перечисленных элементов (дубликаты схлопываются).
func NewSet[T comparable](items ...T) *Set[T] {
	panic("TODO: реализуй NewSet")
}

// Add добавляет элемент.
func (s *Set[T]) Add(v T) { panic("TODO: реализуй Add") }

// Has сообщает, есть ли элемент в множестве.
func (s *Set[T]) Has(v T) bool { panic("TODO: реализуй Has") }

// Remove удаляет элемент; если его нет, ничего не происходит.
func (s *Set[T]) Remove(v T) { panic("TODO: реализуй Remove") }

// Len возвращает число элементов.
func (s *Set[T]) Len() int { panic("TODO: реализуй Len") }

// Union возвращает НОВОЕ множество: все элементы s и other.
// s и other не меняются.
func (s *Set[T]) Union(other *Set[T]) *Set[T] { panic("TODO: реализуй Union") }

// Intersect возвращает НОВОЕ множество: элементы, которые есть и в s, и в other.
// s и other не меняются.
func (s *Set[T]) Intersect(other *Set[T]) *Set[T] { panic("TODO: реализуй Intersect") }
