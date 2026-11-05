package l03middleware

import (
	"net/http"
	"slices"

	"github.com/gin-gonic/gin"
)

// CORS возвращает middleware для браузерных клиентов с разрешёнными origins.
//
//   - Нет заголовка Origin: ничего не трогать, дальше по цепочке.
//   - Origin разрешён: Access-Control-Allow-Origin: <этот origin> (не "*") и Vary: Origin.
//   - Preflight это OPTIONS с заголовком Access-Control-Request-Method. Для разрешённого
//     origin ответить 204 с заголовками
//     Access-Control-Allow-Methods: GET, POST, PUT, PATCH, DELETE
//     Access-Control-Allow-Headers: Authorization, Content-Type, X-Request-ID
//     Access-Control-Max-Age: 600
//     и прервать цепочку (до проверки токена дело не доходит). Для чужого origin: 403, тоже прервать.
//   - Обычный запрос с чужим origin: CORS-заголовков не ставить, запрос пропустить
//     (браузер сам не отдаст ответ странице).
func CORS(origins ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin == "" {
			c.Next()
			return
		}
		ok := slices.Contains(origins, origin)
		if ok {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Vary", "Origin")
		}
		if c.Request.Method == http.MethodOptions && c.GetHeader("Access-Control-Request-Method") != "" {
			if !ok {
				c.AbortWithStatus(http.StatusForbidden)
				return
			}
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE")
			c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type, X-Request-ID")
			c.Header("Access-Control-Max-Age", "600")
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
