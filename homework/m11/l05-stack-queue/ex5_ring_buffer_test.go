package l05stackqueue

import "testing"

func TestRing(t *testing.T) {
	r := NewRing(3)
	if r.Cap() != 3 || r.Len() != 0 {
		t.Fatalf("NewRing(3): Cap() = %d, Len() = %d, want 3, 0", r.Cap(), r.Len())
	}
	if v, ok := r.Pop(); ok {
		t.Fatalf("Pop() на пустом = %d, true, want 0, false", v)
	}
	for _, v := range []int{1, 2, 3} {
		if !r.Push(v) {
			t.Fatalf("Push(%d) = false на неполном буфере", v)
		}
	}
	if r.Push(4) {
		t.Fatal("Push(4) = true на полном буфере, want false")
	}
	if r.Len() != 3 {
		t.Fatalf("Len() после отказа = %d, want 3", r.Len())
	}
	if v, ok := r.Pop(); !ok || v != 1 {
		t.Fatalf("Pop() = %d, %v, want 1, true", v, ok)
	}
	// хвост переходит через конец слайса в начало
	for _, v := range []int{4} {
		if !r.Push(v) {
			t.Fatalf("Push(%d) = false после Pop", v)
		}
	}
	for _, want := range []int{2, 3, 4} {
		if v, ok := r.Pop(); !ok || v != want {
			t.Fatalf("Pop() = %d, %v, want %d, true", v, ok, want)
		}
	}
	if v, ok := r.Pop(); ok {
		t.Fatalf("Pop() на опустевшем = %d, true, want 0, false", v)
	}
	// много кругов подряд: голова и хвост многократно проходят через границу
	next, want := 0, 0
	for round := range 10 {
		for r.Len() < r.Cap() {
			r.Push(next)
			next++
		}
		for range 2 {
			if v, ok := r.Pop(); !ok || v != want {
				t.Fatalf("круг %d: Pop() = %d, %v, want %d, true", round, v, ok, want)
			}
			want++
		}
	}
}

func TestRingTiny(t *testing.T) {
	for _, c := range []int{0, -5} {
		r := NewRing(c)
		if r.Push(1) {
			t.Errorf("NewRing(%d).Push(1) = true, want false", c)
		}
		if r.Len() != 0 || r.Cap() != 0 {
			t.Errorf("NewRing(%d): Len() = %d, Cap() = %d, want 0, 0", c, r.Len(), r.Cap())
		}
		if _, ok := r.Pop(); ok {
			t.Errorf("NewRing(%d).Pop() вернул true", c)
		}
	}
	r := NewRing(1)
	for i := range 5 {
		if !r.Push(i) || r.Push(i) {
			t.Fatalf("NewRing(1): первый Push должен пройти, второй нет (шаг %d)", i)
		}
		if v, ok := r.Pop(); !ok || v != i {
			t.Fatalf("NewRing(1).Pop() = %d, %v, want %d, true", v, ok, i)
		}
	}
}

func TestRingBig(t *testing.T) {
	const capacity, ops = 100_000, 1_000_000
	r := NewRing(capacity)
	bad := false
	within(t, func() {
		for i := range capacity {
			r.Push(i)
		}
		// буфер полон: снимаем один, кладём один. Сдвиг слайса на каждом
		// Pop это O(capacity) на операцию.
		for i := range ops {
			v, ok := r.Pop()
			if !ok || v != i {
				bad = true
				return
			}
			r.Push(capacity + i)
		}
	})
	if bad {
		t.Fatal("Pop() вернул не тот элемент на большом тесте")
	}
	// AllocsPerRun делит число аллокаций на число прогонов нацело, поэтому
	// меряем 100 000 операций одним прогоном: очередь на append и q[1:]
	// перевыделяет массив редко, но всё же перевыделяет.
	allocs := testing.AllocsPerRun(1, func() {
		for range 100_000 {
			r.Pop()
			r.Push(1)
		}
	})
	if allocs != 0 {
		t.Errorf("%v аллокаций на 100000 Push+Pop заполненного буфера, want 0: память выделяется только в NewRing", allocs)
	}
}
