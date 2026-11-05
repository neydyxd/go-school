package l06configlogsdocs

import (
	"bytes"
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() { gin.SetMode(gin.TestMode) }

func TestFromContext(t *testing.T) {
	if FromContext(context.Background()) == nil {
		t.Fatal("FromContext(пустой контекст) = nil, want slog.Default()")
	}
	l := slog.New(slog.DiscardHandler)
	if FromContext(WithLogger(context.Background(), l)) != l {
		t.Error("FromContext не вернул логгер, положенный через WithLogger")
	}
}

func TestRequestLogger(t *testing.T) {
	tests := []struct {
		name      string
		target    string
		reqID     string
		status    int
		wantLevel string
		wantID    string // пусто: должен быть сгенерирован новый
	}{
		{"успех с чужим id", "/v1/tasks?token=leaked-in-query", "abc-123", 201, "INFO", "abc-123"},
		{"без id", "/v1/tasks", "", 200, "INFO", ""},
		{"инъекция в id", "/v1/tasks", "x\n{\"level\":\"ERROR\"}", 200, "INFO", ""},
		{"слишком длинный id", "/v1/tasks", strings.Repeat("a", 65), 200, "INFO", ""},
		{"ошибка клиента", "/v1/tasks/7", "id-404", 404, "WARN", "id-404"},
		{"ошибка сервера", "/v1/tasks/7", "id-500", 500, "ERROR", "id-500"},
	}
	hex32 := regexp.MustCompile(`^[0-9a-f]{32}$`)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			base := slog.New(slog.NewJSONHandler(&buf, nil))
			r := gin.New()
			r.Use(RequestLogger(base))
			h := func(c *gin.Context) {
				FromContext(c.Request.Context()).Info("внутри обработчика")
				c.Status(tt.status)
			}
			r.GET("/v1/tasks", h)
			r.GET("/v1/tasks/:id", h)

			req := httptest.NewRequest(http.MethodGet, tt.target, nil)
			req.Header.Set("Authorization", "Bearer super-secret-jwt")
			if tt.reqID != "" {
				req.Header.Set("X-Request-ID", tt.reqID)
			}
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			out := buf.String()
			for _, leak := range []string{"super-secret-jwt", "leaked-in-query"} {
				if strings.Contains(out, leak) {
					t.Errorf("в лог утёк %q:\n%s", leak, out)
				}
			}
			ls := lines(t, &buf)
			if len(ls) != 2 {
				t.Fatalf("строк лога %d, want 2 (обработчик + итог запроса):\n%s", len(ls), out)
			}
			inner, final := ls[0], ls[1]

			id := w.Header().Get("X-Request-ID")
			if tt.wantID != "" && id != tt.wantID {
				t.Errorf("заголовок X-Request-ID = %q, want %q", id, tt.wantID)
			}
			if tt.wantID == "" && !hex32.MatchString(id) {
				t.Errorf("сгенерированный X-Request-ID = %q, want 32 hex-символа", id)
			}
			if inner["request_id"] != id || final["request_id"] != id {
				t.Errorf("request_id в логах: обработчик %v, итог %v; want %q в обоих", inner["request_id"], final["request_id"], id)
			}
			if inner["msg"] != "внутри обработчика" {
				t.Errorf("первая строка msg = %v, want \"внутри обработчика\"", inner["msg"])
			}
			path := strings.SplitN(tt.target, "?", 2)[0]
			want := map[string]any{
				"msg": "http request", "level": tt.wantLevel, "method": "GET",
				"path": path, "status": float64(tt.status),
			}
			for k, v := range want {
				if final[k] != v {
					t.Errorf("итоговая строка %s = %v, want %v\n%s", k, final[k], v, out)
				}
			}
			if _, ok := final["duration_ms"].(float64); !ok {
				t.Errorf("итоговая строка без числового duration_ms: %v", final)
			}
		})
	}
}
