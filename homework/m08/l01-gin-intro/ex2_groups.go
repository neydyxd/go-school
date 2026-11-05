package l01ginintro

import "github.com/gin-gonic/gin"

// AdminToken это токен, который ждёт группа /admin в заголовке X-Admin-Token.
const AdminToken = "s3cret"

// NewAPI собирает версионированный API на группах маршрутов (gin.New(), без Logger и Recovery).
//
//   - GET /api/v1/users/:id: 200 {"id": "<id как строка>", "version": 1}.
//   - GET /api/v2/users/:id: 200 {"data": {"id": "<id>"}, "version": 2}.
//   - Группа /admin с проверкой прямо в группе (r.Group("/admin", check)):
//     заголовок X-Admin-Token не равен AdminToken, тогда 403 {"error": "forbidden"},
//     и обработчики группы не выполняются. GET /admin/stats: 200 {"users": 2}.
//   - Неизвестный маршрут: 404 {"error": "not found"} (r.NoRoute).
//   - Известный путь с чужим методом (POST /api/v1/users/1): 405
//     {"error": "method not allowed"}. По умолчанию Gin отвечает на это 404,
//     найди поле Engine, которое включает 405, и обработчик r.NoMethod.
func NewAPI() *gin.Engine {
	panic("TODO: реализуй NewAPI")
}
