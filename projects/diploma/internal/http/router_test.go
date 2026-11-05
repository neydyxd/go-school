package http

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
)

// Пример теста транспортного слоя: настоящий роутер, фейковые зависимости,
// httptest вместо сети. Так же тестируй свои хендлеры с fake-сервисами.
func TestServiceEndpoints(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ok := func(context.Context) error { return nil }
	down := func(context.Context) error { return errors.New("connection refused") }

	tests := []struct {
		name     string
		path     string
		checks   map[string]Check
		draining bool
		wantCode int
		wantBody string
	}{
		{"liveness не смотрит на базу", "/healthz", map[string]Check{"postgres": down}, false, 200, `"status":"ok"`},
		{"все зависимости живы", "/readyz", map[string]Check{"postgres": ok, "redis": ok}, false, 200, `"status":"ready"`},
		{"одна зависимость упала", "/readyz", map[string]Check{"postgres": ok, "redis": down}, false, 503, `"redis":"fail"`},
		{"ошибка базы не утекает клиенту", "/readyz", map[string]Check{"postgres": down}, false, 503, `"postgres":"fail"`},
		{"после SIGTERM не готов", "/readyz", map[string]Check{"postgres": ok}, true, 503, `"status":"draining"`},
		{"неизвестный маршрут в едином формате", "/nope", nil, false, 404, `"code":"not_found"`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			draining := &atomic.Bool{}
			draining.Store(tt.draining)
			r := NewRouter(Deps{
				Log:      slog.New(slog.DiscardHandler),
				Registry: prometheus.NewRegistry(),
				Checks:   tt.checks,
				Draining: draining,
			})
			w := httptest.NewRecorder()
			r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, tt.path, nil))

			if w.Code != tt.wantCode || !strings.Contains(w.Body.String(), tt.wantBody) {
				t.Errorf("GET %s = %d %s, want %d с %s", tt.path, w.Code, w.Body, tt.wantCode, tt.wantBody)
			}
			if strings.Contains(w.Body.String(), "connection refused") {
				t.Errorf("GET %s: текст ошибки зависимости попал в ответ: %s", tt.path, w.Body)
			}
		})
	}
}

func TestPanicIsInternalError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := NewRouter(Deps{Log: slog.New(slog.DiscardHandler), Registry: prometheus.NewRegistry(), Draining: &atomic.Bool{}})
	r.GET("/boom", func(*gin.Context) { panic("секрет из паники") })

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/boom", nil))
	if w.Code != 500 || !strings.Contains(w.Body.String(), `"code":"internal"`) || strings.Contains(w.Body.String(), "секрет") {
		t.Errorf("GET /boom = %d %s, want 500 internal без текста паники", w.Code, w.Body)
	}
}

func TestMetricsUseRouteTemplate(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := NewRouter(Deps{Log: slog.New(slog.DiscardHandler), Registry: prometheus.NewRegistry(), Draining: &atomic.Bool{}})
	r.GET("/api/v1/items/:id", func(c *gin.Context) { c.Status(http.StatusNoContent) })

	for _, p := range []string{"/api/v1/items/1", "/api/v1/items/2"} {
		r.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, p, nil))
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	body, _ := io.ReadAll(w.Body)

	want := `http_requests_total{code="204",method="GET",route="/api/v1/items/:id"} 2`
	if !strings.Contains(string(body), want) {
		t.Errorf("/metrics не содержит %q:\n%s", want, body)
	}
}
