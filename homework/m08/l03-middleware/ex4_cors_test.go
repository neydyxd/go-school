package l03middleware

import (
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
)

const allowed = "https://app.example.com"

func TestCORS(t *testing.T) {
	calls := 0
	r := gin.New()
	r.Use(CORS(allowed, "http://localhost:5173"))
	r.GET("/items", func(c *gin.Context) { calls++; c.String(http.StatusOK, "items") })

	tests := []struct {
		name       string
		method     string
		header     map[string]string
		status     int
		wantACAO   string
		wantCalled bool
		preflight  bool
	}{
		{"без Origin", "GET", nil, 200, "", true, false},
		{"разрешённый origin", "GET", map[string]string{"Origin": allowed}, 200, allowed, true, false},
		{"второй разрешённый", "GET", map[string]string{"Origin": "http://localhost:5173"}, 200, "http://localhost:5173", true, false},
		{"чужой origin", "GET", map[string]string{"Origin": "https://evil.example"}, 200, "", true, false},
		{"origin-префикс не считается", "GET", map[string]string{"Origin": allowed + ".evil.example"}, 200, "", true, false},
		{"preflight разрешённого", "OPTIONS", map[string]string{"Origin": allowed, "Access-Control-Request-Method": "POST"}, 204, allowed, false, true},
		// OPTIONS без Access-Control-Request-Method это не preflight: идёт дальше, маршрута нет, 404.
		{"OPTIONS без ACRM не preflight", "OPTIONS", map[string]string{"Origin": allowed}, 404, allowed, false, false},
		{"preflight чужого", "OPTIONS", map[string]string{"Origin": "https://evil.example", "Access-Control-Request-Method": "POST"}, 403, "", false, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calls = 0
			w := serve(r, tt.method, "/items", tt.header)
			if w.Code != tt.status {
				t.Fatalf("%s /items: код %d, want %d", tt.method, w.Code, tt.status)
			}
			h := w.Header()
			if got := h.Get("Access-Control-Allow-Origin"); got != tt.wantACAO {
				t.Errorf("Access-Control-Allow-Origin = %q, want %q", got, tt.wantACAO)
			}
			if tt.wantACAO != "" && h.Get("Vary") != "Origin" {
				t.Errorf("Vary = %q, want \"Origin\": иначе кеш отдаст ответ с чужим origin", h.Get("Vary"))
			}
			if (calls == 1) != tt.wantCalled {
				t.Errorf("обработчик вызван %d раз, want вызван=%v", calls, tt.wantCalled)
			}
			if tt.preflight {
				want := map[string]string{
					"Access-Control-Allow-Methods": "GET, POST, PUT, PATCH, DELETE",
					"Access-Control-Allow-Headers": "Authorization, Content-Type, X-Request-ID",
					"Access-Control-Max-Age":       "600",
				}
				for k, v := range want {
					if h.Get(k) != v {
						t.Errorf("%s = %q, want %q", k, h.Get(k), v)
					}
				}
			}
		})
	}
}
