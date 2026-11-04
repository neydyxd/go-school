package l04slidingwindow

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

func TestMaxSumK(t *testing.T) {
	tests := []struct {
		name   string
		nums   []int
		k      int
		want   int
		wantOK bool
	}{
		{"пустой", nil, 1, 0, false},
		{"k ноль", []int{1, 2}, 0, 0, false},
		{"k отрицательный", []int{1, 2}, -1, 0, false},
		{"k больше длины", []int{1, 2}, 3, 0, false},
		{"k равно длине", []int{1, 2, 3}, 3, 6, true},
		{"классика", []int{1, 12, -5, -6, 50, 3}, 4, 51, true},
		{"все отрицательные", []int{-5, -2, -8, -1}, 2, -7, true},
		{"окно в конце", []int{1, 1, 1, 9, 9}, 2, 18, true},
		{"окно в начале", []int{9, 9, 1, 1, 1}, 2, 18, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := MaxSumK(tt.nums, tt.k)
			if got != tt.want || ok != tt.wantOK {
				t.Errorf("MaxSumK(%v, %d) = %d, %v, want %d, %v", tt.nums, tt.k, got, ok, tt.want, tt.wantOK)
			}
		})
	}
}

func TestMaxSumKBig(t *testing.T) {
	const n = 400_000
	const k = 200_000
	nums := make([]int, n)
	for i := range nums {
		nums[i] = i%5 - 2 // сумма любых 5 подряд равна 0
	}
	nums[n-1] = 1_000_000
	var got int
	var ok bool
	within(t, func() { got, ok = MaxSumK(nums, k) })
	// Последнее окно: сумма первых k-1 его элементов плюс 10^6.
	want := 1_000_000
	for _, v := range nums[n-k : n-1] {
		want += v
	}
	if got != want || !ok {
		t.Errorf("MaxSumK(большой вход, %d) = %d, %v, want %d, true", k, got, ok, want)
	}
}
