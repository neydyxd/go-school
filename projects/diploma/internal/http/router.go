// Package http это транспортный слой. Разбирает запрос, зовёт сервис, переводит
// результат и ошибки в HTTP. Бизнес-правил здесь нет.
package http

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Check проверяет одну зависимость для /readyz. pool.Ping подходит без обёрток.
type Check func(ctx context.Context) error

// Deps содержит всё, что роутеру нужно снаружи. Сервисы своей темы добавляй сюда по интерфейсам.
type Deps struct {
	Log      *slog.Logger
	Registry *prometheus.Registry // свой реестр, а не глобальный: тесты не падают на повторной регистрации
	Checks   map[string]Check     // "postgres": pool.Ping, "redis": ...
	Draining *atomic.Bool         // true после SIGTERM: /readyz сразу отвечает 503
}

// NewRouter собирает middleware, служебные эндпоинты и маршруты темы.
func NewRouter(d Deps) *gin.Engine {
	r := gin.New()
	_ = r.SetTrustedProxies(nil) // ClientIP не верит X-Forwarded-For от кого попало
	// Паника: значение и стек в JSON-лог запроса, клиенту 500 в едином формате.
	// io.Discard: без него gin печатает тот же стек ещё и текстом в stderr.
	r.Use(requestLogger(d.Log), metrics(d.Registry), gin.CustomRecoveryWithWriter(io.Discard, func(c *gin.Context, rec any) {
		writeError(c, fmt.Errorf("panic: %v\n%s", rec, debug.Stack())) // одна строка лога с err и стеком
	}))
	r.NoRoute(func(c *gin.Context) {
		fail(c, http.StatusNotFound, "not_found", "маршрут не найден", nil)
	})

	r.GET("/healthz", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })
	r.GET("/readyz", readiness(d.Checks, d.Draining, 800*time.Millisecond))
	r.GET("/metrics", gin.WrapH(promhttp.HandlerFor(d.Registry, promhttp.HandlerOpts{})))

	// Маршруты темы. Хендлеры держи в отдельных файлах (bookings.go, flags.go, orders.go),
	// сервисы получай через Deps по интерфейсам, объявленным здесь, у потребителя:
	//
	//	v1 := r.Group("/api/v1")
	//	v1.POST("/bookings", h.createBooking)
	return r
}

// readiness: liveness (/healthz) про сам процесс, а здесь проверяем зависимости.
// Параллельно и с общим таймаутом меньше секунды: kubelet по умолчанию ждёт ответа
// секунду (timeoutSeconds: 1), и ответ должен успеть прийти до этого.
func readiness(checks map[string]Check, draining *atomic.Bool, timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		if draining.Load() {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "draining"})
			return
		}
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()

		log := loggerFrom(c) // gin.Context в горутины не отдаём
		var (
			mu  sync.Mutex
			wg  sync.WaitGroup
			res = make(map[string]string, len(checks))
		)
		for name, check := range checks {
			wg.Go(func() {
				status := "ok"
				if err := check(ctx); err != nil {
					status = "fail"
					log.Warn("readyz: зависимость недоступна", "dep", name, "err", err)
				}
				mu.Lock()
				res[name] = status
				mu.Unlock()
			})
		}
		wg.Wait()

		code, status := http.StatusOK, "ready"
		for _, s := range res {
			if s != "ok" {
				code, status = http.StatusServiceUnavailable, "not_ready"
			}
		}
		c.JSON(code, gin.H{"status": status, "checks": res}) // текст ошибки только в лог
	}
}
