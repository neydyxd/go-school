package l03middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// Формат тела ответа здесь упрощён ради краткости ({"error": "..."}). В настоящем сервисе
// ответы идут в едином формате курса из урока про binding: {"error": {"code", "message", "fields"}}.
//
// RequireToken возвращает middleware проверки токена. tokens: токен -> имя пользователя.
//
//   - Заголовок Authorization: "Bearer <токен>" (схема без учёта регистра, "bearer" тоже годится).
//     Токен есть в tokens: c.Set("user", имя) и дальше по цепочке.
//   - Нет заголовка, другая схема (Basic), пустой или неизвестный токен:
//     401 {"error": "unauthorized"}, заголовок WWW-Authenticate: Bearer,
//     и следующие обработчики не выполняются.
func RequireToken(tokens map[string]string) gin.HandlerFunc {
	return func(c *gin.Context) {
		scheme, token, ok := strings.Cut(c.GetHeader("Authorization"), " ")
		user, known := tokens[token]
		if !ok || !strings.EqualFold(scheme, "Bearer") || token == "" || !known {
			c.Header("WWW-Authenticate", "Bearer")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		c.Set("user", user)
		c.Next()
	}
}
