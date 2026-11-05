package l02livecoding

import "sync"

// Lazy: однократная инициализация с повтором после ошибки.
type Lazy[T any] struct {
	mu   sync.Mutex
	done bool
	v    T
	init func() (T, error)
}

// NewLazy создаёт обёртку.
func NewLazy[T any](init func() (T, error)) *Lazy[T] {
	return &Lazy[T]{init: init}
}

// Get возвращает значение, при необходимости вызывая init.
// ponytail: мьютекс держится во время init, поэтому ждущие проверяют done по очереди;
// для горячего пути после инициализации можно добавить atomic.Bool перед Lock
func (l *Lazy[T]) Get() (T, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.done {
		return l.v, nil
	}
	v, err := l.init()
	if err != nil {
		var zero T
		return zero, err
	}
	l.v, l.done = v, true
	return v, nil
}
