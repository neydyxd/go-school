package l02controlflow

import (
	"slices"
	"testing"
)

// collect собирает не больше max значений и превращает панику в ошибку теста.
func collect(t *testing.T, limit, max int) (got []int) {
	t.Helper()
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Evens(%d): паника при переборе: %v", limit, r)
		}
	}()
	got = []int{}
	for v := range Evens(limit) {
		if len(got) == max {
			break
		}
		got = append(got, v)
	}
	return got
}

func TestEvens(t *testing.T) {
	tests := []struct {
		name  string
		limit int
		max   int
		want  []int
	}{
		{"ноль", 0, 100, []int{}},
		{"отрицательный", -4, 100, []int{}},
		{"единица", 1, 100, []int{0}},
		{"до семи", 7, 100, []int{0, 2, 4, 6}},
		{"limit не включается", 6, 100, []int{0, 2, 4}},
		{"ранний break", 100, 3, []int{0, 2, 4}},
		{"break на первом", 100, 0, []int{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := collect(t, tt.limit, tt.max); !slices.Equal(got, tt.want) {
				t.Errorf("Evens(%d) с остановкой после %d = %v, want %v", tt.limit, tt.max, got, tt.want)
			}
		})
	}
}
