package l03middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() { gin.SetMode(gin.TestMode) }

func serve(h http.Handler, method, url string, header map[string]string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(method, url, nil)
	for k, v := range header {
		req.Header.Set(k, v)
	}
	h.ServeHTTP(w, req)
	return w
}

// decodeOne требует ровно один JSON-объект в теле (два подряд: обработчик выполнился после отказа).
func decodeOne(t *testing.T, body []byte) map[string]any {
	t.Helper()
	dec := json.NewDecoder(bytes.NewReader(body))
	var got map[string]any
	if err := dec.Decode(&got); err != nil {
		t.Fatalf("тело не JSON-объект: %q (%v)", body, err)
	}
	if _, err := dec.Token(); err != io.EOF {
		t.Fatalf("в теле больше одного JSON-значения: %q (забыт Abort?)", body)
	}
	return got
}

var hex32 = regexp.MustCompile(`^[0-9a-f]{32}$`)

func TestRequestID(t *testing.T) {
	r := gin.New()
	r.Use(RequestID())
	var seen string
	r.GET("/", func(c *gin.Context) {
		seen = RequestIDFrom(c.Request.Context())
		c.Status(http.StatusNoContent)
	})

	tests := []struct {
		name     string
		incoming string
		keep     bool
	}{
		{"нет заголовка", "", false},
		{"допустимый от балансировщика", "lb-7f3a_01", true},
		{"ровно 64 символа", strings.Repeat("a", 64), true},
		{"65 символов", strings.Repeat("a", 65), false},
		{"пробел", "abc def", false},
		{"перевод строки", "abc\nX-Admin: 1", false},
		{"кириллица", "запрос-1", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			seen = ""
			h := map[string]string{}
			if tt.incoming != "" {
				h[HeaderRequestID] = tt.incoming
			}
			w := serve(r, http.MethodGet, "/", h)
			got := w.Header().Get(HeaderRequestID)
			if tt.keep && got != tt.incoming {
				t.Fatalf("X-Request-ID в ответе %q, want входящий %q", got, tt.incoming)
			}
			if !tt.keep && !hex32.MatchString(got) {
				t.Fatalf("X-Request-ID в ответе %q, want новый ID из 32 hex-символов", got)
			}
			if seen != got {
				t.Fatalf("RequestIDFrom(c.Request.Context()) = %q, want %q (как в ответе)", seen, got)
			}
		})
	}
}

func TestRequestIDUnique(t *testing.T) {
	r := gin.New()
	r.Use(RequestID())
	r.GET("/", func(c *gin.Context) {})
	ids := map[string]bool{}
	for range 200 {
		id := serve(r, http.MethodGet, "/", nil).Header().Get(HeaderRequestID)
		if ids[id] {
			t.Fatalf("ID %q повторился: генерируй случайный (crypto/rand)", id)
		}
		ids[id] = true
	}
}

func TestRequestIDFromEmpty(t *testing.T) {
	if got := RequestIDFrom(context.Background()); got != "" {
		t.Errorf("RequestIDFrom(Background) = %q, want \"\"", got)
	}
	ctx := context.WithValue(context.Background(), "request_id", "чужой")
	if got := RequestIDFrom(ctx); got != "" {
		t.Errorf("RequestIDFrom(ctx со строковым ключом) = %q, want \"\": ключ должен быть приватного типа", got)
	}
}
