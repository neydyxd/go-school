package l03internalsperf

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

// withTimeout выполняет f и падает с понятным сообщением, если f зависла.
func withTimeout(t *testing.T, d time.Duration, what string, f func()) {
	t.Helper()
	done := make(chan struct{})
	go func() {
		defer close(done)
		f()
	}()
	select {
	case <-done:
	case <-time.After(d):
		t.Fatalf("%s: не закончилось за %v (кто-то ждёт и его не разбудили?)", what, d)
	}
}

// panics сообщает, запаниковала ли f.
func panics(f func()) (p bool) {
	defer func() { p = recover() != nil }()
	f()
	return false
}

func TestMiniChanFIFO(t *testing.T) {
	withTimeout(t, 2*time.Second, "FIFO", func() {
		c := NewMiniChan[string](3)
		c.Send("a")
		c.Send("b")
		c.Send("c")
		if c.Len() != 3 {
			t.Errorf("Len() = %d, want 3", c.Len())
		}
		for _, want := range []string{"a", "b", "c"} {
			if v, ok := c.Recv(); v != want || !ok {
				t.Errorf("Recv() = %q, %v, want %q, true", v, ok, want)
			}
		}
	})
}

func TestMiniChanSendBlocksWhenFull(t *testing.T) {
	c := NewMiniChan[int](1)
	withTimeout(t, 2*time.Second, "первый Send в пустой канал", func() { c.Send(1) })
	sent := make(chan struct{})
	go func() {
		c.Send(2)
		close(sent)
	}()
	select {
	case <-sent:
		t.Fatal("Send в полный канал вернулся сразу, want ожидание, пока освободится место")
	case <-time.After(50 * time.Millisecond):
	}
	withTimeout(t, 2*time.Second, "Recv из полного канала", func() {
		if v, _ := c.Recv(); v != 1 {
			t.Errorf("Recv() = %d, want 1", v)
		}
	})
	select {
	case <-sent:
	case <-time.After(2 * time.Second):
		t.Fatal("после Recv ждущий Send не проснулся за 2s")
	}
	withTimeout(t, 2*time.Second, "Recv второго значения", func() {
		if v, _ := c.Recv(); v != 2 {
			t.Errorf("Recv() = %d, want 2", v)
		}
	})
}

func TestMiniChanRecvBlocksWhenEmpty(t *testing.T) {
	c := NewMiniChan[int](2)
	got := make(chan int, 1)
	go func() {
		v, _ := c.Recv()
		got <- v
	}()
	select {
	case v := <-got:
		t.Fatalf("Recv из пустого канала вернул %d сразу, want ожидание", v)
	case <-time.After(50 * time.Millisecond):
	}
	withTimeout(t, 2*time.Second, "Send", func() { c.Send(7) })
	select {
	case v := <-got:
		if v != 7 {
			t.Errorf("Recv() = %d, want 7", v)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("после Send ждущий Recv не проснулся за 2s")
	}
}

func TestMiniChanCloseDrains(t *testing.T) {
	withTimeout(t, 2*time.Second, "дочитывание закрытого канала", func() {
		c := NewMiniChan[int](4)
		c.Send(1)
		c.Send(2)
		c.Close()
		for _, want := range []int{1, 2} {
			if v, ok := c.Recv(); v != want || !ok {
				t.Errorf("Recv() после Close = %d, %v, want %d, true (сначала дочитай буфер)", v, ok, want)
			}
		}
		for range 2 {
			if v, ok := c.Recv(); v != 0 || ok {
				t.Errorf("Recv() из закрытого пустого = %d, %v, want 0, false", v, ok)
			}
		}
	})
}

func TestMiniChanCloseWakesReceivers(t *testing.T) {
	c := NewMiniChan[int](1)
	const n = 3
	var wg sync.WaitGroup
	results := make(chan bool, n)
	for range n {
		wg.Go(func() {
			_, ok := c.Recv()
			results <- ok
		})
	}
	time.Sleep(20 * time.Millisecond) // дать получателям уснуть; без сна тест тоже корректен
	c.Close()
	withTimeout(t, 2*time.Second, "Close должен разбудить всех ждущих получателей", wg.Wait)
	close(results)
	for ok := range results {
		if ok {
			t.Error("Recv() после Close пустого канала вернул ok = true, want false")
		}
	}
}

func TestMiniChanPanics(t *testing.T) {
	withTimeout(t, 2*time.Second, "проверка паник", func() {
		c := NewMiniChan[int](1)
		c.Close()
		if !panics(func() { c.Send(1) }) {
			t.Error("Send в закрытый канал не запаниковал")
		}
		if !panics(c.Close) {
			t.Error("повторный Close не запаниковал")
		}
		if !panics(func() { NewMiniChan[int](0) }) {
			t.Error("NewMiniChan(0) не запаниковал")
		}
	})
}

func TestMiniChanCloseWakesSender(t *testing.T) {
	c := NewMiniChan[int](1)
	withTimeout(t, 2*time.Second, "Send в пустой канал", func() { c.Send(1) })
	result := make(chan bool, 1)
	go func() { result <- panics(func() { c.Send(2) }) }()
	time.Sleep(20 * time.Millisecond)
	c.Close()
	select {
	case p := <-result:
		if !p {
			t.Error("ждущий Send после Close завершился без паники, want панику (как в Go)")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Close не разбудил ждущего отправителя за 2s")
	}
}

func TestMiniChanConcurrent(t *testing.T) {
	// Ёмкость 1 и много горутин с обеих сторон: на ней проявляется
	// «пробуждение не того»: отправитель будит отправителя, и все спят.
	for _, capacity := range []int{1, 8} {
		const producers, consumers, perProducer = 8, 8, 2000
		c := NewMiniChan[int](capacity)
		seen := make([]int, producers*perProducer)
		var mu sync.Mutex
		withTimeout(t, 10*time.Second, fmt.Sprintf("ёмкость %d, 8 отправителей и 8 получателей", capacity), func() {
			var prod, cons sync.WaitGroup
			for p := range producers {
				prod.Go(func() {
					for i := range perProducer {
						c.Send(p*perProducer + i)
					}
				})
			}
			for range consumers {
				cons.Go(func() {
					for {
						v, ok := c.Recv()
						if !ok {
							return
						}
						mu.Lock()
						seen[v]++
						mu.Unlock()
					}
				})
			}
			prod.Wait()
			c.Close()
			cons.Wait()
		})
		for v, n := range seen {
			if n != 1 {
				t.Fatalf("ёмкость %d: значение %d получено %d раз, want ровно 1", capacity, v, n)
			}
		}
	}
}
