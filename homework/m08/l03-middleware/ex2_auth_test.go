package l03middleware

import (
	"net/http"
	"reflect"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRequireToken(t *testing.T) {
	calls := 0
	r := gin.New()
	r.GET("/me", RequireToken(map[string]string{"t-alice": "alice", "t-bob": "bob"}), func(c *gin.Context) {
		calls++
		c.JSON(http.StatusOK, gin.H{"user": c.GetString("user")})
	})

	tests := []struct {
		name   string
		auth   string
		status int
		want   map[string]any
	}{
		{"alice", "Bearer t-alice", 200, map[string]any{"user": "alice"}},
		{"bob, схема строчными", "bearer t-bob", 200, map[string]any{"user": "bob"}},
		{"нет заголовка", "", 401, map[string]any{"error": "unauthorized"}},
		{"неизвестный токен", "Bearer guess", 401, map[string]any{"error": "unauthorized"}},
		{"пустой токен", "Bearer ", 401, map[string]any{"error": "unauthorized"}},
		{"токен без схемы", "t-alice", 401, map[string]any{"error": "unauthorized"}},
		{"другая схема", "Basic t-alice", 401, map[string]any{"error": "unauthorized"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calls = 0
			h := map[string]string{}
			if tt.auth != "" {
				h["Authorization"] = tt.auth
			}
			w := serve(r, http.MethodGet, "/me", h)
			if w.Code != tt.status {
				t.Fatalf("код %d, want %d (тело %q)", w.Code, tt.status, w.Body)
			}
			if got := decodeOne(t, w.Body.Bytes()); !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("тело %v, want %v", got, tt.want)
			}
			wantCalls := 0
			if tt.status == 200 {
				wantCalls = 1
			}
			if calls != wantCalls {
				t.Fatalf("обработчик вызван %d раз, want %d (после отказа нужен Abort)", calls, wantCalls)
			}
			if tt.status == 401 && w.Header().Get("WWW-Authenticate") != "Bearer" {
				t.Errorf("WWW-Authenticate = %q, want \"Bearer\"", w.Header().Get("WWW-Authenticate"))
			}
		})
	}
}
