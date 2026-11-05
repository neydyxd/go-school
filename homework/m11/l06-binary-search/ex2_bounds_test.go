package l06binarysearch

import "testing"

func TestBounds(t *testing.T) {
	noImports(t, "ex2_bounds.go", "sort", "slices")
	tests := []struct {
		name         string
		nums         []int
		x            int
		lower, upper int
	}{
		{"пустой", nil, 3, 0, 0},
		{"пример", []int{1, 2, 2, 2, 5}, 2, 1, 4},
		{"нет, между", []int{1, 2, 2, 2, 5}, 3, 4, 4},
		{"меньше всех", []int{1, 2, 2, 2, 5}, 0, 0, 0},
		{"больше всех", []int{1, 2, 2, 2, 5}, 9, 5, 5},
		{"равен последнему", []int{1, 2, 2, 2, 5}, 5, 4, 5},
		{"равен первому", []int{1, 2, 2, 2, 5}, 1, 0, 1},
		{"все равны", []int{7, 7, 7}, 7, 0, 3},
		{"один, меньше", []int{4}, 3, 0, 0},
		{"один, равен", []int{4}, 4, 0, 1},
		{"один, больше", []int{4}, 5, 1, 1},
		{"отрицательные", []int{-5, -3, -3, 0}, -3, 1, 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var lo, up int
			within(t, func() { lo, up = LowerBound(tt.nums, tt.x), UpperBound(tt.nums, tt.x) })
			if lo != tt.lower {
				t.Errorf("LowerBound(%v, %d) = %d, want %d", tt.nums, tt.x, lo, tt.lower)
			}
			if up != tt.upper {
				t.Errorf("UpperBound(%v, %d) = %d, want %d", tt.nums, tt.x, up, tt.upper)
			}
		})
	}
}

func TestBoundsBig(t *testing.T) {
	const n = 200_000
	nums := make([]int, n)
	for i := range nums {
		nums[i] = i / 4 * 10 // каждое значение 4 раза: 0 0 0 0 10 10 10 10 ...
	}
	failed, msg := false, ""
	within(t, func() {
		for v := -5; v < n/4*10+5; v += 5 {
			wantL := min(max((v+9)/10*4, 0), n)
			wantU := min(max((v/10+1)*4, 0), n)
			if v < 0 {
				wantL, wantU = 0, 0
			}
			if got := LowerBound(nums, v); got != wantL {
				failed, msg = true, "LowerBound"
				return
			}
			if got := UpperBound(nums, v); got != wantU {
				failed, msg = true, "UpperBound"
				return
			}
		}
	})
	if failed {
		t.Fatalf("%s вернул неверный ответ на большом входе", msg)
	}
}
