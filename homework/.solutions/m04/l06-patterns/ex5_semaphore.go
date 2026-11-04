package l06patterns

import "context"

// Semaphore ограничивает число одновременных владельцев слота.
// Реализуй на буферизованном канале.
type Semaphore struct {
	slots chan struct{}
}

// NewSemaphore создаёт семафор на n слотов; n < 1 считать равным 1.
func NewSemaphore(n int) *Semaphore {
	return &Semaphore{slots: make(chan struct{}, max(n, 1))}
}

// Acquire занимает слот, при необходимости ждёт освобождения.
// Если ctx отменят раньше, вернуть ctx.Err() и слот НЕ занимать.
func (s *Semaphore) Acquire(ctx context.Context) error {
	select {
	case s.slots <- struct{}{}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// TryAcquire занимает слот без ожидания: true, если получилось.
func (s *Semaphore) TryAcquire() bool {
	select {
	case s.slots <- struct{}{}:
		return true
	default:
		return false
	}
}

// Release освобождает слот. Release без занятого слота: паника с текстом
// "semaphore: release without acquire" (это баг вызывающего, молчать нельзя).
func (s *Semaphore) Release() {
	select {
	case <-s.slots:
	default:
		panic("semaphore: release without acquire")
	}
}
