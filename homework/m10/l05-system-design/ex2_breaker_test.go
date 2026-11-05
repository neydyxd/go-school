package l05systemdesign

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// clock: фейковые часы, двигаются только руками.
type clock struct {
	mu sync.Mutex
	t  time.Time
}

func newClock() *clock { return &clock{t: time.Date(2026, 11, 5, 12, 0, 0, 0, time.UTC)} }

func (c *clock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.t
}

func (c *clock) Advance(d time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.t = c.t.Add(d)
}

var errDown = errors.New("зависимость лежит")

func ok() error   { return nil }
func fail() error { return errDown }

func TestBreakerOpensAfterThreshold(t *testing.T) {
	c := newClock()
	b := NewBreaker(3, 10*time.Second, c.Now)
	if s := b.State(); s != Closed {
		t.Fatalf("начальное состояние = %v, want closed", s)
	}
	// Неудачи вперемешку с успехом не размыкают: считаем подряд.
	for _, f := range []func() error{fail, fail, ok, fail, fail} {
		b.Call(f)
	}
	if s := b.State(); s != Closed {
		t.Fatalf("после fail,fail,ok,fail,fail состояние = %v, want closed (успех обнуляет счётчик)", s)
	}
	if err := b.Call(fail); !errors.Is(err, errDown) {
		t.Fatalf("Call(fail) = %v, want ошибку самой fn", err)
	}
	if s := b.State(); s != Open {
		t.Fatalf("после 3 неудач подряд состояние = %v, want open", s)
	}
	called := false
	if err := b.Call(func() error { called = true; return nil }); !errors.Is(err, ErrOpen) || called {
		t.Errorf("Call в open = %v, вызвана fn = %v, want ErrOpen без вызова", err, called)
	}
}

func TestBreakerHalfOpen(t *testing.T) {
	c := newClock()
	b := NewBreaker(2, 10*time.Second, c.Now)
	b.Call(fail)
	b.Call(fail)

	c.Advance(9 * time.Second)
	if err := b.Call(ok); !errors.Is(err, ErrOpen) {
		t.Fatalf("через 9s при cooldown 10s Call = %v, want ErrOpen", err)
	}
	c.Advance(time.Second)
	if s := b.State(); s != HalfOpen {
		t.Fatalf("через 10s State = %v, want half-open", s)
	}

	// Проба провалилась: снова open, cooldown отсчитывается заново.
	if err := b.Call(fail); !errors.Is(err, errDown) {
		t.Fatalf("пробный Call = %v, want ошибку fn", err)
	}
	if s := b.State(); s != Open {
		t.Fatalf("после неудачной пробы State = %v, want open", s)
	}
	c.Advance(5 * time.Second)
	if err := b.Call(ok); !errors.Is(err, ErrOpen) {
		t.Fatalf("через 5s после неудачной пробы Call = %v, want ErrOpen: cooldown заново", err)
	}
	c.Advance(5 * time.Second)

	// Проба успешна: closed со сброшенным счётчиком.
	if err := b.Call(ok); err != nil {
		t.Fatalf("успешная проба = %v, want nil", err)
	}
	if s := b.State(); s != Closed {
		t.Fatalf("после успешной пробы State = %v, want closed", s)
	}
	b.Call(fail)
	if s := b.State(); s != Closed {
		t.Errorf("одна неудача после восстановления при threshold 2: State = %v, want closed (счётчик сброшен)", s)
	}
}

func TestBreakerSingleProbe(t *testing.T) {
	c := newClock()
	b := NewBreaker(1, time.Second, c.Now)
	b.Call(fail)
	c.Advance(time.Second)

	entered, release := make(chan struct{}), make(chan struct{})
	probeDone := make(chan error, 1)
	go func() {
		probeDone <- b.Call(func() error {
			close(entered)
			<-release
			return nil
		})
	}()
	select {
	case <-entered:
	case <-time.After(3 * time.Second):
		t.Fatal("пробный вызов не начался за 3s")
	}

	// Пока проба идёт, остальные получают ErrOpen и не ходят в зависимость.
	var extra atomic.Int32
	var wg sync.WaitGroup
	for range 20 {
		wg.Go(func() {
			if err := b.Call(func() error { extra.Add(1); return nil }); !errors.Is(err, ErrOpen) {
				t.Errorf("Call во время пробы = %v, want ErrOpen", err)
			}
		})
	}
	waited := make(chan struct{})
	go func() {
		wg.Wait()
		close(waited)
	}()
	select {
	case <-waited:
	case <-time.After(3 * time.Second):
		close(release)
		t.Fatal("Call во время пробы не вернулся за 3s: fn выполняется под мьютексом?")
	}
	if n := extra.Load(); n != 0 {
		t.Errorf("во время пробы fn вызвана ещё %d раз, want 0", n)
	}
	close(release)
	if err := <-probeDone; err != nil {
		t.Fatalf("проба = %v", err)
	}
	if s := b.State(); s != Closed {
		t.Errorf("после пробы State = %v, want closed", s)
	}
}

func TestBreakerNoLockDuringCall(t *testing.T) {
	b := NewBreaker(5, time.Second, newClock().Now)
	entered, release := make(chan struct{}), make(chan struct{})
	go b.Call(func() error {
		close(entered)
		<-release
		return nil
	})
	<-entered
	defer close(release)

	done := make(chan struct{})
	go func() {
		b.Call(ok)
		b.State()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("Call и State ждут, пока закончится чужой медленный вызов: fn выполняется под мьютексом")
	}
}

func TestBreakerConcurrent(t *testing.T) {
	c := newClock()
	b := NewBreaker(3, time.Second, c.Now)
	var wg sync.WaitGroup
	for i := range 200 {
		wg.Go(func() {
			if i%2 == 0 {
				b.Call(fail)
			} else {
				b.Call(ok)
			}
			b.State()
			if i%50 == 0 {
				c.Advance(time.Second)
			}
		})
	}
	wg.Wait()
}
