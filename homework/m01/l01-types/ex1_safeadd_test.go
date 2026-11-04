package l01types

import (
	"math"
	"testing"
)

func TestSafeAdd(t *testing.T) {
	tests := []struct {
		name   string
		a, b   int64
		want   int64
		wantOK bool
	}{
		{"обычные числа", 2, 3, 5, true},
		{"с отрицательным", -10, 3, -7, true},
		{"ноль", 0, 0, 0, true},
		{"ровно максимум", math.MaxInt64 - 1, 1, math.MaxInt64, true},
		{"ровно минимум", math.MinInt64 + 1, -1, math.MinInt64, true},
		{"переполнение вверх", math.MaxInt64, 1, 0, false},
		{"переполнение вниз", math.MinInt64, -1, 0, false},
		{"два больших", math.MaxInt64, math.MaxInt64, 0, false},
		{"два малых", math.MinInt64, math.MinInt64, 0, false},
		{"разные знаки не переполняются", math.MaxInt64, math.MinInt64, -1, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := SafeAdd(tt.a, tt.b)
			if got != tt.want || ok != tt.wantOK {
				t.Errorf("SafeAdd(%d, %d) = (%d, %v), want (%d, %v)", tt.a, tt.b, got, ok, tt.want, tt.wantOK)
			}
		})
	}
}
