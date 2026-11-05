package l05auth

import (
	"net/http"
	"slices"
	"strings"

	"github.com/gin-gonic/gin"
)

// User это аутентифицированный пользователь запроса.
type User struct {
	ID   string
	Role string
}

// userKey это ключ, под которым Authenticate кладёт User в gin.Context.
const userKey = "auth.user"

// CurrentUser достаёт пользователя, которого положил Authenticate. Готово, менять не нужно.
func CurrentUser(c *gin.Context) (User, bool) {
	v, ok := c.Get(userKey)
	if !ok {
		return User{}, false
	}
	u, ok := v.(User)
	return u, ok
}

// Verifier проверяет токен и возвращает его владельца. В проде это обёртка над
// ParseAccess, в тестах фейк: так middleware тестируется без настоящих JWT.
type Verifier func(token string) (User, error)

// Authenticate это middleware аутентификации.
//
//   - Ждёт заголовок "Authorization: Bearer <token>". Схема регистронезависима
//     ("bearer" тоже годится, так по RFC 9110), токен не пустой.
//   - Нет заголовка, другая схема, пустой токен, verify вернул ошибку: прервать цепочку
//     ответом 401 с телом {"error":"unauthorized"} и заголовком "WWW-Authenticate: Bearer".
//   - Успех: положить User под userKey (c.Set) и передать управление дальше.
func Authenticate(verify Verifier) gin.HandlerFunc {
	return func(c *gin.Context) {
		scheme, token, ok := strings.Cut(c.GetHeader("Authorization"), " ")
		token = strings.TrimSpace(token)
		if !ok || !strings.EqualFold(scheme, "Bearer") || token == "" {
			unauthorized(c)
			return
		}
		u, err := verify(token)
		if err != nil {
			unauthorized(c)
			return
		}
		c.Set(userKey, u)
		c.Next()
	}
}

func unauthorized(c *gin.Context) {
	c.Header("WWW-Authenticate", "Bearer")
	c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
}

// RequireRole это middleware авторизации, ставится после Authenticate.
//
//   - Пользователя в контексте нет (забыли Authenticate или он не отработал): 401 {"error":"unauthorized"}.
//   - Роль пользователя не входит в roles: 403 {"error":"forbidden"}.
//   - Иначе пропустить дальше.
func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		u, ok := CurrentUser(c)
		if !ok {
			unauthorized(c)
			return
		}
		if !slices.Contains(roles, u.Role) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
		c.Next()
	}
}
