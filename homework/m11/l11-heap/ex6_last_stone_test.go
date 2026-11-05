package l11heap

import (
	"slices"
	"testing"
)

func TestLastStoneWeight(t *testing.T) {
	tests := []struct {
		name string
		in   []int
		want int
	}{
		{"пустой", nil, 0},
		{"один камень", []int{1}, 1},
		{"LeetCode", []int{2, 7, 4, 1, 8, 1}, 1},
		{"два равных", []int{5, 5}, 0},
		{"два разных", []int{3, 10}, 7},
		{"остаток снова в игре", []int{10, 4, 3}, 3},
		{"все равные, чётно", []int{2, 2, 2, 2}, 0},
		{"все равные, нечётно", []int{2, 2, 2}, 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := LastStoneWeight(slices.Clone(tt.in)); got != tt.want {
				t.Errorf("LastStoneWeight(%v) = %d, want %d", tt.in, got, tt.want)
			}
		})
	}
}

func TestLastStoneWeightBig(t *testing.T) {
	// камни 1, 2, ..., n: каждый шаг сталкивает два соседних по весу,
	// остатки (единицы) копятся и долго живут в куче
	const n = 200_000
	stones := make([]int, n)
	for i := range stones {
		stones[i] = i + 1
	}
	var got int
	within(t, func() { got = LastStoneWeight(stones) })
	// сумма 1..n чётна при n кратном 4, и каждый шаг сохраняет чётность
	// суммы, поэтому ответ чётный; для этого входа он равен 0
	if got != 0 {
		t.Errorf("LastStoneWeight(1..%d) = %d, want 0", n, got)
	}
}
