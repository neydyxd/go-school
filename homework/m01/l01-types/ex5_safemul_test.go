package l01types

import (
	"math"
	"testing"
)

func TestSafeMul(t *testing.T) {
	tests := []struct {
		name   string
		a, b   int64
		want   int64
		wantOK bool
	}{
		{"обычные", 6, 7, 42, true},
		{"на ноль", math.MaxInt64, 0, 0, true},
		{"ноль на минимум", 0, math.MinInt64, 0, true},
		{"отрицательные", -4, -5, 20, true},
		{"разные знаки", -4, 5, -20, true},
		{"максимум на 1", math.MaxInt64, 1, math.MaxInt64, true},
		{"минимум на 1", math.MinInt64, 1, math.MinInt64, true},
		{"максимум на -1", math.MaxInt64, -1, -math.MaxInt64, true},
		{"переполнение", math.MaxInt64, 2, 0, false},
		{"переполнение в минус", math.MinInt64 / 2, 3, 0, false},
		{"квадрат большого", 1 << 32, 1 << 32, 0, false},
		{"минимум на -1", math.MinInt64, -1, 0, false},
		{"-1 на минимум", -1, math.MinInt64, 0, false},
		{"ровно минимум", math.MinInt64 / 2, 2, math.MinInt64, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := SafeMul(tt.a, tt.b)
			if got != tt.want || ok != tt.wantOK {
				t.Errorf("SafeMul(%d, %d) = (%d, %v), want (%d, %v)", tt.a, tt.b, got, ok, tt.want, tt.wantOK)
			}
		})
	}
}
