package l06binarysearch

import "testing"

func TestSearch(t *testing.T) {
	noImports(t, "ex1_search.go", "sort", "slices")
	tests := []struct {
		name   string
		nums   []int
		target int
		want   int
	}{
		{"пустой", nil, 5, -1},
		{"один, есть", []int{5}, 5, 0},
		{"один, нет", []int{5}, 4, -1},
		{"в середине", []int{-1, 0, 3, 5, 9, 12}, 9, 4},
		{"нет в середине", []int{-1, 0, 3, 5, 9, 12}, 2, -1},
		{"первый", []int{-1, 0, 3, 5, 9, 12}, -1, 0},
		{"последний", []int{-1, 0, 3, 5, 9, 12}, 12, 5},
		{"меньше всех", []int{1, 2, 3}, -100, -1},
		{"больше всех", []int{1, 2, 3}, 100, -1},
		{"два элемента, правый", []int{1, 2}, 2, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got int
			within(t, func() { got = Search(tt.nums, tt.target) })
			if got != tt.want {
				t.Errorf("Search(%v, %d) = %d, want %d", tt.nums, tt.target, got, tt.want)
			}
		})
	}
}

func TestSearchBig(t *testing.T) {
	const n = 200_000
	nums := make([]int, n)
	for i := range nums {
		nums[i] = 2 * i // чётные: нечётные запросы промахиваются
	}
	badQ, badGot, badWant := 0, 0, 0
	failed := false
	within(t, func() {
		for q := -1; q <= 2*n; q++ {
			want := -1
			if q >= 0 && q%2 == 0 && q < 2*n {
				want = q / 2
			}
			if got := Search(nums, q); got != want {
				badQ, badGot, badWant, failed = q, got, want, true
				return
			}
		}
	})
	if failed {
		t.Fatalf("Search(большой вход, %d) = %d, want %d", badQ, badGot, badWant)
	}
}
