package l11heap

import (
	"math/rand/v2"
	"slices"
	"testing"
)

func TestMedianFinder(t *testing.T) {
	tests := []struct {
		name string
		in   []int
		want []float64 // медиана после каждого AddNum
	}{
		{"LeetCode", []int{1, 2, 3}, []float64{1, 1.5, 2}},
		{"по убыванию", []int{5, 4, 3, 2, 1}, []float64{5, 4.5, 4, 3.5, 3}},
		{"отрицательные", []int{-1, -2, -3, -4}, []float64{-1, -1.5, -2, -2.5}},
		{"повторы", []int{2, 2, 2, 1}, []float64{2, 2, 2, 2}},
		{"скачки", []int{6, 10, 2, 6, 5, 0}, []float64{6, 8, 6, 6, 6, 5.5}},
		{"нечётная сумма средних", []int{1, 4}, []float64{1, 2.5}},
		{"большие числа", []int{1_000_000_000, 999_999_999}, []float64{1e9, 999_999_999.5}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewMedianFinder()
			for i, v := range tt.in {
				m.AddNum(v)
				if got := m.FindMedian(); got != tt.want[i] {
					t.Errorf("после AddNum %v FindMedian() = %v, want %v", tt.in[:i+1], got, tt.want[i])
				}
			}
		})
	}
	t.Run("пустой поток", func(t *testing.T) {
		if got := NewMedianFinder().FindMedian(); got != 0 {
			t.Errorf("FindMedian() пустого потока = %v, want 0", got)
		}
	})
	t.Run("независимые экземпляры", func(t *testing.T) {
		a, b := NewMedianFinder(), NewMedianFinder()
		a.AddNum(1)
		b.AddNum(100)
		if a.FindMedian() != 1 || b.FindMedian() != 100 {
			t.Errorf("два MedianFinder мешают друг другу: %v и %v, want 1 и 100", a.FindMedian(), b.FindMedian())
		}
	})
}

func TestMedianFinderBig(t *testing.T) {
	const n = 1_000_000
	r := rand.New(rand.NewPCG(25, 26))
	in := make([]int, n)
	for i := range in {
		in[i] = r.IntN(2_000_000_001) - 1_000_000_000
	}
	checkAt := map[int]bool{0: true, 1: true, 1000: true, n/2 - 1: true, n - 1: true}
	got := map[int]float64{}
	within(t, func() {
		m := NewMedianFinder()
		for i, v := range in {
			m.AddNum(v)
			med := m.FindMedian()
			if checkAt[i] {
				got[i] = med
			}
		}
	})
	for i := range checkAt {
		s := slices.Sorted(slices.Values(in[:i+1]))
		c := len(s)
		want := float64(s[c/2])
		if c%2 == 0 {
			want = float64(s[c/2-1]+s[c/2]) / 2
		}
		if got[i] != want {
			t.Errorf("FindMedian() после %d чисел = %v, want %v", i+1, got[i], want)
		}
	}
}
