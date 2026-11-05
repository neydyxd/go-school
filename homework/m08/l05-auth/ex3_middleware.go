package l05auth

import (
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
	panic("TODO: реализуй Authenticate")
}

// RequireRole это middleware авторизации, ставится после Authenticate.
//
//   - Пользователя в контексте нет (забыли Authenticate или он не отработал): 401 {"error":"unauthorized"}.
//   - Роль пользователя не входит в roles: 403 {"error":"forbidden"}.
//   - Иначе пропустить дальше.
func RequireRole(roles ...string) gin.HandlerFunc {
	panic("TODO: реализуй RequireRole")
}
