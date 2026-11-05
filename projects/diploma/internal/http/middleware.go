package http

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"regexp"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
)

type loggerKey struct{}

var serviceRoutes = map[string]bool{"/healthz": true, "/readyz": true, "/metrics": true}

var validRequestID = regexp.MustCompile(`^[A-Za-z0-9-]{1,64}$`)

// requestLogger: request_id (свой или из заголовка), логгер запроса в контексте
// и итоговая строка лога на каждый запрос.
func requestLogger(base *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		id := c.GetHeader("X-Request-ID")
		if !validRequestID.MatchString(id) {
			b := make([]byte, 16)
			_, _ = rand.Read(b)
			id = hex.EncodeToString(b)
		}
		c.Header("X-Request-ID", id)
		l := base.With("request_id", id)
		c.Request = c.Request.WithContext(context.WithValue(c.Request.Context(), loggerKey{}, l))
		c.Next()

		level := slog.LevelInfo
		switch {
		case serviceRoutes[c.FullPath()]:
			// HEALTHCHECK и Prometheus дёргают их постоянно, а 503 от /readyz при остановке
			// это штатное состояние: в info и error это шум.
			level = slog.LevelDebug
		case c.Writer.Status() >= 500:
			level = slog.LevelError
		}
		l.Log(c.Request.Context(), level, "http request",
			"method", c.Request.Method, "path", c.Request.URL.Path, // без query: там бывают токены
			"status", c.Writer.Status(), "duration_ms", time.Since(start).Milliseconds())
	}
}

func loggerFrom(c *gin.Context) *slog.Logger {
	if l, ok := c.Request.Context().Value(loggerKey{}).(*slog.Logger); ok {
		return l
	}
	return slog.Default()
}

// metrics: RED-метрики (rate, errors, duration) на каждый маршрут.
// Метка route это шаблон Gin (/api/v1/rooms/:id), а не сырой путь: иначе
// каждый id станет отдельным временным рядом и Prometheus захлебнётся.
func metrics(reg prometheus.Registerer) gin.HandlerFunc {
	requests := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "HTTP-запросы по маршруту, методу и коду ответа.",
	}, []string{"method", "route", "code"})
	duration := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "http_request_duration_seconds",
		Help:    "Длительность обработки HTTP-запроса.",
		Buckets: prometheus.DefBuckets,
	}, []string{"method", "route"})
	reg.MustRegister(requests, duration)

	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		route := c.FullPath()
		if route == "" {
			route = "unmatched"
		}
		requests.WithLabelValues(c.Request.Method, route, strconv.Itoa(c.Writer.Status())).Inc()
		duration.WithLabelValues(c.Request.Method, route).Observe(time.Since(start).Seconds())
	}
}
