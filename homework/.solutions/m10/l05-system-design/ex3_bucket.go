package l05systemdesign

import (
	"sync"
	"time"
)

// TokenBucket: rate limiter «ведро с токенами».
type TokenBucket struct {
	rate  float64 // токенов в секунду
	burst float64 // вместимость
	now   func() time.Time

	mu     sync.Mutex
	tokens float64
	last   time.Time
}

// NewTokenBucket создаёт полное ведро на burst токенов, которое пополняется
// со скоростью rate токенов в секунду. now: часы (фейк в тестах).
func NewTokenBucket(rate float64, burst int, now func() time.Time) *TokenBucket {
	return &TokenBucket{rate: rate, burst: float64(burst), now: now, tokens: float64(burst), last: now()}
}

// Allow забирает один токен и возвращает true, если он был, иначе false.
//
//   - Пополнение непрерывное: за время t добавляется t*rate токенов, включая
//     дробные (два раза по полтокена дают один токен).
//   - В ведре никогда не больше burst: после часа простоя разрешено ровно
//     burst запросов подряд.
//   - Если часы пошли назад (перевели время), пополнения нет, токены не
//     сгорают, а новое время становится точкой отсчёта для следующего
//     пополнения.
//   - Безопасен для конкурентного использования.
func (b *TokenBucket) Allow() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	now := b.now()
	if elapsed := now.Sub(b.last); elapsed > 0 {
		b.tokens = min(b.burst, b.tokens+elapsed.Seconds()*b.rate)
	}
	b.last = now
	if b.tokens < 1 {
		return false
	}
	b.tokens--
	return true
}
