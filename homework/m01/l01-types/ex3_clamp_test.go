package l01types

import (
	"math"
	"testing"
)

func TestClampUint8(t *testing.T) {
	tests := []struct {
		name string
		in   int
		want uint8
	}{
		{"ноль", 0, 0},
		{"середина", 128, 128},
		{"верхняя граница", 255, 255},
		{"чуть выше", 256, 255},
		{"300 не превращается в 44", 300, 255},
		{"отрицательное", -1, 0},
		{"очень большое", math.MaxInt, 255},
		{"очень маленькое", math.MinInt, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ClampUint8(tt.in); got != tt.want {
				t.Errorf("ClampUint8(%d) = %d, want %d", tt.in, got, tt.want)
			}
		})
	}
}
