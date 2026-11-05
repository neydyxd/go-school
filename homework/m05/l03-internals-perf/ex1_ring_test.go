package l03internalsperf

import (
	"slices"
	"testing"
)

func TestRingFIFOAndWrap(t *testing.T) {
	r := NewRing[int](3)
	if r.Cap() != 3 || r.Len() != 0 {
		t.Fatalf("NewRing(3): Cap() = %d, Len() = %d, want 3, 0", r.Cap(), r.Len())
	}
	var got []int
	push := func(v int, want bool) {
		t.Helper()
		if ok := r.Push(v); ok != want {
			t.Fatalf("Push(%d) = %v, want %v (Len %d, Cap %d)", v, ok, want, r.Len(), r.Cap())
		}
	}
	pop := func() {
		t.Helper()
		v, ok := r.Pop()
		if !ok {
			t.Fatalf("Pop() на непустом буфере вернул false")
		}
		got = append(got, v)
	}
	push(1, true)
	push(2, true)
	push(3, true)
	push(4, false) // полон
	pop()          // 1
	push(4, true)  // sendx переходит через конец
	pop()          // 2
	pop()          // 3
	push(5, true)
	push(6, true)
	if r.Len() != 3 {
		t.Errorf("Len() = %d, want 3", r.Len())
	}
	for r.Len() > 0 {
		pop()
	}
	if want := []int{1, 2, 3, 4, 5, 6}; !slices.Equal(got, want) {
		t.Errorf("порядок Pop = %v, want %v", got, want)
	}
	if v, ok := r.Pop(); ok || v != 0 {
		t.Errorf("Pop() на пустом = %d, %v, want 0, false", v, ok)
	}
}

func TestRingLongRun(t *testing.T) {
	// Много кругов: индексы не должны уходить за пределы буфера.
	r := NewRing[int](4)
	next, want := 0, 0
	for i := range 1000 {
		for range i%4 + 1 {
			if r.Push(next) {
				next++
			}
		}
		for range i%3 + 1 {
			v, ok := r.Pop()
			if !ok {
				break
			}
			if v != want {
				t.Fatalf("итерация %d: Pop() = %d, want %d", i, v, want)
			}
			want++
		}
		if r.Len() < 0 || r.Len() > r.Cap() {
			t.Fatalf("итерация %d: Len() = %d вне [0, %d]", i, r.Len(), r.Cap())
		}
	}
}

func TestRingZeroCapacity(t *testing.T) {
	r := NewRing[string](0)
	if r.Push("x") {
		t.Error(`Push("x") в буфер ёмкости 0 = true, want false`)
	}
	if _, ok := r.Pop(); ok {
		t.Error("Pop() из буфера ёмкости 0 = true, want false")
	}
}

func TestRingNegativeCapacity(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("NewRing(-1) не запаниковал, want панику")
		}
	}()
	NewRing[int](-1)
}

func TestRingClearsSlot(t *testing.T) {
	r := NewRing[*int](2)
	a, b := new(1), new(2)
	r.Push(a)
	r.Push(b)
	r.Pop()
	if len(r.buf) != 2 {
		t.Fatalf("len(buf) = %d, want 2", len(r.buf))
	}
	for i, p := range r.buf {
		if p == a {
			t.Errorf("buf[%d] после Pop всё ещё держит выданный указатель, want nil (обнули слот)", i)
		}
	}
}

func TestRingAllocs(t *testing.T) {
	if raceEnabled {
		t.Skip("под -race счётчик аллокаций не показателен")
	}
	r := NewRing[int](8)
	allocs := testing.AllocsPerRun(100, func() {
		for i := range 5 {
			r.Push(i)
		}
		for range 5 {
			r.Pop()
		}
	})
	if allocs != 0 {
		t.Errorf("Push/Pop: %.0f аллокаций на круг, want 0 (буфер выделяется один раз в NewRing)", allocs)
	}
}
