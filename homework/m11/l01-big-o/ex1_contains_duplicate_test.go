package l01bigo

import (
	"slices"
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

func TestContainsDuplicate(t *testing.T) {
	tests := []struct {
		name string
		in   []int
		want bool
	}{
		{"nil", nil, false},
		{"пустой", []int{}, false},
		{"один элемент", []int{7}, false},
		{"повтор в конце", []int{1, 2, 3, 1}, true},
		{"все разные", []int{1, 2, 3, 4}, false},
		{"много повторов", []int{1, 1, 1, 3, 3, 4, 3, 2, 4, 2}, true},
		{"отрицательные и ноль", []int{-1, 0, 1, -1}, true},
		{"соседние", []int{5, 5}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			orig := slices.Clone(tt.in)
			if got := ContainsDuplicate(tt.in); got != tt.want {
				t.Errorf("ContainsDuplicate(%v) = %v, want %v", orig, got, tt.want)
			}
			if !slices.Equal(tt.in, orig) {
				t.Errorf("ContainsDuplicate изменила вход: было %v, стало %v", orig, tt.in)
			}
		})
	}
}

func TestContainsDuplicateBig(t *testing.T) {
	const n = 400_000
	nums := make([]int, n)
	for i := range nums {
		nums[i] = (i * 7919) % n // все разные: 7919 простое и не делит n
	}
	var got bool
	within(t, func() { got = ContainsDuplicate(nums) })
	if got {
		t.Errorf("ContainsDuplicate(%d разных чисел) = true, want false", n)
	}
	nums[n-1] = nums[0]
	within(t, func() { got = ContainsDuplicate(nums) })
	if !got {
		t.Errorf("ContainsDuplicate(повтор первого и последнего) = false, want true")
	}
}
