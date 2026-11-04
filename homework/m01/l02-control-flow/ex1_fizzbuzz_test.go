package l02controlflow

import (
	"slices"
	"testing"
)

func TestFizzBuzz(t *testing.T) {
	tests := []struct {
		name string
		n    int
		want []string
	}{
		{"ноль", 0, []string{}},
		{"отрицательное", -3, []string{}},
		{"одно число", 1, []string{"1"}},
		{"до пяти", 5, []string{"1", "2", "Fizz", "4", "Buzz"}},
		{"до пятнадцати", 15, []string{"1", "2", "Fizz", "4", "Buzz", "Fizz", "7", "8", "Fizz", "Buzz", "11", "Fizz", "13", "14", "FizzBuzz"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FizzBuzz(tt.n)
			if len(got) != len(tt.want) || !slices.Equal(got, tt.want) {
				t.Errorf("FizzBuzz(%d) = %q, want %q", tt.n, got, tt.want)
			}
		})
	}
}
