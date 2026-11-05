package l06binarysearch

import "testing"

func TestSearchRange(t *testing.T) {
	noImports(t, "ex3_search_range.go", "sort", "slices")
	tests := []struct {
		name        string
		nums        []int
		target      int
		first, last int
	}{
		{"пустой", nil, 0, -1, -1},
		{"пример", []int{5, 7, 7, 8, 8, 10}, 8, 3, 4},
		{"нет", []int{5, 7, 7, 8, 8, 10}, 6, -1, -1},
		{"одно вхождение", []int{5, 7, 7, 8, 8, 10}, 10, 5, 5},
		{"в начале", []int{5, 7, 7, 8, 8, 10}, 5, 0, 0},
		{"больше всех", []int{5, 7, 7, 8, 8, 10}, 11, -1, -1},
		{"меньше всех", []int{5, 7, 7, 8, 8, 10}, 1, -1, -1},
		{"один элемент", []int{1}, 1, 0, 0},
		{"все равны", []int{2, 2, 2, 2}, 2, 0, 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var f, l int
			within(t, func() { f, l = SearchRange(tt.nums, tt.target) })
			if f != tt.first || l != tt.last {
				t.Errorf("SearchRange(%v, %d) = %d, %d, want %d, %d", tt.nums, tt.target, f, l, tt.first, tt.last)
			}
		})
	}
}

func TestSearchRangeBig(t *testing.T) {
	// все элементы равны: «найти любое и расширяться» проходит весь слайс
	// на каждом запросе
	const n, q = 200_000, 100_000
	nums := make([]int, n)
	for i := range nums {
		nums[i] = 42
	}
	nums[0], nums[n-1] = 41, 43
	var f, l int
	within(t, func() {
		for range q {
			f, l = SearchRange(nums, 42)
		}
	})
	if f != 1 || l != n-2 {
		t.Fatalf("SearchRange(большой вход, 42) = %d, %d, want 1, %d", f, l, n-2)
	}
}
