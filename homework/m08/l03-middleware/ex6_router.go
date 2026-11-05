package l03middleware

import (
	"log/slog"

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
	panic("TODO: реализуй NewRouter")
}
