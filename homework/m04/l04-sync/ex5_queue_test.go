package l04sync

import (
	"slices"
	"sync"
	"sync/atomic"
	"testing"
	"testing/synctest"
	"time"
)

func newQueue(t *testing.T, capacity int) *Queue {
	t.Helper()
	var q *Queue
	within(t, time.Second, "NewQueue", func() { q = NewQueue(capacity) })
	return q
}

// started запускает f в горутине; канал закрывается, когда f вернулась.
func started(f func()) <-chan struct{} {
	done := make(chan struct{})
	go func() {
		defer close(done)
		defer func() { recover() }()
		f()
	}()
	return done
}

func TestQueueFIFO(t *testing.T) {
	q := newQueue(t, 3)
	var got []int
	within(t, time.Second, "Put/Get без ожидания", func() {
		for _, v := range []int{1, 2, 3} {
			if !q.Put(v) {
				t.Errorf("Put(%d) = false на открытой очереди", v)
			}
		}
		for range 3 {
			v, ok := q.Get()
			if !ok {
				t.Error("Get() вернул ok = false на непустой очереди")
			}
			got = append(got, v)
		}
	})
	if want := []int{1, 2, 3}; !slices.Equal(got, want) {
		t.Errorf("порядок Get = %v, want %v", got, want)
	}
}

// Тесты на блокировку идут в пузыре synctest: synctest.Wait() дожидается,
// пока все горутины уснут (sync.Cond.Wait тоже считается сном), поэтому
// «Get ждёт» и «Close разбудил всех» проверяются без пауз и допусков.
func TestQueueGetBlocksUntilPut(t *testing.T) {
	bubble(t, func(t *testing.T) {
		q := newQueue(t, 1)
		var v int
		done := started(func() { v, _ = q.Get() })
		synctest.Wait()
		select {
		case <-done:
			t.Fatal("Get на пустой очереди вернулся, не дождавшись Put")
		default:
		}
		within(t, time.Second, "Put", func() { q.Put(5) })
		synctest.Wait()
		select {
		case <-done:
		default:
			t.Fatal("Get не проснулся после Put")
		}
		if v != 5 {
			t.Errorf("Get() = %d, want 5", v)
		}
	})
}

func TestQueuePutBlocksWhenFull(t *testing.T) {
	bubble(t, func(t *testing.T) {
		q := newQueue(t, 1)
		within(t, time.Second, "первый Put", func() { q.Put(1) })
		done := started(func() { q.Put(2) })
		synctest.Wait()
		select {
		case <-done:
			t.Fatal("Put в полную очередь вернулся, не дождавшись Get")
		default:
		}
		var got []int
		within(t, time.Second, "Get освобождает место", func() {
			v, _ := q.Get()
			got = append(got, v)
		})
		synctest.Wait()
		select {
		case <-done:
		default:
			t.Fatal("Put не проснулся после Get")
		}
		within(t, time.Second, "Get", func() {
			v, _ := q.Get()
			got = append(got, v)
		})
		if want := []int{1, 2}; !slices.Equal(got, want) {
			t.Errorf("Get = %v, want %v", got, want)
		}
	})
}

func TestQueueCapacityZero(t *testing.T) {
	q := newQueue(t, 0)
	within(t, time.Second, "Put/Get при capacity 0 (как 1)", func() {
		q.Put(7)
		if v, ok := q.Get(); v != 7 || !ok {
			t.Errorf("Get() = (%d, %v), want (7, true)", v, ok)
		}
	})
}

func TestQueueClose(t *testing.T) {
	bubble(t, func(t *testing.T) {
		q := newQueue(t, 2)
		within(t, time.Second, "Put", func() { q.Put(1) })

		// Получатели на пустой очереди после того, как остаток разберут.
		within(t, time.Second, "Get остатка", func() { q.Get() })
		const waiters = 5
		results := make(chan bool, waiters)
		for range waiters {
			go func() {
				defer func() { recover() }()
				_, ok := q.Get()
				results <- ok
			}()
		}
		synctest.Wait() // все пять уснули в Get
		within(t, time.Second, "Close", q.Close)
		synctest.Wait() // все, кого Close разбудил, уже вернулись
		if n := len(results); n != waiters {
			t.Fatalf("Close разбудил только %d из %d ждущих Get (Signal вместо Broadcast?)", n, waiters)
		}
		for range waiters {
			if <-results {
				t.Error("Get после Close пустой очереди вернул ok = true")
			}
		}
		within(t, time.Second, "Put/Close после Close", func() {
			if q.Put(9) {
				t.Error("Put после Close вернул true")
			}
			q.Close()
		})
	})
}

