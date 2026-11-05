package l02livecoding

import (
	"context"
	"sync"
)

// Group: свой errgroup с ограничением параллелизма.
type Group struct {
	wg     sync.WaitGroup
	sem    chan struct{} // nil: без лимита
	once   sync.Once
	err    error
	cancel context.CancelFunc
}

// WithContext возвращает группу и производный контекст.
func WithContext(ctx context.Context) (*Group, context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	return &Group{cancel: cancel}, ctx
}

// SetLimit ограничивает число одновременно работающих задач.
func (g *Group) SetLimit(n int) {
	if n <= 0 {
		g.sem = nil
		return
	}
	g.sem = make(chan struct{}, n)
}

// Go запускает f; блокируется, если слоты заняты.
func (g *Group) Go(f func() error) {
	if g.sem != nil {
		g.sem <- struct{}{}
	}
	g.wg.Go(func() {
		defer func() {
			if g.sem != nil {
				<-g.sem
			}
		}()
		if err := f(); err != nil {
			g.once.Do(func() { // первая ошибка побеждает, отмена сразу
				g.err = err
				if g.cancel != nil {
					g.cancel()
				}
			})
		}
	})
}

// Wait ждёт все задачи и возвращает первую ошибку.
func (g *Group) Wait() error {
	g.wg.Wait() // happens-before: запись g.err в задачах видна после Wait
	if g.cancel != nil {
		g.cancel()
	}
	return g.err
}
