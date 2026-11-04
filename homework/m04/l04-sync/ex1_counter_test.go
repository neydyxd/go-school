package l04sync

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestCounter(t *testing.T) {
	var c *Counter
	within(t, time.Second, "NewCounter", func() { c = NewCounter() })
	within(t, time.Second, "Get/Inc", func() {
		if got := c.Get("нет"); got != 0 {
			t.Errorf("Get(%q) = %d, want 0", "нет", got)
		}
		c.Inc("a")
		c.Inc("a")
		c.Inc("b")
		if got := c.Get("a"); got != 2 {
			t.Errorf("Get(%q) = %d, want 2", "a", got)
		}
	})
}

func TestCounterConcurrent(t *testing.T) {
	var c *Counter
	within(t, time.Second, "NewCounter", func() { c = NewCounter() })
	const goroutines, iters = 20, 500
	keys := []string{"x", "y", "z"}
	within(t, 5*time.Second, "конкурентные Inc", func() {
		var wg sync.WaitGroup
		for g := range goroutines {
			wg.Go(func() {
				for i := range iters {
					c.Inc(keys[(g+i)%len(keys)])
				}
			})
		}
		wg.Wait()
	})
	total := 0
	for _, k := range keys {
		total += c.Get(k)
	}
	if total != goroutines*iters {
		t.Errorf("сумма Get по ключам = %d, want %d (потерянные инкременты)", total, goroutines*iters)
	}
}

// Snapshot читаем и меняем, пока другие горутины делают Inc.
// Если Snapshot отдаёт внутреннюю мапу, race detector это поймает.
func TestCounterSnapshot(t *testing.T) {
	var c *Counter
	within(t, time.Second, "NewCounter", func() { c = NewCounter() })
	within(t, time.Second, "Inc", func() { c.Inc("a") })

	stop := make(chan struct{})
	var wg sync.WaitGroup
	wg.Go(func() {
		for i := 0; ; i++ {
			select {
			case <-stop:
				return
			default:
				c.Inc(fmt.Sprint("k", i%10))
			}
		}
	})
	within(t, 5*time.Second, "Snapshot под нагрузкой", func() {
		for range 100 {
			s := c.Snapshot()
			for k := range s {
				s[k] = -1
			}
			s["чужой"] = 100
		}
	})
	close(stop)
	wg.Wait()

	if got := c.Get("a"); got != 1 {
		t.Errorf("после изменения снимка Get(%q) = %d, want 1: Snapshot отдал внутреннюю мапу", "a", got)
	}
	if got := c.Get("чужой"); got != 0 {
		t.Errorf("Get(%q) = %d, want 0: Snapshot отдал внутреннюю мапу", "чужой", got)
	}
}
