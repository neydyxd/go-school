package l05slices

import (
	"slices"
	"testing"
)

func TestChunk(t *testing.T) {
	tests := []struct {
		name string
		in   []int
		size int
		want [][]int
	}{
		{"ровно делится", []int{1, 2, 3, 4}, 2, [][]int{{1, 2}, {3, 4}}},
		{"хвост короче", []int{1, 2, 3, 4, 5}, 2, [][]int{{1, 2}, {3, 4}, {5}}},
		{"size больше длины", []int{1, 2}, 5, [][]int{{1, 2}}},
		{"size 1", []int{7, 8}, 1, [][]int{{7}, {8}}},
		{"size 0", []int{1, 2}, 0, nil},
		{"отрицательный size", []int{1, 2}, -3, nil},
		{"пустой вход", []int{}, 2, nil},
		{"nil", nil, 2, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Chunk(tt.in, tt.size)
			if !slices.EqualFunc(got, tt.want, slices.Equal) || (got == nil) != (tt.want == nil) {
				t.Errorf("Chunk(%v, %d) = %v, want %v", tt.in, tt.size, got, tt.want)
			}
		})
	}
}

func TestChunkAppendIsSafe(t *testing.T) {
	in := []int{1, 2, 3, 4, 5, 6}
	parts := Chunk(in, 2)
	if len(parts) != 3 {
		t.Fatalf("Chunk(%v, 2) вернул %d частей, want 3", in, len(parts))
	}
	parts[0] = append(parts[0], 100)
	if !slices.Equal(parts[1], []int{3, 4}) {
		t.Errorf("append к первой части испортил вторую: %v, want [3 4]", parts[1])
	}
	if !slices.Equal(in, []int{1, 2, 3, 4, 5, 6}) {
		t.Errorf("append к первой части испортил исходный слайс: %v", in)
	}
}
