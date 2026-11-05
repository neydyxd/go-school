package l02livecoding

import (
	"fmt"
	"sync"
	"testing"
	"testing/synctest"
	"time"
)

func TestCacheTTL(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		c := NewCache[string, int](time.Minute)
		if v, ok := c.Get("a"); ok || v != 0 {
			t.Fatalf(`Get("a") в пустом кеше = %v, %v; want 0, false`, v, ok)
		}
		c.Set("a", 1)
		c.Set("b", 2)
		time.Sleep(time.Minute - time.Millisecond)
		if v, ok := c.Get("a"); !ok || v != 1 {
			t.Fatalf(`Get("a") за 1ms до истечения = %v, %v; want 1, true`, v, ok)
		}
		c.Set("b", 20) // перезапись продлевает жизнь
		time.Sleep(time.Millisecond)
		// Сначала Cleanup: решение может удалять протухшее и лениво в Get или Len,
		// поэтому до Cleanup протухшие ключи не трогаем.
		if n := c.Cleanup(); n != 1 {
			t.Errorf("Cleanup() ровно через ttl = %d, want 1 (протух только a)", n)
		}
		if n := c.Cleanup(); n != 0 {
			t.Errorf("повторный Cleanup() = %d, want 0: протухшие должны удаляться из памяти", n)
		}
		if v, ok := c.Get("a"); ok || v != 0 {
			t.Errorf(`Get("a") после истечения = %v, %v; want 0, false`, v, ok)
		}
		if v, ok := c.Get("b"); !ok || v != 20 {
			t.Errorf(`Get("b") после перезаписи = %v, %v; want 20, true: Set должен продлевать срок`, v, ok)
		}
		if n := c.Len(); n != 1 {
			t.Errorf("Len() = %d, want 1", n)
		}
		c.Delete("b")
		c.Delete("нет такого")
		if _, ok := c.Get("b"); ok || c.Len() != 0 {
			t.Errorf("после Delete запись осталась: Len() = %d", c.Len())
		}
	})
}

func TestCacheExactTTL(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		c := NewCache[string, int](time.Minute)
		c.Set("x", 1)
		c.Set("y", 2)
		time.Sleep(time.Minute)
		if v, ok := c.Get("x"); ok || v != 0 {
			t.Errorf(`Get("x") ровно через ttl = %v, %v; want 0, false`, v, ok)
		}
		if n := c.Len(); n != 0 {
			t.Errorf("Len() ровно через ttl = %d, want 0: протухшие записи не считаются", n)
		}
	})
}

func TestCacheNoTTL(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		c := NewCache[int, string](0)
		c.Set(1, "x")
		time.Sleep(1000 * time.Hour)
		if v, ok := c.Get(1); !ok || v != "x" || c.Len() != 1 || c.Cleanup() != 0 {
			t.Errorf("ttl 0: через 1000h Get(1) = %q, %v, Len %d; want x, true, 1", v, ok, c.Len())
		}
	})
}

func TestCacheConcurrent(t *testing.T) {
	c := NewCache[string, int](time.Hour)
	var wg sync.WaitGroup
	for g := range 20 {
		wg.Go(func() {
			for i := range 200 {
				k := fmt.Sprint("k", i%10)
				c.Set(k, g)
				c.Get(k)
				c.Len()
				if i%50 == 0 {
					c.Cleanup()
					c.Delete(k)
				}
			}
		})
	}
	wg.Wait()
	if n := c.Len(); n < 1 || n > 10 {
		t.Errorf("после конкурентной работы Len() = %d, want 1..10", n)
	}
}
