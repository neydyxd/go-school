package l03internalsperf

import (
	"math"
	"testing"
)

func TestJoinIDs(t *testing.T) {
	tests := []struct {
		name string
		ids  []int64
		sep  string
		want string
	}{
		{"nil", nil, ", ", ""},
		{"пустой", []int64{}, ",", ""},
		{"один", []int64{42}, ", ", "42"},
		{"несколько", []int64{1, 22, -3}, ", ", "1, 22, -3"},
		{"пустой разделитель", []int64{1, 2, 3}, "", "123"},
		{"нули", []int64{0, 0}, "|", "0|0"},
		{"границы int64", []int64{math.MinInt64, math.MaxInt64}, " ", "-9223372036854775808 9223372036854775807"},
		{"юникод в разделителе", []int64{7, 8}, " → ", "7 → 8"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := JoinIDs(tt.ids, tt.sep); got != tt.want {
				t.Errorf("JoinIDs(%v, %q) = %q, want %q", tt.ids, tt.sep, got, tt.want)
			}
		})
	}
}

// benchIDs это тысяча идентификаторов разной длины, включая отрицательные.
var benchIDs = func() []int64 {
	ids := make([]int64, 1000)
	for i := range ids {
		ids[i] = int64(i*i*7919) - 500_000
	}
	return ids
}()

func TestJoinIDsAllocs(t *testing.T) {
	if raceEnabled {
		t.Skip("под -race счётчик аллокаций не показателен")
	}
	allocs := testing.AllocsPerRun(20, func() { JoinIDs(benchIDs, ", ") })
	if allocs > 1 {
		t.Errorf("JoinIDs(1000 чисел): %.0f аллокаций на вызов, want не больше 1 (только итоговая строка)", allocs)
	}
}

func BenchmarkJoinIDs(b *testing.B) {
	for b.Loop() {
		JoinIDs(benchIDs, ", ")
	}
}
