package l01goquestions

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"testing/synctest"
)

const hangHint = "Do не возвращается: мьютекс держится во время fn, ждущие не просыпаются после завершения вызова или ждут на мьютексе (synctest его не видит, жди на канале или WaitGroup)"

func TestFlightShared(t *testing.T) {
	defer watchdog(t, hangHint)()
	synctest.Test(t, func(t *testing.T) {
		var f Flight
		var calls atomic.Int32
		release := make(chan struct{})
		fn := func() (string, error) {
			calls.Add(1)
			<-release
			return "v", nil
		}
		const n = 10
		var wg sync.WaitGroup
		var leaders, sharers atomic.Int32
		for range n {
			wg.Go(func() {
				v, err, shared := f.Do("k", fn)
				if v != "v" || err != nil {
					t.Errorf(`Do("k") = %q, %v; want "v", nil`, v, err)
				}
				if shared {
					sharers.Add(1)
				} else {
					leaders.Add(1)
				}
			})
		}
		synctest.Wait() // все 10 уже внутри Do
		if c := calls.Load(); c != 1 {
			t.Errorf("пока вызов идёт, fn запущена %d раз, want 1", c)
		}
		close(release)
		wg.Wait()
		if c := calls.Load(); c != 1 {
			t.Errorf("fn вызвана %d раз на 10 одновременных Do, want 1", c)
		}
		if leaders.Load() != 1 || sharers.Load() != n-1 {
			t.Errorf("shared=false у %d вызовов, shared=true у %d; want 1 и %d", leaders.Load(), sharers.Load(), n-1)
		}
	})
}

func TestFlightNotCache(t *testing.T) {
	var f Flight
	calls := 0
	fn := func() (string, error) { calls++; return "v", nil }
	for i := range 2 {
		if v, err, shared := f.Do("k", fn); v != "v" || err != nil || shared {
			t.Errorf("последовательный Do #%d = %q, %v, shared %v; want \"v\", nil, false", i+1, v, err, shared)
		}
	}
	if calls != 2 {
		t.Errorf("два последовательных Do вызвали fn %d раз, want 2: Flight не кеш", calls)
	}
}

func TestFlightKeysIndependent(t *testing.T) {
	defer watchdog(t, hangHint)()
	synctest.Test(t, func(t *testing.T) {
		var f Flight
		release := make(chan struct{})
		done := make(chan struct{})
		go func() {
			defer close(done)
			f.Do("a", func() (string, error) { <-release; return "a", nil })
		}()
		synctest.Wait() // вызов "a" висит внутри fn
		v, err, shared := f.Do("b", func() (string, error) { return "b", nil })
		if v != "b" || err != nil || shared {
			t.Errorf(`Do("b") = %q, %v, %v; want "b", nil, false`, v, err, shared)
		}
		close(release)
		<-done
	})
}

func TestFlightError(t *testing.T) {
	defer watchdog(t, hangHint)()
	synctest.Test(t, func(t *testing.T) {
		var f Flight
		boom := errors.New("boom")
		release := make(chan struct{})
		var wg sync.WaitGroup
		for range 3 {
			wg.Go(func() {
				_, err, _ := f.Do("k", func() (string, error) { <-release; return "", boom })
				if !errors.Is(err, boom) {
					t.Errorf("ждущий Do получил ошибку %v, want boom", err)
				}
			})
		}
		synctest.Wait()
		close(release)
		wg.Wait()
		v, err, _ := f.Do("k", func() (string, error) { return "ok", nil })
		if v != "ok" || err != nil {
			t.Errorf("Do после ошибки = %q, %v; want \"ok\", nil: ошибка не должна кешироваться", v, err)
		}
	})
}