func TestQueueCloseDrainsRest(t *testing.T) {
	q := newQueue(t, 3)
	within(t, time.Second, "Put, Close, Get", func() {
		q.Put(1)
		q.Put(2)
		q.Close()
		for _, want := range []int{1, 2} {
			if v, ok := q.Get(); v != want || !ok {
				t.Errorf("Get() после Close = (%d, %v), want (%d, true): остаток надо отдать", v, ok, want)
			}
		}
		if v, ok := q.Get(); ok {
			t.Errorf("Get() на пустой закрытой = (%d, true), want (0, false)", v)
		}
	})
}

func TestQueueCloseWakesPut(t *testing.T) {
	bubble(t, func(t *testing.T) {
		q := newQueue(t, 1)
		within(t, time.Second, "Put", func() { q.Put(1) })
		result := make(chan bool, 1)
		go func() {
			defer func() { recover() }()
			result <- q.Put(2)
		}()
		synctest.Wait() // Put уснул на полной очереди
		within(t, time.Second, "Close", q.Close)
		synctest.Wait()
		select {
		case ok := <-result:
			if ok {
				t.Error("Put, разбуженный Close, вернул true")
			}
		default:
			t.Fatal("Close не разбудил ждущий Put")
		}
	})
}

// Много производителей и потребителей: всё положенное должно быть
// получено ровно один раз.
func TestQueueConcurrent(t *testing.T) {
	q := newQueue(t, 4)
	const producers, perProducer, consumers = 8, 250, 5
	got := make([][]int, consumers)
	within(t, 5*time.Second, "производители и потребители", func() {
		var prod, cons sync.WaitGroup
		for c := range consumers {
			cons.Go(func() {
				for {
					v, ok := q.Get()
					if !ok {
						return
					}
					got[c] = append(got[c], v)
				}
			})
		}
		for p := range producers {
			prod.Go(func() {
				for i := range perProducer {
					q.Put(p*perProducer + i)
				}
			})
		}
		prod.Wait()
		q.Close()
		cons.Wait()
	})
	all := slices.Concat(got...)
	slices.Sort(all)
	if len(all) != producers*perProducer {
		t.Fatalf("получено %d значений, want %d", len(all), producers*perProducer)
	}
	for i, v := range all {
		if v != i {
			t.Fatalf("значение %d потеряно или получено дважды", i)
		}
	}
}

// Get не должен возвращать ok = false, пока очередь не закрыта. Потребителей
// много, и пока разбуженный Get перехватывает мьютекс, значение может
// забрать другой: условие ожидания проверяют в цикле, а не одним if.
func TestQueueNoFalseBeforeClose(t *testing.T) {
	const rounds, consumers, items = 10, 16, 20000
	for round := range rounds {
		q := newQueue(t, 1)
		var closed atomic.Bool
		var early, received atomic.Int32
		within(t, 10*time.Second, "производитель и потребители", func() {
			var cons sync.WaitGroup
			for range consumers {
				cons.Go(func() {
					for {
						_, ok := q.Get()
						if !ok {
							if !closed.Load() {
								early.Add(1)
							}
							return
						}
						received.Add(1)
					}
				})
			}
			for i := range items {
				q.Put(i)
			}
			// Ждём, пока всё разберут, и только потом закрываем.
			for received.Load() < items && early.Load() < consumers {
				time.Sleep(time.Millisecond)
			}
			closed.Store(true)
			q.Close()
			cons.Wait()
		})
		if n := early.Load(); n > 0 {
			t.Fatalf("раунд %d: Get вернул ok = false до Close в %d потребителях: ожидание надо проверять в цикле for", round+1, n)
		}
	}
}
