package l05auth

import (
	"math"
	"net/http"
	"strconv"
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
	l.mu.Lock()
	lim, ok := l.limiters[key]
	if !ok {
		// ponytail: карта растёт без очистки; в проде вытесняй старые ключи (LRU/TTL) или держи лимиты в Redis
		lim = rate.NewLimiter(l.r, l.burst)
		l.limiters[key] = lim
	}
	l.mu.Unlock()
	return lim.AllowN(now, 1)
}

// KeyByUserOrIP: "user:<id>", если в контексте есть пользователь (CurrentUser),
// иначе "ip:<c.ClientIP()>".
func KeyByUserOrIP(c *gin.Context) string {
	if u, ok := CurrentUser(c); ok {
		return "user:" + u.ID
	}
	return "ip:" + c.ClientIP()
}

// RateLimit это middleware: ключ запроса считает key, время берётся из now().
// Лимит исчерпан: прервать цепочку, ответить 429 {"error":"too many requests"}
// и заголовком Retry-After в целых секундах: ceil(1/r), но не меньше 1.
func RateLimit(l *KeyedLimiter, key func(*gin.Context) string, now func() time.Time) gin.HandlerFunc {
	retry := strconv.Itoa(max(1, int(math.Ceil(1/float64(l.r)))))
	return func(c *gin.Context) {
		if !l.Allow(key(c), now()) {
			c.Header("Retry-After", retry)
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "too many requests"})
			return
		}
		c.Next()
	}
}
