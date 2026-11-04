package l03hashing

import (
	"testing"
	"time"
)

// bigTimeout: сколько даём решению на большой вход (под -race умножается
// на raceScale). Решение с верной асимптотикой укладывается в доли секунды,
// а квадратичное работает десятки секунд.
const bigTimeout = 3 * time.Second

// within запускает f и валит тест, если f не уложилась в bigTimeout*raceScale.
// Горутина и select здесь нужны только для таймаута, их разберём в модуле
// про конкурентность.
func within(t *testing.T, f func()) {
	t.Helper()
	done := make(chan any, 1)
	go func() {
		defer func() { done <- recover() }()
		f()
	}()
	limit := bigTimeout * raceScale
	select {
	case p := <-done:
		if p != nil {
			t.Fatalf("паника: %v", p)
		}
	case <-time.After(limit):
		t.Fatalf("не уложилось в %v на большом входе: проверь асимптотику", limit)
	}
}

func TestTwoSum(t *testing.T) {
	tests := []struct {
		name   string
		nums   []int
		target int
		found  bool
	}{
		{"пустой", nil, 1, false},
		{"классика", []int{2, 7, 11, 15}, 9, true},
		{"нельзя брать элемент дважды", []int{3, 2, 4}, 6, true},
		{"одинаковые", []int{3, 3}, 6, true},
		{"один элемент", []int{3}, 6, false},
		{"отрицательные", []int{-3, 4, 3, 90}, 0, true},
		{"нет пары", []int{1, 5, 9}, 3, false},
		{"ноль и ноль", []int{0, 4, 3, 0}, 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i, j := TwoSum(tt.nums, tt.target)
			if !tt.found {
				if i != -1 || j != -1 {
					t.Errorf("TwoSum(%v, %d) = %d, %d, want -1, -1", tt.nums, tt.target, i, j)
				}
				return
			}
			if i < 0 || j >= len(tt.nums) || i >= j || tt.nums[i]+tt.nums[j] != tt.target {
				t.Errorf("TwoSum(%v, %d) = %d, %d: не пара i < j с суммой %d", tt.nums, tt.target, i, j, tt.target)
			}
		})
	}
}

func TestTwoSumBig(t *testing.T) {
	const n = 200_000
	nums := make([]int, n)
	for i := range nums {
		nums[i] = (i * 7919) % n // неотрицательные, все разные
	}
	nums[n-2], nums[n-1] = -5, -7 // единственная пара с суммой -12
	var i, j int
	within(t, func() { i, j = TwoSum(nums, -12) })
	if i != n-2 || j != n-1 {
		t.Errorf("TwoSum(большой вход, -12) = %d, %d, want %d, %d", i, j, n-2, n-1)
	}
}
