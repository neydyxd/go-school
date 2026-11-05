package l07linkedlists

import "testing"

func TestLRU(t *testing.T) {
	c := NewLRU(2)
	expect := func(key, want int, wantOK bool) {
		t.Helper()
		if v, ok := c.Get(key); v != want || ok != wantOK {
			t.Fatalf("Get(%d) = %d, %v, want %d, %v", key, v, ok, want, wantOK)
		}
	}
	expect(1, 0, false)
	c.Put(1, 1)
	c.Put(2, 2)
	expect(1, 1, true) // теперь 1 свежее 2
	c.Put(3, 3)        // вытесняет 2
	expect(2, 0, false)
	expect(3, 3, true)
	expect(1, 1, true)
	c.Put(4, 4) // вытесняет 3: 1 читали позже
	expect(3, 0, false)
	expect(4, 4, true)
	expect(1, 1, true)
	if c.Len() != 2 {
		t.Fatalf("Len() = %d, want 2", c.Len())
	}
	// обновление существующего ключа никого не вытесняет и освежает ключ
	c.Put(4, 40)
	if c.Len() != 2 {
		t.Fatalf("Len() после обновления = %d, want 2", c.Len())
	}
	c.Put(5, 5) // вытесняет 1: 4 обновили позже
	expect(1, 0, false)
	expect(4, 40, true)
	expect(5, 5, true)

	one := NewLRU(1)
	one.Put(1, 1)
	one.Put(2, 2)
	if _, ok := one.Get(1); ok {
		t.Fatal("NewLRU(1): после Put(2) ключ 1 должен быть вытеснен")
	}
	if v, ok := one.Get(2); !ok || v != 2 {
		t.Fatalf("NewLRU(1).Get(2) = %d, %v, want 2, true", v, ok)
	}
}

func TestLRUBig(t *testing.T) {
	const capacity, ops = 100_000, 300_000
	c := NewLRU(capacity)
	failed := ""
	within(t, func() {
		for i := range ops {
			c.Put(i, i)
			if i >= capacity {
				// самый старый живой ключ i-capacity+1 ещё на месте, i-capacity уже вытеснен
				if _, ok := c.Get(i - capacity); ok {
					failed = "ключ должен был быть вытеснен"
					return
				}
			}
		}
	})
	if failed != "" {
		t.Fatal(failed)
	}
	if c.Len() != capacity {
		t.Fatalf("Len() = %d, want %d", c.Len(), capacity)
	}
}

func TestLRUHotGets(t *testing.T) {
	// Get живых ключей из середины порядка: перенос ключа в начало должен
	// быть O(1). Слайс порядка с поиском и удалением ключа здесь O(capacity)
	// на каждый Get и не уложится.
	const capacity, ops = 100_000, 300_000
	c := NewLRU(capacity)
	for i := range capacity {
		c.Put(i, i)
	}
	bad := -1
	within(t, func() {
		for i := range ops {
			k := i * 7919 % capacity
			if v, ok := c.Get(k); !ok || v != k {
				bad = k
				return
			}
		}
	})
	if bad >= 0 {
		t.Fatalf("Get(%d) не нашёл живой ключ или вернул не то значение", bad)
	}
}
