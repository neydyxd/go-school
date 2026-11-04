package l02twopointers

import (
	"slices"
	"testing"
)

func TestRangeSums(t *testing.T) {
	tests := []struct {
		name    string
		nums    []int
		queries [][2]int
		want    []int
	}{
		{"без запросов", []int{1, 2}, nil, []int{}},
		{"классика", []int{-2, 0, 3, -5, 2, -1}, [][2]int{{0, 2}, {2, 5}, {0, 5}}, []int{1, -1, -3}},
		{"один элемент", []int{7}, [][2]int{{0, 0}}, []int{7}},
		{"точечные запросы", []int{4, 5, 6}, [][2]int{{0, 0}, {1, 1}, {2, 2}}, []int{4, 5, 6}},
		{"весь массив и хвост", []int{1, 2, 3, 4}, [][2]int{{0, 3}, {3, 3}, {1, 3}}, []int{10, 4, 9}},
		{"большие числа", []int{1_000_000_000_000, 1_000_000_000_000, -1}, [][2]int{{0, 2}}, []int{1_999_999_999_999}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RangeSums(tt.nums, tt.queries)
			if len(got) != len(tt.want) || (len(got) > 0 && !slices.Equal(got, tt.want)) {
				t.Errorf("RangeSums(%v, %v) = %v, want %v", tt.nums, tt.queries, got, tt.want)
			}
		})
	}
}

func TestRangeSumsBig(t *testing.T) {
	const n = 200_000
	nums := make([]int, n)
	for i := range nums {
		nums[i] = i%7 - 3
	}
	queries := make([][2]int, n)
	for i := range queries {
		queries[i] = [2]int{i % 10, n - 1 - i%10}
	}
	var got []int
	within(t, func() { got = RangeSums(nums, queries) })
	if len(got) != n {
		t.Fatalf("RangeSums(большой вход) вернула %d ответов, want %d", len(got), n)
	}
	for _, i := range []int{0, 1, 9, n / 2, n - 1} {
		want := 0
		for _, v := range nums[queries[i][0] : queries[i][1]+1] {
			want += v
		}
		if got[i] != want {
			t.Errorf("RangeSums(большой вход): ответ %d на %v = %d, want %d", i, queries[i], got[i], want)
		}
	}
}
