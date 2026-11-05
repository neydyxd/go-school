package l11heap

import (
	"math"
	"math/rand/v2"
	"slices"
	"testing"
)

func TestHeapify(t *testing.T) {
	tests := []struct {
		name string
		in   []int
	}{
		{"пустой", nil},
		{"один", []int{1}},
		{"два", []int{2, 1}},
		{"по убыванию", []int{9, 8, 7, 6, 5, 4, 3, 2, 1}},
		{"уже куча", []int{1, 2, 3, 4, 5}},
		{"повторы", []int{5, 1, 5, 1, 5, 1}},
		{"края int", []int{math.MaxInt, 0, math.MinInt, -1}},
		// минимум в последнем листе должен подняться до корня
		{"минимум в конце", []int{10, 20, 30, 40, 50, 60, 70, -1}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := slices.Clone(tt.in)
			Heapify(a)
			if !isMinHeap(a) {
				t.Errorf("Heapify(%v) = %v: не куча", tt.in, a)
			}
			if !slices.Equal(slices.Sorted(slices.Values(a)), slices.Sorted(slices.Values(tt.in))) {
				t.Errorf("Heapify(%v) = %v: потеряны или добавлены элементы", tt.in, a)
			}
		})
	}
}

// TestHeapifyLayout: просеивание вниз от последнего узла с детьми к корню
// для различных значений даёт ровно одну раскладку. Так ловятся и
// сортировка (отсортированный слайс тоже куча), и построение вставками
// с просеиванием вверх за O(n log n).
func TestHeapifyLayout(t *testing.T) {
	in := []int{9, 8, 7, 6, 5, 4, 3, 2, 1}
	want := []int{1, 2, 3, 6, 5, 4, 7, 8, 9}
	a := slices.Clone(in)
	Heapify(a)
	if !slices.Equal(a, want) {
		t.Errorf("Heapify(%v) = %v, want %v (просеивание вниз снизу вверх)", in, a, want)
	}
}

func TestHeapifyBig(t *testing.T) {
	const n = 1_000_000
	r := rand.New(rand.NewPCG(15, 16))
	a := make([]int, n)
	for i := range a {
		a[i] = r.IntN(1000) // много повторов
	}
	sum := 0
	for _, v := range a {
		sum += v
	}
	within(t, func() { Heapify(a) })
	got := 0
	for _, v := range a {
		got += v
	}
	if !isMinHeap(a) || got != sum {
		t.Errorf("Heapify(%d случайных): куча %v, сумма сохранилась %v", n, isMinHeap(a), got == sum)
	}
}
