package l04generics

import "testing"

func TestSet(t *testing.T) {
	t.Run("дубликаты схлопываются", func(t *testing.T) {
		s := NewSet("go", "rust", "go")
		if s.Len() != 2 || !s.Has("go") || !s.Has("rust") || s.Has("java") {
			t.Errorf("NewSet(go, rust, go): Len=%d, Has(go)=%v, Has(rust)=%v, Has(java)=%v; want 2, true, true, false",
				s.Len(), s.Has("go"), s.Has("rust"), s.Has("java"))
		}
	})
	t.Run("пустое множество", func(t *testing.T) {
		s := NewSet[int]()
		if s.Len() != 0 || s.Has(0) {
			t.Errorf("NewSet[int](): Len=%d, Has(0)=%v; want 0, false", s.Len(), s.Has(0))
		}
		s.Add(0)
		if s.Len() != 1 || !s.Has(0) {
			t.Errorf("после Add(0): Len=%d, Has(0)=%v; want 1, true", s.Len(), s.Has(0))
		}
	})
	t.Run("Remove", func(t *testing.T) {
		s := NewSet(1, 2, 3)
		s.Remove(2)
		s.Remove(42)
		if s.Len() != 2 || s.Has(2) {
			t.Errorf("после Remove(2), Remove(42): Len=%d, Has(2)=%v; want 2, false", s.Len(), s.Has(2))
		}
	})
	t.Run("Union и Intersect не меняют исходные", func(t *testing.T) {
		a, b := NewSet(1, 2, 3), NewSet(3, 4)
		u := a.Union(b)
		in := a.Intersect(b)
		if u.Len() != 4 || !u.Has(1) || !u.Has(4) {
			t.Errorf("Union: Len=%d, want 4 и элементы 1..4", u.Len())
		}
		if in.Len() != 1 || !in.Has(3) {
			t.Errorf("Intersect: Len=%d, Has(3)=%v; want 1, true", in.Len(), in.Has(3))
		}
		if a.Len() != 3 || b.Len() != 2 || a.Has(4) {
			t.Errorf("исходные изменились: a.Len=%d, b.Len=%d, a.Has(4)=%v; want 3, 2, false", a.Len(), b.Len(), a.Has(4))
		}
		u.Add(100)
		if a.Has(100) || b.Has(100) {
			t.Errorf("Union вернул не новое множество: Add в результат попал в исходное")
		}
	})
	t.Run("структуры как элементы", func(t *testing.T) {
		type pt struct{ X, Y int }
		s := NewSet(pt{1, 2}, pt{1, 2}, pt{2, 1})
		if s.Len() != 2 || !s.Has(pt{2, 1}) {
			t.Errorf("Set[pt]: Len=%d, want 2", s.Len())
		}
	})
}
