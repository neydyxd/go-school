package l08structs

import "testing"

func TestStackZeroValue(t *testing.T) {
	var s Stack
	if s.Len() != 0 {
		t.Fatalf("нулевой Stack: Len() = %d, want 0", s.Len())
	}
	if v, ok := s.Pop(); ok || v != "" {
		t.Errorf("Pop() на пустом стеке = %q, %v, want \"\", false", v, ok)
	}
	s.Push("a")
	s.Push("b")
	s.Push("c")
	steps := []struct {
		name   string
		want   string
		wantOK bool
	}{
		{"последний пришёл", "c", true},
		{"следующий", "b", true},
		{"первый", "a", true},
		{"снова пусто", "", false},
	}
	for _, st := range steps {
		t.Run(st.name, func(t *testing.T) {
			if v, ok := s.Pop(); v != st.want || ok != st.wantOK {
				t.Errorf("Pop() = %q, %v, want %q, %v", v, ok, st.want, st.wantOK)
			}
		})
	}
	if s.Len() != 0 {
		t.Errorf("после всех Pop: Len() = %d, want 0", s.Len())
	}
}

func TestStackInStruct(t *testing.T) {
	// Стек как поле другой структуры тоже работает без инициализации.
	var editor struct {
		undo Stack
	}
	editor.undo.Push("ввод")
	if editor.undo.Len() != 1 {
		t.Errorf("Len() = %d, want 1", editor.undo.Len())
	}
}
