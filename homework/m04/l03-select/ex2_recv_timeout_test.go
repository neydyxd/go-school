package l03select

import (
	"errors"
	"testing"
	"time"
)

// Тест идёт в пузыре synctest: время фейковое, поэтому «сколько ждали»
// проверяется точно, без допусков на скорость машины.
func TestRecvTimeout(t *testing.T) {
	const d = 50 * time.Millisecond
	tests := []struct {
		name    string
		ch      func() chan int
		wantV   int
		wantErr error
		wait    time.Duration // ровно столько фейкового времени должно пройти
	}{
		{"значение уже есть", func() chan int { c := make(chan int, 1); c <- 3; return c }, 3, nil, 0},
		{"значение придёт позже", func() chan int {
			c := make(chan int)
			go func() { time.Sleep(10 * time.Millisecond); c <- 8 }()
			return c
		}, 8, nil, 10 * time.Millisecond},
		{"ничего не придёт", func() chan int { return make(chan int) }, 0, ErrTimeout, d},
		{"nil-канал", func() chan int { return nil }, 0, ErrTimeout, d},
		{"закрытый пустой: сразу", func() chan int { c := make(chan int); close(c); return c }, 0, ErrClosed, 0},
		{"закрытый с остатком", func() chan int { c := make(chan int, 1); c <- 6; close(c); return c }, 6, nil, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bubble(t, func(t *testing.T) {
				ch := tt.ch()
				var v int
				var err error
				start := time.Now()
				within(t, time.Second, "RecvTimeout", func() { v, err = RecvTimeout(ch, d) })
				elapsed := time.Since(start)
				if v != tt.wantV || !errors.Is(err, tt.wantErr) {
					t.Errorf("RecvTimeout = (%d, %v), want (%d, %v)", v, err, tt.wantV, tt.wantErr)
				}
				if elapsed != tt.wait {
					t.Errorf("RecvTimeout вернулся через %v, want ровно %v", elapsed, tt.wait)
				}
			})
		})
	}
}

// d <= 0 и готовое значение: таймер тоже «готов» сразу. Наивный select
// выбирает случайно и примерно в половине случаев теряет значение.
func TestRecvTimeoutZeroPrefersValue(t *testing.T) {
	for _, d := range []time.Duration{0, -time.Second} {
		for i := range 200 {
			ch := make(chan int, 1)
			ch <- i
			var v int
			var err error
			within(t, time.Second, "RecvTimeout", func() { v, err = RecvTimeout(ch, d) })
			if err != nil || v != i {
				t.Fatalf("RecvTimeout(готовое значение %d, d=%v) = (%d, %v), want (%d, nil): готовое значение важнее таймера", i, d, v, err, i)
			}
		}
		var err error
		within(t, time.Second, "RecvTimeout", func() { _, err = RecvTimeout(make(chan int), d) })
		if !errors.Is(err, ErrTimeout) {
			t.Fatalf("RecvTimeout(пустой канал, d=%v) error = %v, want ErrTimeout", d, err)
		}
	}
}
