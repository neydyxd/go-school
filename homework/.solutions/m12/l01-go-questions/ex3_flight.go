package l01goquestions

import "sync"

// Flight схлопывает одновременные вызовы с одинаковым ключом.
type Flight struct {
	mu    sync.Mutex
	calls map[string]*call
}

type call struct {
	done chan struct{} // закрывается, когда v и err записаны
	v    string
	err  error
}

// Do выполняет fn один раз на группу одновременных вызовов с одним key.
func (f *Flight) Do(key string, fn func() (string, error)) (v string, err error, shared bool) {
	f.mu.Lock()
	if c, ok := f.calls[key]; ok {
		f.mu.Unlock()
		<-c.done // close даёт happens-before: v и err уже записаны
		return c.v, c.err, true
	}
	if f.calls == nil {
		f.calls = map[string]*call{}
	}
	c := &call{done: make(chan struct{})}
	f.calls[key] = c
	f.mu.Unlock() // fn работает без мьютекса: другие ключи не ждут

	c.v, c.err = fn()

	f.mu.Lock()
	delete(f.calls, key)
	f.mu.Unlock()
	close(c.done)
	return c.v, c.err, false
}
