package l11heap

import (
	"math"
	"math/rand/v2"
	"slices"
	"testing"
)

func TestMinHeapEmpty(t *testing.T) {
	var h MinHeap
	if n := h.Len(); n != 0 {
		t.Errorf("Len() пустой кучи = %d, want 0", n)
	}
	if v, ok := h.Pop(); ok || v != 0 {
		t.Errorf("Pop() пустой кучи = (%d, %v), want (0, false)", v, ok)
	}
	if v, ok := h.Peek(); ok || v != 0 {
		t.Errorf("Peek() пустой кучи = (%d, %v), want (0, false)", v, ok)
	}
	h.Push(7)
	h.Pop()
	if v, ok := h.Pop(); ok {
		t.Errorf("Pop() после опустошения = (%d, %v), want (0, false)", v, ok)
	}
}

func TestMinHeapOrder(t *testing.T) {
	tests := []struct {
		name string
		in   []int
	}{
		{"один элемент", []int{5}},
		{"по убыванию", []int{5, 4, 3, 2, 1}},
		{"по возрастанию", []int{1, 2, 3, 4, 5}},
		{"повторы", []int{3, 1, 3, 1, 2, 2}},
		{"отрицательные и края int", []int{0, -5, math.MaxInt, math.MinInt, 7, -5}},
		{"много одинаковых", []int{4, 4, 4, 4, 4, 4, 4}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var h MinHeap
			for _, v := range tt.in {
				h.Push(v)
			}
			if h.Len() != len(tt.in) {
				t.Fatalf("Len() после %d Push = %d", len(tt.in), h.Len())
			}
			want := slices.Sorted(slices.Values(tt.in))
			var got []int
			for h.Len() > 0 {
				top, _ := h.Peek()
				v, ok := h.Pop()
				if !ok || v != top {
					t.Fatalf("Peek() = %d, а Pop() = (%d, %v)", top, v, ok)
				}
				got = append(got, v)
			}
			if !slices.Equal(got, want) {
				t.Errorf("Push %v, затем Pop до конца = %v, want %v", tt.in, got, want)
			}
		})
	}
}

func TestMinHeapMixed(t *testing.T) {
	// вперемешку Push и Pop, сверяем с медленной, но очевидной моделью
	r := rand.New(rand.NewPCG(11, 12))
	var h MinHeap
	var model []int
	for i := range 5000 {
		if r.IntN(3) > 0 || len(model) == 0 {
			v := r.IntN(100) - 50
			h.Push(v)
			model = append(model, v)
			continue
		}
		m := slices.Min(model)
		model = slices.Delete(model, slices.Index(model, m), slices.Index(model, m)+1)
		if v, ok := h.Pop(); !ok || v != m {
			t.Fatalf("шаг %d: Pop() = (%d, %v), want (%d, true)", i, v, ok, m)
		}
		if h.Len() != len(model) {
			t.Fatalf("шаг %d: Len() = %d, want %d", i, h.Len(), len(model))
		}
	}
}

func TestMinHeapBig(t *testing.T) {
	const n = 1_000_000
	r := rand.New(rand.NewPCG(13, 14))
	in := make([]int, n)
	for i := range in {
		in[i] = r.IntN(1 << 40)
	}
	var got []int
	within(t, func() {
		var h MinHeap
		for _, v := range in {
			h.Push(v)
		}
		for h.Len() > 0 {
			v, _ := h.Pop()
			got = append(got, v)
		}
	})
	if len(got) != n || !slices.IsSorted(got) {
		t.Errorf("MinHeap на %d элементах: вытащено %d, по возрастанию %v", n, len(got), slices.IsSorted(got))
	}
}
