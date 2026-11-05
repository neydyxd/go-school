package l05auth

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestKeyedLimiter(t *testing.T) {
	l := NewKeyedLimiter(1, 2) // 1 токен в секунду, ведро на 2
	steps := []struct {
		name string
		key  string
		at   time.Duration // смещение от testNow
		want bool
	}{
		{"A: первый", "A", 0, true},
		{"A: второй (burst)", "A", 0, true},
		{"A: третий сразу", "A", 0, false},
		{"B не страдает от A", "B", 0, true},
		{"A через полсекунды", "A", 500 * time.Millisecond, false},
		{"A через секунду", "A", time.Second, true},
		{"A ещё раз в ту же секунду", "A", time.Second, false},
		{"B второй (burst)", "B", time.Second, true},
	}
	for _, s := range steps {
		if got := l.Allow(s.key, testNow.Add(s.at)); got != s.want {
			t.Fatalf("%s: Allow(%q, +%v) = %v, want %v", s.name, s.key, s.at, got, s.want)
		}
	}
}

func TestKeyedLimiterConcurrent(t *testing.T) {
	l := NewKeyedLimiter(1, 5)
	var allowed atomic.Int64
	var wg sync.WaitGroup
	for i := range 100 {
		wg.Go(func() {
			key := "same"
			if i%2 == 0 {
				key = "other"
			}
			if l.Allow(key, testNow) {
				allowed.Add(1)
			}
		})
	}
	wg.Wait()
	if got := allowed.Load(); got != 10 {
		t.Errorf("100 одновременных запросов по двум ключам с burst 5: пропущено %d, want 10", got)
	}
}

func TestRateLimitMiddleware(t *testing.T) {
	now := testNow
	l := NewKeyedLimiter(0.5, 2) // токен раз в 2 секунды
	r := gin.New()
	reached := 0
	r.GET("/login", RateLimit(l, KeyByUserOrIP, func() time.Time { return now }), func(c *gin.Context) {
		reached++
		c.Status(http.StatusNoContent)
	})
	do := func(ip string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodGet, "/login", nil)
		req.RemoteAddr = ip + ":40000"
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		return w
	}

	for i := range 2 {
		if w := do("203.0.113.7"); w.Code != http.StatusNoContent {
			t.Fatalf("запрос %d с 203.0.113.7: код %d, want 204", i+1, w.Code)
		}
	}
	w := do("203.0.113.7")
	if w.Code != http.StatusTooManyRequests {
		t.Fatalf("третий запрос с 203.0.113.7: код %d, want 429", w.Code)
	}
	if got := w.Header().Get("Retry-After"); got != "2" {
		t.Errorf("Retry-After = %q, want \"2\" (ceil(1/0.5))", got)
	}
	if got := w.Body.String(); got != `{"error":"too many requests"}` {
		t.Errorf("тело 429 = %s, want {\"error\":\"too many requests\"}", got)
	}
	if reached != 2 {
		t.Errorf("обработчик вызван %d раз, want 2: на 429 цепочку надо прервать", reached)
	}
	if w := do("198.51.100.1"); w.Code != http.StatusNoContent {
		t.Errorf("другой IP: код %d, want 204: лимит должен быть отдельным на каждый ключ", w.Code)
	}
	now = now.Add(2 * time.Second)
	if w := do("203.0.113.7"); w.Code != http.StatusNoContent {
		t.Errorf("203.0.113.7 через 2 секунды: код %d, want 204", w.Code)
	}
}

func TestKeyByUserOrIP(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	c.Request.RemoteAddr = "192.0.2.10:5555"
	if got := KeyByUserOrIP(c); got != "ip:192.0.2.10" {
		t.Errorf("KeyByUserOrIP(аноним) = %q, want \"ip:192.0.2.10\"", got)
	}
	c.Set(userKey, User{ID: "u7", Role: "user"})
	if got := KeyByUserOrIP(c); got != "user:u7" {
		t.Errorf("KeyByUserOrIP(u7) = %q, want \"user:u7\"", got)
	}
}
