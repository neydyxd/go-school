package l02channels

import (
	"sync"
	"sync/atomic"
	"testing"
	"testing/synctest"
	"time"
)

// Обычный int под ChanMutex: race detector проверит, что канал даёт
// happens-before между Unlock одной горутины и Lock другой.
func TestChanMutexCounter(t *testing.T) {
	const goroutines, iters = 50, 200
	var m *ChanMutex
	within(t, time.Second, "NewChanMutex", func() { m = NewChanMutex() })
	counter := 0
	var inside atomic.Int32
	var overlap atomic.Bool
	within(t, 5*time.Second, "конкурентные Lock/Unlock", func() {
		var wg sync.WaitGroup
		for range goroutines {
			wg.Go(func() {
				for range iters {
					m.Lock()
					if inside.Add(1) > 1 {
						overlap.Store(true)
					}
					counter++
					inside.Add(-1)
					m.Unlock()
				}
			})
		}
		wg.Wait()
	})
	if overlap.Load() {
		t.Error("две горутины одновременно оказались внутри Lock/Unlock")
	}
	if counter != goroutines*iters {
		t.Errorf("counter = %d, want %d", counter, goroutines*iters)
	}
}

// Тест идёт в пузыре synctest: synctest.Wait() дожидается, пока все горутины
// уснут, поэтому «второй Lock заблокирован» проверяется без пауз и допусков.
func TestChanMutexBlocks(t *testing.T) {
	bubble(t, func(t *testing.T) {
		var m *ChanMutex
		within(t, time.Second, "NewChanMutex + Lock", func() {
			m = NewChanMutex()
			m.Lock()
		})
		acquired := make(chan struct{})
		go func() {
			defer func() { recover() }() // заготовка паникует, тест упадёт выше
			m.Lock()
			close(acquired)
		}()
		synctest.Wait() // второй Lock либо прошёл, либо уснул
		select {
		case <-acquired:
			t.Fatal("второй Lock прошёл, хотя мьютекс захвачен")
		default:
		}
		within(t, time.Second, "Unlock", m.Unlock)
		synctest.Wait()
		select {
		case <-acquired:
		default:
			t.Fatal("после Unlock ожидающий Lock так и не прошёл")
		}
	})
}
