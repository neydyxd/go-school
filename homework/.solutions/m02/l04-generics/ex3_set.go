package l04generics

type Set[T comparable] struct {
	m map[T]struct{}
}

func NewSet[T comparable](items ...T) *Set[T] {
	s := &Set[T]{m: make(map[T]struct{}, len(items))}
	for _, v := range items {
		s.Add(v)
	}
	return s
}

func (s *Set[T]) Add(v T)      { s.m[v] = struct{}{} }
func (s *Set[T]) Has(v T) bool { _, ok := s.m[v]; return ok }
func (s *Set[T]) Remove(v T)   { delete(s.m, v) }
func (s *Set[T]) Len() int     { return len(s.m) }

func (s *Set[T]) Union(other *Set[T]) *Set[T] {
	out := NewSet[T]()
	for v := range s.m {
		out.Add(v)
	}
	for v := range other.m {
		out.Add(v)
	}
	return out
}

func (s *Set[T]) Intersect(other *Set[T]) *Set[T] {
	out := NewSet[T]()
	for v := range s.m {
		if other.Has(v) {
			out.Add(v)
		}
	}
	return out
}
