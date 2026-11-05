package l06binarysearch

import "testing"

func TestSearchRotated(t *testing.T) {
	noImports(t, "ex5_search_rotated.go", "sort", "slices")
	tests := []struct {
		name   string
		nums   []int
		target int
		want   int
	}{
		{"пустой", nil, 1, -1},
		{"пример", []int{4, 5, 6, 7, 0, 1, 2}, 0, 4},
		{"нет", []int{4, 5, 6, 7, 0, 1, 2}, 3, -1},
		{"в левой части", []int{4, 5, 6, 7, 0, 1, 2}, 5, 1},
		{"последний", []int{4, 5, 6, 7, 0, 1, 2}, 2, 6},
		{"первый", []int{4, 5, 6, 7, 0, 1, 2}, 4, 0},
		{"без сдвига", []int{1, 2, 3, 4, 5}, 4, 3},
		{"один", []int{1}, 1, 0},
		{"один, нет", []int{1}, 0, -1},
		{"два, сдвинут", []int{3, 1}, 1, 1},
		{"два, сдвинут, первый", []int{3, 1}, 3, 0},
		{"сдвиг на один", []int{5, 1, 2, 3, 4}, 5, 0},
		{"сдвиг почти полный", []int{2, 3, 4, 5, 1}, 1, 4},
		{"отрицательные", []int{-1, 3, -9, -5}, -5, 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got int
			within(t, func() { got = SearchRotated(tt.nums, tt.target) })
			if got != tt.want {
				t.Errorf("SearchRotated(%v, %d) = %d, want %d", tt.nums, tt.target, got, tt.want)
			}
		})
	}
}

func TestSearchRotatedBig(t *testing.T) {
	const n, shift = 200_000, 123_457
	nums := make([]int, n)
	for i := range nums {
		nums[(i+shift)%n] = 3 * i // nums[shift] = 0 это минимум
	}
	failed, q, got, want := false, 0, 0, 0
	within(t, func() {
		for v := -1; v < 3*n; v++ {
			w := -1
			if v >= 0 && v%3 == 0 {
				w = (v/3 + shift) % n
			}
			if g := SearchRotated(nums, v); g != w {
				failed, q, got, want = true, v, g, w
				return
			}
		}
	})
	if failed {
		t.Fatalf("SearchRotated(большой вход, %d) = %d, want %d", q, got, want)
	}
}
