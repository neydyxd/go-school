package l05stackqueue

import (
	"slices"
	"testing"
)

func TestNextGreaterIndex(t *testing.T) {
	tests := []struct {
		name string
		in   []int
		want []int
	}{
		{"пустой", []int{}, []int{}},
		{"один", []int{5}, []int{-1}},
		{"пример", []int{2, 1, 2, 4, 3}, []int{3, 2, 3, -1, -1}},
		{"возрастает", []int{1, 2, 3}, []int{1, 2, -1}},
		{"убывает", []int{3, 2, 1}, []int{-1, -1, -1}},
		{"равные не больше", []int{7, 7, 7}, []int{-1, -1, -1}},
		{"отрицательные", []int{-5, -9, -1, -3}, []int{2, 2, -1, -1}},
		{"больший далеко", []int{5, 1, 2, 3, 4, 6}, []int{5, 2, 3, 4, 5, -1}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NextGreaterIndex(slices.Clone(tt.in))
			if len(got) != len(tt.want) || (len(got) > 0 && !slices.Equal(got, tt.want)) {
				t.Errorf("NextGreaterIndex(%v) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

func TestNextGreaterIndexBig(t *testing.T) {
	const n = 200_000
	nums := make([]int, n)
	for i := range nums {
		nums[i] = n - i // строго убывает: ответ везде -1, перебор идёт до конца
	}
	nums[n-1] = n + 1 // а последний больше всех: ответ везде n-1
	var got []int
	within(t, func() { got = NextGreaterIndex(nums) })
	for i := range n - 1 {
		if got[i] != n-1 {
			t.Fatalf("NextGreaterIndex(большой вход)[%d] = %d, want %d", i, got[i], n-1)
		}
	}
	if got[n-1] != -1 {
		t.Fatalf("NextGreaterIndex(большой вход)[%d] = %d, want -1", n-1, got[n-1])
	}
}
