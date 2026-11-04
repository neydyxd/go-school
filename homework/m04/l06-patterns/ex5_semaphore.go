package l06patterns

import "context"

// Semaphore ограничивает число одновременных владельцев слота.
// Реализуй на буферизованном канале.
type Semaphore struct {
	// TODO: поля
}

// NewSemaphore создаёт семафор на n слотов; n < 1 считать равным 1.
func NewSemaphore(n int) *Semaphore {
	panic("TODO: реализуй NewSemaphore")
}

// Acquire занимает слот, при необходимости ждёт освобождения.
// Если ctx отменят раньше, вернуть ctx.Err() и слот НЕ занимать.
func (s *Semaphore) Acquire(ctx context.Context) error {
	panic("TODO: реализуй Acquire")
}

// TryAcquire занимает слот без ожидания: true, если получилось.
func (s *Semaphore) TryAcquire() bool {
	panic("TODO: реализуй TryAcquire")
}

// Release освобождает слот. Release без занятого слота: паника с текстом
// "semaphore: release without acquire" (это баг вызывающего, молчать нельзя).
func (s *Semaphore) Release() {
	panic("TODO: реализуй Release")
}
