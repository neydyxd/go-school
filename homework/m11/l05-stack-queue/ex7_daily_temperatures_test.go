package l05stackqueue

import (
	"slices"
	"testing"
)

func TestDailyTemperatures(t *testing.T) {
	tests := []struct {
		name string
		in   []int
		want []int
	}{
		{"пустой", []int{}, []int{}},
		{"один день", []int{20}, []int{0}},
		{"пример", []int{73, 74, 75, 71, 69, 72, 76, 73}, []int{1, 1, 4, 2, 1, 1, 0, 0}},
		{"растёт", []int{30, 40, 50, 60}, []int{1, 1, 1, 0}},
		{"падает", []int{60, 50, 40}, []int{0, 0, 0}},
		{"одинаковые не теплее", []int{10, 10, 10, 11}, []int{3, 2, 1, 0}},
		{"мороз", []int{-20, -25, -10}, []int{2, 1, 0}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DailyTemperatures(slices.Clone(tt.in))
			if len(got) != len(tt.want) || (len(got) > 0 && !slices.Equal(got, tt.want)) {
				t.Errorf("DailyTemperatures(%v) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

func TestDailyTemperaturesBig(t *testing.T) {
	const n = 200_000
	temps := make([]int, n)
	for i := range temps {
		temps[i] = 15 // долгое плато
	}
	temps[n-1] = 16
	var got []int
	within(t, func() { got = DailyTemperatures(temps) })
	for i := range n {
		if got[i] != n-1-i {
			t.Fatalf("DailyTemperatures(плато)[%d] = %d, want %d", i, got[i], n-1-i)
		}
	}
}
