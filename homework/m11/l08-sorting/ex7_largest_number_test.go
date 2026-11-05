package l08sorting

import (
	"slices"
	"strings"
	"testing"
)

func TestLargestNumber(t *testing.T) {
	tests := []struct {
		name string
		in   []int
		want string
	}{
		{"пустой", nil, ""},
		{"одно", []int{7}, "7"},
		{"два", []int{10, 2}, "210"},
		{"пример", []int{3, 30, 34, 5, 9}, "9534330"},
		{"все нули", []int{0, 0, 0}, "0"},
		{"один ноль", []int{0}, "0"},
		{"ноль и число", []int{0, 1}, "10"},
		{"общий префикс", []int{121, 12}, "12121"},
		{"общий префикс 2", []int{824, 8247}, "8248247"},
		{"одинаковые", []int{5, 5, 5}, "555"},
		{"длинное", []int{999999999, 999999998, 999999997}, "999999999999999998999999997"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in := slices.Clone(tt.in)
			var got string
			within(t, func() { got = LargestNumber(in) })
			if got != tt.want {
				t.Errorf("LargestNumber(%v) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestLargestNumberBig(t *testing.T) {
	const n = 100_000
	in := make([]int, n)
	for i := range in {
		in[i] = i % 10 // по 10 000 каждой цифры
	}
	var got string
	within(t, func() { got = LargestNumber(in) })
	var want strings.Builder
	for d := 9; d >= 0; d-- {
		want.WriteString(strings.Repeat(string(rune('0'+d)), n/10))
	}
	if got != want.String() {
		t.Fatalf("LargestNumber(%d цифр) вернул неверную строку длины %d", n, len(got))
	}
}
