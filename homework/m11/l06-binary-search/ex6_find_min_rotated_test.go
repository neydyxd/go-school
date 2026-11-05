package l06binarysearch

import "testing"

func TestFindMinRotated(t *testing.T) {
	tests := []struct {
		name string
		nums []int
		want int
	}{
		{"один", []int{7}, 7},
		{"пример", []int{3, 4, 5, 1, 2}, 1},
		{"leetcode", []int{4, 5, 6, 7, 0, 1, 2}, 0},
		{"без сдвига", []int{11, 13, 15, 17}, 11},
		{"два, сдвинут", []int{2, 1}, 1},
		{"два, без сдвига", []int{1, 2}, 1},
		{"минимум в конце", []int{2, 3, 4, 5, 1}, 1},
		{"минимум вторым", []int{5, 1, 2, 3, 4}, 1},
		{"отрицательные", []int{0, 3, -7, -4}, -7},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got int
			within(t, func() { got = FindMinRotated(tt.nums) })
			if got != tt.want {
				t.Errorf("FindMinRotated(%v) = %d, want %d", tt.nums, got, tt.want)
			}
		})
	}
}

func TestFindMinRotatedBig(t *testing.T) {
	const n = 200_000
	failed, shift, got := false, 0, 0
	within(t, func() {
		nums := make([]int, n)
		for _, sh := range []int{0, 1, n / 3, n / 2, n - 1} {
			for i := range nums {
				nums[(i+sh)%n] = i - 1000
			}
			for range 20_000 {
				if g := FindMinRotated(nums); g != -1000 {
					failed, shift, got = true, sh, g
					return
				}
			}
		}
	})
	if failed {
		t.Fatalf("FindMinRotated(большой вход со сдвигом %d) = %d, want -1000", shift, got)
	}
}
