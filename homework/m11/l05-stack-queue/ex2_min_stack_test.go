package l05stackqueue

import "testing"

func TestMinStack(t *testing.T) {
	var s MinStack
	if v, ok := s.Pop(); ok || v != 0 {
		t.Fatalf("Pop() на пустом = %d, %v, want 0, false", v, ok)
	}
	if v, ok := s.Min(); ok || v != 0 {
		t.Fatalf("Min() на пустом = %d, %v, want 0, false", v, ok)
	}
	if v, ok := s.Top(); ok || v != 0 {
		t.Fatalf("Top() на пустом = %d, %v, want 0, false", v, ok)
	}
	// Шаги: операция и ожидаемые Top, Min, Len после неё.
	steps := []struct {
		op              string
		arg             int
		top, minV, size int
	}{
		{"push", 5, 5, 5, 1},
		{"push", 3, 3, 3, 2},
		{"push", 7, 7, 3, 3},
		{"push", 3, 3, 3, 4}, // дубль минимума
		{"pop", 3, 7, 3, 3},  // снят один из двух минимумов, второй остался
		{"pop", 7, 3, 3, 2},
		{"pop", 3, 5, 5, 1}, // минимум вернулся к 5
		{"push", -2, -2, -2, 2},
		{"push", 10, 10, -2, 3},
	}
	for i, st := range steps {
		switch st.op {
		case "push":
			s.Push(st.arg)
		case "pop":
			if v, ok := s.Pop(); !ok || v != st.arg {
				t.Fatalf("шаг %d: Pop() = %d, %v, want %d, true", i, v, ok, st.arg)
			}
		}
		if v, ok := s.Top(); !ok || v != st.top {
			t.Fatalf("шаг %d (%s %d): Top() = %d, %v, want %d, true", i, st.op, st.arg, v, ok, st.top)
		}
		if v, ok := s.Min(); !ok || v != st.minV {
			t.Fatalf("шаг %d (%s %d): Min() = %d, %v, want %d, true", i, st.op, st.arg, v, ok, st.minV)
		}
		if got := s.Len(); got != st.size {
			t.Fatalf("шаг %d (%s %d): Len() = %d, want %d", i, st.op, st.arg, got, st.size)
		}
	}
	for s.Len() > 0 {
		s.Pop()
	}
	if v, ok := s.Min(); ok {
		t.Fatalf("Min() после опустошения = %d, true, want 0, false", v)
	}
}

func TestMinStackBig(t *testing.T) {
	const n = 200_000
	var s MinStack
	bad := -1
	within(t, func() {
		// убывающая последовательность: минимум меняется на каждом Push,
		// затем всё снимаем и проверяем Min на каждом шаге
		for i := range n {
			s.Push(n - i)
			if m, _ := s.Min(); m != n-i && bad < 0 {
				bad = i
			}
		}
		for i := range n - 1 {
			s.Pop()
			if m, _ := s.Min(); m != i+2 && bad < 0 {
				bad = n + i
			}
		}
	})
	if bad >= 0 {
		t.Fatalf("Min() вернул неверное значение на шаге %d большого теста", bad)
	}
}
