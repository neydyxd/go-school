package l03internalsperf

import "sync"

// MiniChan: notEmpty будит получателей, notFull отправителей. В hchan вместо
// Cond стоят очереди sudog, и рантайм будит ровно одну нужную горутину,
// а при закрытии всех. Здесь Signal будит одну, Broadcast всех.
type MiniChan[T any] struct {
	mu       sync.Mutex
	notEmpty *sync.Cond
	notFull  *sync.Cond
	ring     *Ring[T]
	closed   bool
}

// NewMiniChan создаёт канал ёмкости capacity.
func NewMiniChan[T any](capacity int) *MiniChan[T] {
	if capacity < 1 {
		panic("NewMiniChan: ёмкость должна быть не меньше 1")
	}
	c := &MiniChan[T]{ring: NewRing[T](capacity)}
	c.notEmpty = sync.NewCond(&c.mu)
	c.notFull = sync.NewCond(&c.mu)
	return c
}

// Send ждёт места или закрытия.
func (c *MiniChan[T]) Send(v T) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for !c.closed && c.ring.Len() == c.ring.Cap() {
		c.notFull.Wait()
	}
	if c.closed {
		panic("send on closed MiniChan")
	}
	c.ring.Push(v)
	c.notEmpty.Signal()
}

// Recv ждёт значения или закрытия.
func (c *MiniChan[T]) Recv() (v T, ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for !c.closed && c.ring.Len() == 0 {
		c.notEmpty.Wait()
	}
	if c.ring.Len() == 0 {
		return v, false // закрыт и пуст
	}
	v, _ = c.ring.Pop()
	c.notFull.Signal()
	return v, true
}

// Close будит всех ждущих с обеих сторон.
func (c *MiniChan[T]) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		panic("close of closed MiniChan")
	}
	c.closed = true
	c.notEmpty.Broadcast()
	c.notFull.Broadcast()
}

// Len возвращает число значений в буфере.
func (c *MiniChan[T]) Len() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.ring.Len()
}
