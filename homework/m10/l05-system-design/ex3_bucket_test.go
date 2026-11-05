package l05systemdesign

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// allowN вызывает Allow n раз и возвращает число разрешённых.
func allowN(b *TokenBucket, n int) int {
	ok := 0
	for range n {
		if b.Allow() {
			ok++
		}
	}
	return ok
}

func TestTokenBucket(t *testing.T) {
	c := newClock()
	b := NewTokenBucket(1, 3, c.Now) // 1 токен в секунду, ведро на 3

	if got := allowN(b, 5); got != 3 {
		t.Fatalf("полное ведро: разрешено %d из 5, want 3 (burst)", got)
	}
	c.Advance(500 * time.Millisecond)
	if b.Allow() {
		t.Fatal("через 0.5s после опустошения Allow = true, want false (полтокена мало)")
	}
	c.Advance(500 * time.Millisecond)
	if !b.Allow() {
		t.Fatal("ещё через 0.5s Allow = false, want true: половинки должны сложиться")
	}
	if b.Allow() {
		t.Fatal("второй Allow подряд = true, want false")
	}
	c.Advance(2 * time.Second)
	if got := allowN(b, 5); got != 2 {
		t.Errorf("через 2s: разрешено %d, want 2", got)
	}
	c.Advance(time.Hour)
	if got := allowN(b, 10); got != 3 {
		t.Errorf("после часа простоя: разрешено %d, want 3 (не больше burst)", got)
	}
}

func TestTokenBucketFractional(t *testing.T) {
	c := newClock()
	b := NewTokenBucket(2, 2, c.Now) // 2 токена в секунду, ведро на 2
	allowN(b, 2)
	got := 0
	for range 40 {
		c.Advance(250 * time.Millisecond) // по полтокена
		if b.Allow() {
			got++
		}
	}
	// За 10 секунд набегает 20 токенов, и ни один не теряется: половинки
	// складываются, а до потолка ведра остаток не доходит.
	if got != 20 {
		t.Errorf("за 10s при 2 токенах/с разрешено %d, want 20 (дробные токены не должны теряться)", got)
	}
}

func TestTokenBucketClockBackwards(t *testing.T) {
	c := newClock()
	b := NewTokenBucket(1, 2, c.Now)
	allowN(b, 2)
	c.Advance(-time.Hour)
	if b.Allow() {
		t.Fatal("часы назад: Allow = true, want false (пополнения нет)")
	}
	c.Advance(2 * time.Second)
	if got := allowN(b, 3); got != 2 {
		t.Errorf("через 2s после перевода часов разрешено %d, want 2: отсчёт от нового времени, токены не уходят в минус", got)
	}
}

func TestTokenBucketConcurrent(t *testing.T) {
	c := newClock()
	b := NewTokenBucket(10, 50, c.Now)
	var allowed atomic.Int32
	var wg sync.WaitGroup
	for range 200 {
		wg.Go(func() {
			if b.Allow() {
				allowed.Add(1)
			}
		})
	}
	wg.Wait()
	if n := allowed.Load(); n != 50 {
		t.Errorf("200 конкурентных запросов к ведру на 50: разрешено %d, want 50", n)
	}
}
