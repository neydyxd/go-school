package l04sync

import (
	"errors"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestMemoizeBasic(t *testing.T) {
	var calls atomic.Int32
	errBad := errors.New("плохой ключ")
	f := func(key string) (int, error) {
		calls.Add(1)
		if key == "bad" {
			return 0, errBad
		}
		return strconv.Atoi(key)
	}
	var get func(string) (int, error)
	within(t, time.Second, "Memoize", func() { get = Memoize(f) })

	tests := []struct {
		name      string
		key       string
		want      int
		wantErr   error
		wantCalls int32
	}{
		{"первый вызов", "42", 42, nil, 1},
		{"тот же ключ из кеша", "42", 42, nil, 1},
		{"другой ключ", "7", 7, nil, 2},
		{"ошибка", "bad", 0, errBad, 3},
		{"ошибка тоже кешируется", "bad", 0, errBad, 3},
		{"пустой ключ тоже ключ", "", 0, strconv.ErrSyntax, 4},
		{"пустой ключ из кеша", "", 0, strconv.ErrSyntax, 4},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got int
			var err error
			within(t, time.Second, "вызов мемоизированной функции", func() { got, err = get(tt.key) })
			if got != tt.want || !errors.Is(err, tt.wantErr) {
				t.Errorf("get(%q) = (%d, %v), want (%d, %v)", tt.key, got, err, tt.want, tt.wantErr)
			}
			if c := calls.Load(); c != tt.wantCalls {
				t.Errorf("после get(%q) f вызвана %d раз, want %d", tt.key, c, tt.wantCalls)
			}
		})
	}
}

// 50 горутин одновременно просят один ключ, пока f ещё считает.
func TestMemoizeConcurrentSameKey(t *testing.T) {
	var calls atomic.Int32
	release := make(chan struct{})
	f := func(key string) (int, error) {
		calls.Add(1)
		<-release
		return len(key), nil
	}
	var get func(string) (int, error)
	within(t, time.Second, "Memoize", func() { get = Memoize(f) })

	const n = 50
	results := make([]int, n)
	var wg sync.WaitGroup
	for i := range n {
		wg.Go(func() {
			defer func() { recover() }() // заготовка паникует, проверим ниже
			results[i], _ = get("hello")
		})
	}
	time.Sleep(50 * time.Millisecond) // все успевают встать в очередь
	close(release)
	within(t, 2*time.Second, "ожидание всех вызовов", wg.Wait)

	if c := calls.Load(); c != 1 {
		t.Errorf("f вызвана %d раз для одного ключа, want 1", c)
	}
	for i, r := range results {
		if r != 5 {
			t.Fatalf("горутина %d получила %d, want 5", i, r)
		}
	}
}

// Медленный ключ не должен блокировать быстрый.
func TestMemoizeKeysIndependent(t *testing.T) {
	release := make(chan struct{})
	defer close(release)
	slowStarted := make(chan struct{})
	f := func(key string) (int, error) {
		if key == "slow" {
			close(slowStarted)
			<-release
		}
		return 1, nil
	}
	var get func(string) (int, error)
	within(t, time.Second, "Memoize", func() { get = Memoize(f) })

	go func() {
		defer func() { recover() }()
		get("slow")
	}()
	select {
	case <-slowStarted:
	case <-time.After(time.Second):
		t.Fatal("f для медленного ключа так и не вызвана")
	}
	within(t, time.Second, "get(\"fast\") пока считается \"slow\" (мьютекс держится во время f?)", func() { get("fast") })
}
