package l12dpgreedy

import (
	"math"
	"testing"
)

func TestLengthOfLIS(t *testing.T) {
	tests := []struct {
		name string
		in   []int
		want int
	}{
		{"пустой", nil, 0},
		{"один", []int{5}, 1},
		{"LeetCode 1", []int{10, 9, 2, 5, 3, 7, 101, 18}, 4},
		{"LeetCode 2", []int{0, 1, 0, 3, 2, 3}, 4},
		{"все равные", []int{7, 7, 7, 7}, 1},
		{"по убыванию", []int{5, 4, 3, 2, 1}, 1},
		{"по возрастанию", []int{1, 2, 3, 4, 5}, 5},
		{"повторы внутри", []int{1, 3, 3, 4, 4, 5}, 4},
		{"края int", []int{math.MinInt, 0, math.MaxInt}, 3},
		// жадно брать «следующий больший» неверно
		{"жадный ошибается", []int{1, 100, 2, 3, 4}, 4},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := LengthOfLIS(tt.in); got != tt.want {
				t.Errorf("LengthOfLIS(%v) = %d, want %d", tt.in, got, tt.want)
			}
		})
	}
}

func TestLengthOfLISBig(t *testing.T) {
	// 200 000 чисел: возрастающая «пила» 0..999 повторяется 200 раз,
	// каждый следующий зубец на 1 выше, ответ 1000 + 199
	const teeth, width = 200, 1000
	nums := make([]int, 0, teeth*width)
	for k := range teeth {
		for i := range width {
			nums = append(nums, k+i)
		}
	}
	var got int
	within(t, func() { got = LengthOfLIS(nums) })
	if want := width + teeth - 1; got != want {
		t.Errorf("LengthOfLIS(пила из %d чисел) = %d, want %d", len(nums), got, want)
	}
}
