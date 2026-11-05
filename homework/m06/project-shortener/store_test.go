package shortener

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// within выполняет f и валит тест, если она не вернулась за 2 секунды:
// неверное решение (бесконечный цикл, взаимоблокировка на мьютексе)
// не должно вешать go test до общего таймаута в 10 минут.
func within(t *testing.T, what string, f func()) {
	t.Helper()
	done := make(chan struct{})
	go func() {
		defer close(done)
		f()
	}()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatalf("%s не вернулся за 2с: бесконечный цикл или взаимоблокировка на мьютексе?", what)
	}
}

func create(t *testing.T, s *Store, url string) (l Link, err error) {
	t.Helper()
	within(t, "Create("+url+")", func() { l, err = s.Create(url) })
	return l, err
}

func resolve(t *testing.T, s *Store, code string) (u string, err error) {
	t.Helper()
	within(t, "Resolve("+code+")", func() { u, err = s.Resolve(code) })
	return u, err
}

func stats(t *testing.T, s *Store, code string) (l Link, err error) {
	t.Helper()
	within(t, "Stats("+code+")", func() { l, err = s.Stats(code) })
	return l, err
}

var t0 = time.Date(2026, 11, 5, 10, 0, 0, 0, time.UTC)

func fixedNow() time.Time { return t0 }

// seq выдаёт коды по списку, потом c1, c2, ...
func seq(codes ...string) func() string {
	var mu sync.Mutex
	i := 0
	return func() string {
		mu.Lock()
		defer mu.Unlock()
		i++
		if i <= len(codes) {
			return codes[i-1]
		}
		return fmt.Sprintf("c%d", i)
	}
}

func TestStoreBasic(t *testing.T) {
	s := NewStore(seq("abc", "def"), fixedNow)
	l, err := create(t, s, "https://go.dev")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	want := Link{Code: "abc", URL: "https://go.dev", CreatedAt: t0}
	if l != want {
		t.Errorf("Create = %+v, want %+v", l, want)
	}
	l2, _ := create(t, s, "https://go.dev")
	if l2.Code != "def" {
		t.Errorf("повторный Create того же url: код %q, want новый \"def\"", l2.Code)
	}

	for range 3 {
		u, err := resolve(t, s, "abc")
		if err != nil || u != "https://go.dev" {
			t.Fatalf("Resolve(abc) = %q, %v; want https://go.dev", u, err)
		}
	}
	st, err := stats(t, s, "abc")
	if err != nil || st.Hits != 3 {
		t.Errorf("Stats(abc) = %+v, %v; want Hits 3", st, err)
	}
	st, _ = stats(t, s, "abc")
	if st.Hits != 3 {
		t.Errorf("Stats изменил счётчик: %d, want 3", st.Hits)
	}
	if st2, _ := stats(t, s, "def"); st2.Hits != 0 {
		t.Errorf("Stats(def).Hits = %d, want 0: счётчики у кодов свои", st2.Hits)
	}

	st.Hits = 100
	if again, _ := stats(t, s, "abc"); again.Hits != 3 {
		t.Errorf("изменение результата Stats повлияло на хранилище: Hits %d, want 3", again.Hits)
	}
}

func TestStoreNotFound(t *testing.T) {
	s := NewStore(seq(), fixedNow)
	if _, err := resolve(t, s, "nope"); !errors.Is(err, ErrNotFound) {
		t.Errorf("Resolve(nope) = %v, want ErrNotFound", err)
	}
	if _, err := stats(t, s, "nope"); !errors.Is(err, ErrNotFound) {
		t.Errorf("Stats(nope) = %v, want ErrNotFound", err)
	}
}

func TestStoreCollisions(t *testing.T) {
	s := NewStore(seq("dup", "dup", "dup", "fresh"), fixedNow)
	create(t, s, "https://a.ru")
	l, err := create(t, s, "https://b.ru")
	if err != nil || l.Code != "fresh" {
		t.Fatalf("Create при коллизиях = %+v, %v; want код fresh", l, err)
	}
	if u, _ := resolve(t, s, "dup"); u != "https://a.ru" {
		t.Errorf("коллизия перезаписала чужую ссылку: Resolve(dup) = %q, want https://a.ru", u)
	}

	always := NewStore(func() string { return "same" }, fixedNow)
	create(t, always, "https://a.ru")
	if _, err := create(t, always, "https://b.ru"); !errors.Is(err, ErrCodeExhausted) {
		t.Errorf("Create, когда gen всегда даёт занятый код = %v, want ErrCodeExhausted (и без бесконечного цикла)", err)
	}
	var calls atomic.Int64
	counting := NewStore(func() string { calls.Add(1); return "x" }, fixedNow)
	create(t, counting, "https://a.ru")
	calls.Store(0)
	create(t, counting, "https://b.ru")
	if n := calls.Load(); n != 5 {
		t.Errorf("при сплошных коллизиях gen вызван %d раз, want 5", n)
	}
}

func TestStoreConcurrent(t *testing.T) {
	var n atomic.Int64
	s := NewStore(func() string { return fmt.Sprintf("k%d", n.Add(1)) }, fixedNow)
	l, _ := create(t, s, "https://hot.example")

	var wg sync.WaitGroup
	for range 50 {
		wg.Go(func() {
			for range 20 {
				s.Resolve(l.Code)
				s.Stats(l.Code)
			}
		})
		wg.Go(func() { s.Create("https://other.example") })
	}
	within(t, "wg.Wait для параллельных Create/Resolve/Stats", wg.Wait)
	st, _ := stats(t, s, l.Code)
	if st.Hits != 1000 {
		t.Errorf("после 1000 параллельных Resolve Hits = %d, want 1000: потерянные инкременты", st.Hits)
	}
}
