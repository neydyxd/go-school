package l08sorting

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
		t.Fatalf("не уложилось в %v: проверь асимптотику и нет ли бесконечного цикла", limit)
	}
}
