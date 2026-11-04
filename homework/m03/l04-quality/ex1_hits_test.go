package l04quality

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

// withTimeout выполняет f в отдельной горутине и падает с понятным сообщением,
// если f не закончилась за d. Так взаимоблокировка не вешает тест на 10 минут.
func withTimeout(t *testing.T, d time.Duration, f func()) {
	t.Helper()
	done := make(chan struct{})
	go func() {
		defer close(done)
		f()
	}()
	select {
	case <-done:
	case <-time.After(d):
		t.Fatalf("не закончилось за %v: похоже на взаимоблокировку (Lock без Unlock? метод под замком вызывает другой метод, который тоже берёт замок?)", d)
	}
}

func TestHitsSequential(t *testing.T) {
	var get0, get1, total int
	withTimeout(t, 5*time.Second, func() {
		h := NewHits()
		get0 = h.Get("/")
		h.Add("/")
		h.Add("/")
		h.Add("/about")
		get1 = h.Get("/")
		total = h.Total()
	})
	if get0 != 0 {
		t.Errorf("Get(/) на пустом счётчике = %d, want 0", get0)
	}
	if get1 != 2 {
		t.Errorf("Get(/) = %d, want 2", get1)
	}
	if total != 3 {
		t.Errorf("Total() = %d, want 3", total)
	}
}

func TestHitsConcurrent(t *testing.T) {
	const workers, perWorker = 50, 200
	h := NewHits()
	withTimeout(t, 10*time.Second, func() {
		var wg sync.WaitGroup
		for w := range workers {
			wg.Go(func() {
				key := fmt.Sprintf("/page/%d", w%5)
				for range perWorker {
					h.Add(key)
					h.Get(key)
					h.Total()
				}
			})
		}
		wg.Wait()
	})
	var total int
	got := map[string]int{}
	withTimeout(t, 5*time.Second, func() {
		total = h.Total()
		for k := range 5 {
			key := fmt.Sprintf("/page/%d", k)
			got[key] = h.Get(key)
		}
	})
	if total != workers*perWorker {
		t.Errorf("Total() = %d, want %d: часть инкрементов потерялась", total, workers*perWorker)
	}
	for key, n := range got {
		if n != workers/5*perWorker {
			t.Errorf("Get(%s) = %d, want %d", key, n, workers/5*perWorker)
		}
	}
}
