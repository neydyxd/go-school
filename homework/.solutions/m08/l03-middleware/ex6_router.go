package l03middleware

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

// NewRouter собирает сервис из middleware этого урока (gin.New(), без встроенных Logger и Recovery).
//
// Глобальная цепочка, подбери порядок сам: RequestID, Logger, Recovery, CORS(origins...).
// Требования, по которым тесты проверят порядок:
//
//   - запрос, который упал с паникой, всё равно попадает в лог как "request" со статусом 500
//     и с тем же request_id, что ушёл клиенту в X-Request-ID;
//   - preflight на /api/... проходит без токена.
//
// Маршруты:
//
//	GET /health      без авторизации: 200 {"status": "ok"}
//	группа /api      с RequireToken(tokens)
//	GET /api/me      200 {"user": "<имя из c.Set>"}
//	GET /api/panic   panic("boom")
func NewRouter(log *slog.Logger, tokens map[string]string, origins []string) *gin.Engine {
	r := gin.New()
	// Logger снаружи Recovery: иначе паника пролетит мимо записи лога, а 500 он не увидит.
	// CORS глобально: preflight на несуществующий OPTIONS-маршрут проходит только через r.Use.
	r.Use(RequestID(), Logger(log), Recovery(log), CORS(origins...))

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	api := r.Group("/api", RequireToken(tokens))
	api.GET("/me", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"user": c.GetString("user")})
	})
	api.GET("/panic", func(c *gin.Context) {
		panic("boom")
	})
	return r
}
