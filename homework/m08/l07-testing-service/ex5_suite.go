package l07testingservice

import (
	"errors"
	"net/http"
	"slices"
	"strings"

	"github.com/gin-gonic/gin"
)

// ---- Готовый код: middleware и роутер. Менять не нужно. ----

// User это пользователь из токена.
type User struct {
	ID   string
	Role string
}

const userKey = "user"

// Verifier проверяет токен.
type Verifier func(token string) (User, error)

// TestVerifier знает ровно два токена. Его используют и роутер в тестах, и твой AuthSuite.
func TestVerifier(token string) (User, error) {
	switch token {
	case "user-token":
		return User{ID: "u1", Role: "user"}, nil
	case "admin-token":
		return User{ID: "a1", Role: "admin"}, nil
	}
	return User{}, errors.New("invalid token")
}

// Authenticate: нет "Authorization: Bearer <token>" или токен не прошёл verify: 401.
func Authenticate(verify Verifier) gin.HandlerFunc {
	return func(c *gin.Context) {
		scheme, token, ok := strings.Cut(c.GetHeader("Authorization"), " ")
		if !ok || !strings.EqualFold(scheme, "Bearer") || token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		u, err := verify(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		c.Set(userKey, u)
		c.Next()
	}
}

// RequireRole: нет пользователя: 401; роль не из списка: 403.
func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		v, ok := c.Get(userKey)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		if !slices.Contains(roles, v.(User).Role) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
		c.Next()
	}
}

// NewRouter собирает роутер с двумя маршрутами:
//
//	GET    /me              auth;        200 {"id":"<id пользователя>"}
//	DELETE /admin/users/:id auth, admin; 204 без тела
func NewRouter(auth, admin gin.HandlerFunc) http.Handler {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/me", auth, func(c *gin.Context) {
		u, _ := c.Get(userKey)
		id := ""
		if u != nil {
			id = u.(User).ID
		}
		c.JSON(http.StatusOK, gin.H{"id": id})
	})
	r.DELETE("/admin/users/:id", auth, admin, func(c *gin.Context) { c.Status(http.StatusNoContent) })
	return r
}

// T это то, что AuthSuite нужно от *testing.T. Узкий интерфейс позволяет
// проверяющему тесту подсунуть свою реализацию и посчитать твои Errorf.
type T interface {
	Helper()
	Errorf(format string, args ...any)
}

// ---- Задание ----

// AuthSuite это твой набор HTTP-тестов аутентификации и авторизации для h,
// собранного через NewRouter. Выполняй запросы через httptest.NewRecorder и
// h.ServeHTTP, при каждом расхождении вызывай t.Errorf (не Fatalf: его в T нет).
// Токены: "user-token" (u1, роль user) и "admin-token" (a1, роль admin).
//
// Набор должен проходить на правильном роутере и ловить сломанные версии
// middleware (их несколько в ex5_suite_test.go, подглядывать можно, но полезнее
// сначала придумать кейсы самому). Думай о таблице: запрос (метод, путь,
// заголовок Authorization) и ожидание (код, для /me ещё id в теле).
func AuthSuite(t T, h http.Handler) {
	panic("TODO: реализуй AuthSuite")
}
