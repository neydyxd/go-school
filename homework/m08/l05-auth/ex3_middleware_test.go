package l05auth

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() { gin.SetMode(gin.TestMode) }

// fakeVerify знает два токена. Всё остальное, включая пустую строку, отвергает.
func fakeVerify(token string) (User, error) {
	switch token {
	case "good-user":
		return User{ID: "u1", Role: "user"}, nil
	case "good-admin":
		return User{ID: "a1", Role: "admin"}, nil
	}
	return User{}, errors.New("bad token")
}

// authRouter: /me доступен любому вошедшему, /admin только admin,
// /broken повесил RequireRole без Authenticate (ошибка конфигурации роутов).
func authRouter(reached *bool) *gin.Engine {
	r := gin.New()
	me := func(c *gin.Context) {
		*reached = true
		u, _ := CurrentUser(c)
		c.JSON(http.StatusOK, gin.H{"id": u.ID})
	}
	r.GET("/me", Authenticate(fakeVerify), me)
	r.GET("/admin", Authenticate(fakeVerify), RequireRole("admin"), me)
	r.GET("/staff", Authenticate(fakeVerify), RequireRole("admin", "user"), me)
	r.GET("/broken", RequireRole("admin"), me)
	return r
}

func TestAuthMiddleware(t *testing.T) {
	tests := []struct {
		name       string
		path, auth string
		wantCode   int
		wantBody   string // значение поля error или id
	}{
		{"нет заголовка", "/me", "", 401, "unauthorized"},
		{"Bearer без токена", "/me", "Bearer ", 401, "unauthorized"},
		{"только слово Bearer", "/me", "Bearer", 401, "unauthorized"},
		{"схема Basic", "/me", "Basic Z29vZC11c2Vy", 401, "unauthorized"},
		{"токен без схемы", "/me", "good-user", 401, "unauthorized"},
		{"невалидный токен", "/me", "Bearer forged", 401, "unauthorized"},
		{"валидный токен", "/me", "Bearer good-user", 200, "u1"},
		{"схема в нижнем регистре", "/me", "bearer good-user", 200, "u1"},
		{"user в админке: 403, а не 401", "/admin", "Bearer good-user", 403, "forbidden"},
		{"аноним в админке: 401, а не 403", "/admin", "", 401, "unauthorized"},
		{"admin в админке", "/admin", "Bearer good-admin", 200, "a1"},
		{"несколько разрешённых ролей", "/staff", "Bearer good-user", 200, "u1"},
		{"RequireRole без Authenticate", "/broken", "Bearer good-admin", 401, "unauthorized"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reached := false
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			if tt.auth != "" {
				req.Header.Set("Authorization", tt.auth)
			}
			w := httptest.NewRecorder()
			authRouter(&reached).ServeHTTP(w, req)

			if w.Code != tt.wantCode {
				t.Fatalf("GET %s (%q): код %d, want %d; тело %s", tt.path, tt.auth, w.Code, tt.wantCode, w.Body)
			}
			var body map[string]string
			if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
				t.Fatalf("тело не JSON-объект со строками: %q", w.Body)
			}
			if tt.wantCode == 200 {
				if body["id"] != tt.wantBody {
					t.Errorf("id = %q, want %q", body["id"], tt.wantBody)
				}
				return
			}
			if body["error"] != tt.wantBody {
				t.Errorf("error = %q, want %q", body["error"], tt.wantBody)
			}
			if reached {
				t.Errorf("при коде %d обработчик всё равно вызвался: цепочку не прервали", w.Code)
			}
			if w.Code == 401 && w.Header().Get("WWW-Authenticate") != "Bearer" {
				t.Errorf("на 401 заголовок WWW-Authenticate = %q, want \"Bearer\"", w.Header().Get("WWW-Authenticate"))
			}
		})
	}
}
