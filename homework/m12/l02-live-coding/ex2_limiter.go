package l02livecoding

import (
	"context"
	"time"
)

// Limiter: ограничитель частоты по алгоритму token bucket.
// В ведре до burst жетонов, изначально оно полное. Жетоны пополняются
// непрерывно: один жетон за каждый период every (за полпериода накапливается
// полжетона, прогресс не теряется между вызовами). Больше burst не копится.
// Методы безопасны для одновременного вызова из многих горутин.
type Limiter struct {
	// добавь нужные поля
}

// NewLimiter создаёт ограничитель. burst < 1 считать равным 1.
// every <= 0 значит «без ограничения»: Allow всегда true, Wait сразу nil.
func NewLimiter(every time.Duration, burst int) *Limiter {
	panic("TODO: реализуй NewLimiter")
}

// Allow забирает жетон, если он есть, и возвращает true; иначе false, не блокируясь.
func (l *Limiter) Allow() bool {
	panic("TODO: реализуй Allow")
}

// Wait ждёт жетон и забирает его. Если ctx отменился раньше, возвращает ctx.Err()
// и жетон не тратит (следующий жетон достанется другим).
func (l *Limiter) Wait(ctx context.Context) error {
	panic("TODO: реализуй Wait")
}
