package l02livecoding

import (
	"context"
	"sync"
	"time"
)

// Limiter: token bucket в форме GCRA. Вместо счётчика жетонов храним tat,
// «теоретическое время прибытия»: момент, когда ведро снова станет полным.
// Так не нужны дроби и фоновые горутины: всё считается от time.Now().
type Limiter struct {
	mu    sync.Mutex
	every time.Duration
	burst time.Duration // (burst-1) * every: насколько tat может опережать «сейчас»
	tat   time.Time
}

// NewLimiter создаёт ограничитель.
func NewLimiter(every time.Duration, burst int) *Limiter {
	burst = max(burst, 1)
	return &Limiter{every: every, burst: time.Duration(burst-1) * every}
}

// reserve забирает жетон, если он есть, иначе говорит, сколько ждать. Под мьютексом.
func (l *Limiter) reserve(now time.Time) (wait time.Duration) {
	tat := l.tat
	if tat.Before(now) {
		tat = now
	}
	if wait = tat.Sub(now) - l.burst; wait > 0 {
		return wait
	}
	l.tat = tat.Add(l.every)
	return 0
}

// Allow забирает жетон, если он есть.
func (l *Limiter) Allow() bool {
	if l.every <= 0 {
		return true
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.reserve(time.Now()) == 0
}

// Wait ждёт жетон; при отмене ctx жетон не тратится.
func (l *Limiter) Wait(ctx context.Context) error {
	if l.every <= 0 {
		return ctx.Err()
	}
	for {
		l.mu.Lock()
		wait := l.reserve(time.Now())
		l.mu.Unlock() // ждём без мьютекса: другие могут работать
		if wait == 0 {
			return nil
		}
		t := time.NewTimer(wait)
		select {
		case <-t.C:
			// жетон мог достаться другому: идём на следующий круг
		case <-ctx.Done():
			t.Stop()
			return ctx.Err()
		}
	}
}
