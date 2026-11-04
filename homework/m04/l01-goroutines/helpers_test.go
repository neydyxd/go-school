package l01goroutines

import (
	"testing"
	"time"
)

// within запускает f и ждёт её не дольше d. Зависшее решение не вешает
// тесты, а паника (например, panic("TODO")) превращается в понятный провал.
func within(t *testing.T, d time.Duration, what string, f func()) {
	t.Helper()
	done := make(chan any, 1)
	go func() {
		defer func() { done <- recover() }()
		f()
	}()
	select {
	case r := <-done:
		if r != nil {
			t.Fatalf("%s: паника: %v", what, r)
		}
	case <-time.After(d):
		t.Fatalf("%s: не завершилось за %v (зависание или дедлок?)", what, d)
	}
}
