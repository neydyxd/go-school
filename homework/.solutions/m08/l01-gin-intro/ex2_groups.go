package l01ginintro

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

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
	r := gin.New()
	r.HandleMethodNotAllowed = true
	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
	})
	r.NoMethod(func(c *gin.Context) {
		c.JSON(http.StatusMethodNotAllowed, gin.H{"error": "method not allowed"})
	})

	api := r.Group("/api")
	{
		v1 := api.Group("/v1")
		v1.GET("/users/:id", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"id": c.Param("id"), "version": 1})
		})

		v2 := api.Group("/v2")
		v2.GET("/users/:id", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"data": gin.H{"id": c.Param("id")}, "version": 2})
		})
	}

	admin := r.Group("/admin", func(c *gin.Context) {
		if c.GetHeader("X-Admin-Token") != AdminToken {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
		c.Next()
	})
	admin.GET("/stats", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"users": 2})
	})

	return r
}
