package l02benchfuzz

import (
	"math"
	"strconv"
	"testing"
)

func TestJoinInts(t *testing.T) {
	tests := []struct {
		name string
		nums []int
		sep  string
		want string
	}{
		{"nil", nil, ",", ""},
		{"пустой", []int{}, ",", ""},
		{"одно число", []int{42}, ",", "42"},
		{"несколько", []int{1, -20, 300}, ", ", "1, -20, 300"},
		{"пустой разделитель", []int{1, 2, 3}, "", "123"},
		{"юникодный разделитель", []int{7, 8}, " → ", "7 → 8"},
		{"крайние значения", []int{math.MinInt64, 0, math.MaxInt64}, "|", "-9223372036854775808|0|9223372036854775807"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := JoinInts(tt.nums, tt.sep); got != tt.want {
				t.Errorf("JoinInts(%v, %q) = %q, want %q", tt.nums, tt.sep, got, tt.want)
			}
		})
	}
}

func TestJoinIntsAllocs(t *testing.T) {
	nums := make([]int, 1000)
	for i := range nums {
		nums[i] = i * 7919
	}
	nums[0] = math.MinInt64
	allocs := testing.AllocsPerRun(100, func() {
		JoinInts(nums, ", ")
	})
	if allocs > 1 {
		t.Errorf("JoinInts на 1000 чисел: %.0f аллокаций на вызов, want не больше 1", allocs)
	}
}

// joinNaive это наивная версия для сравнения в бенчмарке.
func joinNaive(nums []int, sep string) string {
	s := ""
	for i, n := range nums {
		if i > 0 {
			s += sep
		}
		s += strconv.Itoa(n)
	}
	return s
}

func benchNums() []int {
	nums := make([]int, 1000)
	for i := range nums {
		nums[i] = i * 7919
	}
	return nums
}

// Запуск: go test -bench=JoinInts -benchmem ./m03/l02-bench-fuzz/
func BenchmarkJoinIntsNaive(b *testing.B) {
	nums := benchNums()
	for b.Loop() {
		joinNaive(nums, ", ")
	}
}

func BenchmarkJoinInts(b *testing.B) {
	nums := benchNums()
	for b.Loop() {
		JoinInts(nums, ", ")
	}
}
