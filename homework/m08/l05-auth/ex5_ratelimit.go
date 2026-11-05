package l05auth

import (
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// KeyedLimiter держит отдельный token bucket на каждый ключ (IP, пользователя).
type KeyedLimiter struct {
	mu       sync.Mutex
	limiters map[string]*rate.Limiter
	r        rate.Limit
	burst    int
}

// NewKeyedLimiter создаёт лимитер: r токенов в секунду, ёмкость ведра burst.
func NewKeyedLimiter(r rate.Limit, burst int) *KeyedLimiter {
	return &KeyedLimiter{limiters: make(map[string]*rate.Limiter), r: r, burst: burst}
}

// Allow тратит один токен из ведра ключа key в момент now. У каждого ключа
// своё ведро: клиент A, исчерпавший лимит, не мешает клиенту B.
// Время передаётся явно (rate.Limiter.AllowN), чтобы тесты не зависели от часов.
// Метод вызывается из многих горутин одновременно.
func (l *KeyedLimiter) Allow(key string, now time.Time) bool {
	panic("TODO: реализуй Allow")
}

// KeyByUserOrIP: "user:<id>", если в контексте есть пользователь (CurrentUser),
// иначе "ip:<c.ClientIP()>".
func KeyByUserOrIP(c *gin.Context) string {
	panic("TODO: реализуй KeyByUserOrIP")
}

// RateLimit это middleware: ключ запроса считает key, время берётся из now().
// Лимит исчерпан: прервать цепочку, ответить 429 {"error":"too many requests"}
// и заголовком Retry-After в целых секундах: ceil(1/r), но не меньше 1.
func RateLimit(l *KeyedLimiter, key func(*gin.Context) string, now func() time.Time) gin.HandlerFunc {
	panic("TODO: реализуй RateLimit")
}
