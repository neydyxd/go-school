package l07problems

import (
	"errors"
	"sync"
	"testing"
	"time"
)

// runWithin запускает f и падает, если она не закончилась за d.
// Так дедлок в решении превращается в понятное сообщение, а не вечный тест.
func runWithin(t *testing.T, d time.Duration, what string, f func()) {
	t.Helper()
	done := make(chan struct{})
	go func() {
		defer close(done)
		f()
	}()
	select {
	case <-done:
	case <-time.After(d):
		t.Fatalf("%s: нет завершения за %v, похоже на дедлок", what, d)
	}
}

func TestTransferRules(t *testing.T) {
	tests := []struct {
		name         string
		amount       int
		same         bool
		wantErr      error
		wantA, wantB int
	}{
		{"обычный перевод", 30, false, nil, 70, 30},
		{"ровно весь баланс", 100, false, nil, 0, 100},
		{"не хватает", 101, false, ErrInsufficient, 100, 0},
		{"ноль", 0, false, ErrInvalidAmount, 100, 0},
		{"отрицательная сумма", -5, false, ErrInvalidAmount, 100, 0},
		{"самому себе", 10, true, ErrSameAccount, 100, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a, b := NewAccount(1, 100), NewAccount(2, 0)
			to := b
			if tt.same {
				to = a
			}
			var err error
			runWithin(t, time.Second, "Transfer", func() { err = Transfer(a, to, tt.amount) })
			if !errors.Is(err, tt.wantErr) || (tt.wantErr == nil && err != nil) {
				t.Fatalf("Transfer(a, %d) = %v, want %v", tt.amount, err, tt.wantErr)
			}
			if a.Balance() != tt.wantA || b.Balance() != tt.wantB {
				t.Errorf("балансы после перевода: a=%d, b=%d, want a=%d, b=%d", a.Balance(), b.Balance(), tt.wantA, tt.wantB)
			}
		})
	}
}

func TestTransferNoDeadlock(t *testing.T) {
	a, b, c := NewAccount(1, 1000), NewAccount(2, 1000), NewAccount(3, 1000)
	pairs := [][2]*Account{{a, b}, {b, a}, {b, c}, {c, b}, {c, a}, {a, c}}
	runWithin(t, 5*time.Second, "встречные переводы", func() {
		var wg sync.WaitGroup
		for _, p := range pairs {
			wg.Go(func() {
				for range 2000 {
					_ = Transfer(p[0], p[1], 1)
				}
			})
		}
		wg.Wait()
	})
	if total := a.Balance() + b.Balance() + c.Balance(); total != 3000 {
		t.Errorf("сумма на счетах %d, want 3000: деньги не должны появляться и исчезать", total)
	}
}

// TestTransferLockOrder проверяет порядок захвата детерминированно, без
// надежды поймать дедлок на многоядерной машине. Держим мьютекс счёта с
// меньшим ID и смотрим через TryLock: если Transfer(b, a) успел захватить b,
// значит, он берёт мьютексы не по возрастанию ID.
func TestTransferLockOrder(t *testing.T) {
	a, b := NewAccount(1, 100), NewAccount(2, 100)
	a.mu.Lock()
	done := make(chan error, 1)
	go func() { done <- Transfer(b, a, 1) }()
	deadline := time.Now().Add(200 * time.Millisecond)
	for time.Now().Before(deadline) {
		if !b.mu.TryLock() {
			a.mu.Unlock()
			<-done
			t.Fatal("Transfer(b, a) захватил b раньше a: при встречном переводе это дедлок; бери мьютексы по возрастанию ID")
		}
		b.mu.Unlock()
		time.Sleep(time.Millisecond)
	}
	a.mu.Unlock()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Transfer(b, a, 1) = %v, want nil", err)
		}
	case <-time.After(time.Second):
		t.Fatal("Transfer(b, a, 1): нет завершения за 1s после освобождения мьютекса, похоже на дедлок")
	}
}
