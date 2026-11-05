package l01ginintro

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// NewRouter собирает роутер на gin.New() (без Logger и Recovery) с четырьмя маршрутами.
// Все ответы в JSON. На ошибку отвечай одним объектом {"error": "<текст>"} и сразу выходи
// из обработчика: второй c.JSON после первого допишет второй объект в то же тело.
//
//   - GET /ping: 200 {"message": "pong"}.
//   - GET /users/:id: id это целое число больше нуля (strconv.ParseInt),
//     тогда 200 {"id": 42} (число, не строка). Иначе 400.
//   - GET /files/*path: 200 {"path": "docs/a.txt"} для /files/docs/a.txt.
//     Ведущий слеш из параметра убери. Пустой путь (/files/) это 400.
//   - GET /search?q=go&limit=5: q обязателен (пустой или из одних пробелов это 400),
//     пробелы по краям обрежь. limit по умолчанию 10, допустимо 1..50, не число или
//     вне диапазона это 400. Ответ 200 {"q": "go", "limit": 5}.
func NewRouter() *gin.Engine {
	r := gin.New()

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})

	r.GET("/users/:id", func(c *gin.Context) {
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil || id <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "id должен быть положительным целым числом"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"id": id})
	})

	r.GET("/files/*path", func(c *gin.Context) {
		p := strings.TrimPrefix(c.Param("path"), "/")
		if p == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "пустой путь"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"path": p})
	})

	r.GET("/search", func(c *gin.Context) {
		q := strings.TrimSpace(c.Query("q"))
		if q == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "параметр q обязателен"})
			return
		}
		limit, err := strconv.Atoi(c.DefaultQuery("limit", "10"))
		if err != nil || limit < 1 || limit > 50 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "limit должен быть числом от 1 до 50"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"q": q, "limit": limit})
	})

	return r
}
