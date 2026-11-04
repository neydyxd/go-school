package l03functions

import "testing"

func TestMemoize(t *testing.T) {
	calls := 0
	square := func(x int) int {
		calls++
		return x * x
	}
	m := Memoize(square)
	args := []int{3, 3, 4, 3, 4, 0, 0}
	want := []int{9, 9, 16, 9, 16, 0, 0}
	for i, a := range args {
		if got := m(a); got != want[i] {
			t.Errorf("Memoize(square)(%d) = %d, want %d", a, got, want[i])
		}
	}
	if calls != 3 {
		t.Errorf("исходная функция вызвана %d раз, want 3 (по разу на аргументы 3, 4, 0)", calls)
	}
}

func TestMemoizeSeparateCaches(t *testing.T) {
	calls := 0
	id := func(x int) int {
		calls++
		return x
	}
	a := Memoize(id)
	b := Memoize(id)
	a(1)
	b(1)
	if calls != 2 {
		t.Errorf("две обёртки Memoize вызвали функцию %d раз, want 2: у каждой должен быть свой кеш", calls)
	}
}
