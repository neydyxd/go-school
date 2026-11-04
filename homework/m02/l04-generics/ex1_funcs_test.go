package l04generics

import (
	"slices"
	"strconv"
	"strings"
	"testing"
)

func TestMap(t *testing.T) {
	t.Run("int в string", func(t *testing.T) {
		got := Map([]int{1, 2, 3}, strconv.Itoa)
		if want := []string{"1", "2", "3"}; !slices.Equal(got, want) {
			t.Errorf("Map([1 2 3], Itoa) = %q, want %q", got, want)
		}
	})
	t.Run("строки в длины", func(t *testing.T) {
		got := Map([]string{"go", "", "привет"}, func(s string) int { return len([]rune(s)) })
		if want := []int{2, 0, 6}; !slices.Equal(got, want) {
			t.Errorf("Map(..., runeLen) = %v, want %v", got, want)
		}
	})
	t.Run("пустой вход", func(t *testing.T) {
		if got := Map(nil, func(int) int { return 1 }); len(got) != 0 {
			t.Errorf("Map(nil) = %v, want пусто", got)
		}
	})
}

func TestFilter(t *testing.T) {
	even := func(n int) bool { return n%2 == 0 }
	tests := []struct {
		name string
		in   []int
		want []int
	}{
		{"чётные", []int{1, 2, 3, 4, 6}, []int{2, 4, 6}},
		{"ничего не прошло", []int{1, 3}, nil},
		{"пусто", nil, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			orig := slices.Clone(tt.in)
			got := Filter(tt.in, even)
			if !slices.Equal(got, tt.want) {
				t.Errorf("Filter(%v, even) = %v, want %v", tt.in, got, tt.want)
			}
			if !slices.Equal(tt.in, orig) {
				t.Errorf("Filter испортил входной слайс: было %v, стало %v", orig, tt.in)
			}
		})
	}
	t.Run("со строками", func(t *testing.T) {
		got := Filter([]string{"go", "rust", "gopher"}, func(s string) bool { return strings.HasPrefix(s, "go") })
		if want := []string{"go", "gopher"}; !slices.Equal(got, want) {
			t.Errorf("Filter(prefix go) = %q, want %q", got, want)
		}
	})
}

func TestReduce(t *testing.T) {
	sum := Reduce([]int{1, 2, 3, 4}, 0, func(a, n int) int { return a + n })
	if sum != 10 {
		t.Errorf("Reduce(sum) = %d, want 10", sum)
	}
	joined := Reduce([]int{1, 2, 3}, "", func(a string, n int) string { return a + strconv.Itoa(n) })
	if joined != "123" {
		t.Errorf("Reduce(concat) = %q, want %q (порядок слева направо)", joined, "123")
	}
	if got := Reduce(nil, 7, func(a, n int) int { return a * n }); got != 7 {
		t.Errorf("Reduce(nil, 7) = %d, want 7", got)
	}
}
