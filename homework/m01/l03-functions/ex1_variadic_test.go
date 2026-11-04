package l03functions

import (
	"math"
	"testing"
)

func TestSum(t *testing.T) {
	tests := []struct {
		name string
		in   []int
		want int
	}{
		{"без аргументов", nil, 0},
		{"один", []int{7}, 7},
		{"несколько", []int{1, 2, 3, 4}, 10},
		{"с отрицательными", []int{5, -10, 2}, -3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Sum(tt.in...); got != tt.want {
				t.Errorf("Sum(%v...) = %d, want %d", tt.in, got, tt.want)
			}
		})
	}
}

func TestMinMax(t *testing.T) {
	tests := []struct {
		name           string
		first          int
		rest           []int
		wantLo, wantHi int
	}{
		{"один аргумент", 5, nil, 5, 5},
		{"первый наибольший", 9, []int{1, 4}, 1, 9},
		{"первый наименьший", -3, []int{8, 0}, -3, 8},
		{"все отрицательные", -1, []int{-7, -2}, -7, -1},
		{"все положительные", 3, []int{10, 6}, 3, 10},
		{"крайние значения", 0, []int{math.MaxInt, math.MinInt}, math.MinInt, math.MaxInt},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lo, hi := MinMax(tt.first, tt.rest...)
			if lo != tt.wantLo || hi != tt.wantHi {
				t.Errorf("MinMax(%d, %v...) = (%d, %d), want (%d, %d)", tt.first, tt.rest, lo, hi, tt.wantLo, tt.wantHi)
			}
		})
	}
}
