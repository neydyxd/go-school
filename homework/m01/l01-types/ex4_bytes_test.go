package l01types

import (
	"math"
	"testing"
)

func TestHumanBytes(t *testing.T) {
	tests := []struct {
		name string
		in   int64
		want string
	}{
		{"ноль", 0, "0 B"},
		{"байты", 512, "512 B"},
		{"граница до KB", 1023, "1023 B"},
		{"ровно KB", 1024, "1.0 KB"},
		{"полтора KB", 1536, "1.5 KB"},
		{"ровно MB", 1 << 20, "1.0 MB"},
		{"ровно GB", 1 << 30, "1.0 GB"},
		{"2.5 GB", 5 << 29, "2.5 GB"},
		{"ровно TB", 1 << 40, "1.0 TB"},
		{"ровно PB", 1 << 50, "1.0 PB"},
		{"ровно EB", 1 << 60, "1.0 EB"},
		{"максимум int64", math.MaxInt64, "8.0 EB"},
		{"отрицательное", -2048, "-2.0 KB"},
		{"отрицательные байты", -5, "-5 B"},
		{"минимум int64", math.MinInt64, "-8.0 EB"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HumanBytes(tt.in); got != tt.want {
				t.Errorf("HumanBytes(%d) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}
