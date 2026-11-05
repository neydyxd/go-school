package l09redis

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func testCtx(t *testing.T) context.Context {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	t.Cleanup(cancel)
	return ctx
}

// fakeDB это «основное хранилище» товаров со счётчиком обращений.
type fakeDB struct {
	mu       sync.Mutex
	products map[int64]Product
	calls    atomic.Int64
	release  chan struct{} // если не nil, загрузка ждёт, пока канал не закроют
	onLoad   func()        // если не nil, вызывается в начале каждой загрузки
}

func newFakeDB() *fakeDB {
	return &fakeDB{products: map[int64]Product{
		1:  {ID: 1, Name: "Чайник", Price: 349900},
		42: {ID: 42, Name: "Кружка", Price: 59900},
	}}
}

func (f *fakeDB) load(ctx context.Context, id int64) (Product, error) {
	f.calls.Add(1)
	if f.onLoad != nil {
		f.onLoad()
	}
	if f.release != nil {
		select {
		case <-f.release:
		case <-ctx.Done():
			return Product{}, ctx.Err()
		}
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	p, ok := f.products[id]
	if !ok {
		return Product{}, ErrNotFound
	}
	return p, nil
}

func (f *fakeDB) set(p Product) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.products[p.ID] = p
}
