package l03select

import (
	"testing"
	"time"
)

func TestTrySend(t *testing.T) {
	tests := []struct {
		name   string
		ch     chan int
		want   bool
		wantIn int // сколько значений должно оказаться в буфере
	}{
		{"есть место в буфере", make(chan int, 1), true, 1},
		{"буфер полон", func() chan int { c := make(chan int, 1); c <- 0; return c }(), false, 1},
		{"небуферизованный без получателя", make(chan int), false, 0},
		{"nil-канал", nil, false, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got bool
			within(t, time.Second, "TrySend", func() { got = TrySend(tt.ch, 7) })
			if got != tt.want {
				t.Errorf("TrySend = %v, want %v", got, tt.want)
			}
			if len(tt.ch) != tt.wantIn {
				t.Errorf("после TrySend len(ch) = %d, want %d", len(tt.ch), tt.wantIn)
			}
		})
	}
}

// Небуферизованный канал с ждущим получателем: отправка должна пройти.
func TestTrySendToWaitingReceiver(t *testing.T) {
	ch := make(chan int)
	got := make(chan int, 1)
	go func() { got <- <-ch }()
	deadline := time.Now().Add(time.Second)
	sent := false
	within(t, 2*time.Second, "TrySend", func() {
		// Получатель мог ещё не встать в очередь, поэтому пробуем в цикле.
		for !sent && time.Now().Before(deadline) {
			sent = TrySend(ch, 5)
			time.Sleep(time.Millisecond)
		}
	})
	if !sent {
		t.Fatal("TrySend так и не отправил значение ждущему получателю")
	}
	if v := <-got; v != 5 {
		t.Errorf("получатель прочитал %d, want 5", v)
	}
}

func TestTryRecv(t *testing.T) {
	type res struct {
		v int
		s RecvStatus
	}
	tests := []struct {
		name string
		ch   func() chan int
		want []res // результаты нескольких TryRecv подряд
	}{
		{"пустой открытый", func() chan int { return make(chan int, 1) }, []res{{0, Empty}}},
		{"nil-канал", func() chan int { return nil }, []res{{0, Empty}, {0, Empty}}},
		{"значение в буфере", func() chan int { c := make(chan int, 2); c <- 4; return c }, []res{{4, Received}, {0, Empty}}},
		{"ноль это значение", func() chan int { c := make(chan int, 1); c <- 0; return c }, []res{{0, Received}, {0, Empty}}},
		{"закрытый пустой", func() chan int { c := make(chan int); close(c); return c }, []res{{0, Closed}, {0, Closed}}},
		{"закрытый с остатком", func() chan int {
			c := make(chan int, 2)
			c <- 1
			c <- 2
			close(c)
			return c
		}, []res{{1, Received}, {2, Received}, {0, Closed}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ch := tt.ch()
			for i, w := range tt.want {
				var v int
				var s RecvStatus
				within(t, time.Second, "TryRecv", func() { v, s = TryRecv(ch) })
				if v != w.v || s != w.s {
					t.Fatalf("TryRecv #%d = (%d, %d), want (%d, %d)", i+1, v, s, w.v, w.s)
				}
			}
		})
	}
}
