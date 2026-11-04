package l04sync

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestEvent(t *testing.T) {
	var e *Event
	within(t, time.Second, "NewEvent", func() { e = NewEvent() })

	var fired bool
	var done <-chan struct{}
	within(t, time.Second, "Fired/Done до Fire", func() {
		fired = e.Fired()
		done = e.Done()
	})
	if fired {
		t.Fatal("Fired() = true до вызова Fire")
	}
	if done == nil {
		t.Fatal("Done() вернул nil-канал")
	}
	select {
	case <-done:
		t.Fatal("Done() закрыт до вызова Fire")
	default:
	}

	// Ждущие горутины должны проснуться все разом.
	const waiters = 10
	woke := make(chan struct{}, waiters)
	for range waiters {
		go func() {
			<-done
			woke <- struct{}{}
		}()
	}

	// Много одновременных Fire: ни паники «close of closed channel», ни гонки.
	within(t, 2*time.Second, "конкурентные Fire", func() {
		var wg sync.WaitGroup
		for range 100 {
			wg.Go(e.Fire)
		}
		wg.Wait()
	})

	for i := range waiters {
		select {
		case <-woke:
		case <-time.After(time.Second):
			t.Fatalf("после Fire проснулись только %d из %d ждущих", i, waiters)
		}
	}
	within(t, time.Second, "Fired/Done после Fire", func() {
		fired = e.Fired()
		if e.Done() != done {
			t.Error("Done() вернул другой канал: ждущие на старом никогда не проснутся")
		}
	})
	if !fired {
		t.Error("Fired() = false после Fire")
	}
	within(t, time.Second, "повторный Fire", e.Fire)
}

// Много раундов: горутины стартуют одновременно и сразу зовут Fire.
// Проверка «закрыт ли канал» + close без Once рано или поздно закроет
// канал дважды.
func TestEventFireRace(t *testing.T) {
	const rounds, goroutines = 500, 16
	var panics atomic.Int32
	within(t, 10*time.Second, "раунды конкурентных Fire", func() {
		for range rounds {
			e := NewEvent()
			start := make(chan struct{})
			var wg sync.WaitGroup
			for range goroutines {
				wg.Go(func() {
					defer func() {
						if recover() != nil {
							panics.Add(1)
						}
					}()
					<-start
					e.Fire()
				})
			}
			close(start)
			wg.Wait()
		}
	})
	if n := panics.Load(); n > 0 {
		t.Fatalf("Fire паниковал %d раз при одновременных вызовах (двойной close?)", n)
	}
}
