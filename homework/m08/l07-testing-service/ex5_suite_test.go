package l07testingservice

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

// recorder это поддельный T: считает Errorf вместо того, чтобы валить тест.
type recorder struct{ errors []string }

func (r *recorder) Helper() {}
func (r *recorder) Errorf(format string, args ...any) {
	r.errors = append(r.errors, fmt.Sprintf(format, args...))
}

func abort(c *gin.Context, code int) {
	c.AbortWithStatusJSON(code, gin.H{"error": http.StatusText(code)})
}

// Мутанты: правдоподобно сломанные версии middleware. Каждую должен поймать AuthSuite.
var mutants = []struct {
	name        string
	auth, admin gin.HandlerFunc
}{
	{"пропускает запрос без заголовка", func(c *gin.Context) {
		if c.GetHeader("Authorization") == "" {
			c.Next()
			return
		}
		Authenticate(TestVerifier)(c)
	}, RequireRole("admin")},
	{"не проверяет токен, только префикс", func(c *gin.Context) {
		if strings.HasPrefix(c.GetHeader("Authorization"), "Bearer ") {
			c.Set(userKey, User{ID: "u1", Role: "user"})
			c.Next()
			return
		}
		abort(c, 401)
	}, RequireRole("admin")},
	{"403 вместо 401 на плохой токен", func(c *gin.Context) {
		h := c.GetHeader("Authorization")
		token, ok := strings.CutPrefix(h, "Bearer ")
		if !ok {
			abort(c, 401)
			return
		}
		u, err := TestVerifier(token)
		if err != nil {
			abort(c, 403)
			return
		}
		c.Set(userKey, u)
		c.Next()
	}, RequireRole("admin")},
	{"401 вместо 403 на чужую роль", Authenticate(TestVerifier), func(c *gin.Context) {
		v, _ := c.Get(userKey)
		if u, ok := v.(User); !ok || u.Role != "admin" {
			abort(c, 401)
			return
		}
		c.Next()
	}},
	{"RequireRole ничего не проверяет", Authenticate(TestVerifier), func(c *gin.Context) { c.Next() }},
	{"RequireRole путает роли", Authenticate(TestVerifier), RequireRole("user")},
	{"Authenticate кладёт не того пользователя", func(c *gin.Context) {
		Authenticate(func(string) (User, error) { return User{ID: "a1", Role: "user"}, nil })(c)
	}, RequireRole("admin")},
	{"не принимает admin-токен", Authenticate(func(tok string) (User, error) {
		if tok == "admin-token" {
			return User{}, fmt.Errorf("revoked")
		}
		return TestVerifier(tok)
	}), RequireRole("admin")},
}

func TestAuthSuite(t *testing.T) {
	// 1. Набор обязан проходить на правильной реализации, иначе он врёт.
	AuthSuite(t, NewRouter(Authenticate(TestVerifier), RequireRole("admin")))
	if t.Failed() {
		t.Fatal("AuthSuite падает на правильном роутере: ошибка в самих тестах")
	}
	// 2. И ловить каждого мутанта.
	for _, m := range mutants {
		t.Run(m.name, func(t *testing.T) {
			rec := &recorder{}
			AuthSuite(rec, NewRouter(m.auth, m.admin))
			if len(rec.errors) == 0 {
				t.Errorf("мутант %q выжил: AuthSuite не заметил поломку", m.name)
			}
		})
	}
}
