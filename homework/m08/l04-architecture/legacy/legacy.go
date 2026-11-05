// Package legacy это «спагетти»-версия API библиотеки: всё в одном обработчике.
// Так пишут первый прототип, и так он доживает до продакшна. Прочитай и найди проблемы,
// прежде чем раскладывать код по слоям в пакетах service, postgres и httpapi.
// Этот пакет не компилируется в итоговый сервис и тестов не имеет.
package legacy

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Register вешает маршруты прямо на пул соединений.
func Register(r gin.IRouter, pool *pgxpool.Pool) {
	r.POST("/books", func(c *gin.Context) {
		var req struct {
			Title  string `json:"title"`
			Author string `json:"author"`
			ISBN   string `json:"isbn"`
			Copies int    `json:"copies"`
		}
		if err := c.BindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
		if req.Title == "" {
			c.JSON(400, gin.H{"error": "title required"})
		}
		var id int64
		err := pool.QueryRow(c, `INSERT INTO books (title, author, isbn, copies)
			VALUES ($1, $2, $3, $4) RETURNING id`, req.Title, req.Author, req.ISBN, req.Copies).Scan(&id)
		if err != nil {
			if strings.Contains(err.Error(), "duplicate key") {
				c.JSON(400, gin.H{"error": "isbn exists"})
				return
			}
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, gin.H{"id": id, "title": req.Title, "author": req.Author, "isbn": req.ISBN, "copies": req.Copies})
	})

	r.POST("/books/:id/take", func(c *gin.Context) {
		id, _ := strconv.Atoi(c.Param("id"))
		var copies int
		var title string
		err := pool.QueryRow(c, `SELECT title, copies FROM books WHERE id = $1`, id).Scan(&title, &copies)
		if err != nil {
			c.JSON(404, gin.H{"error": "not found"})
			return
		}
		if copies == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "no copies"})
			return
		}
		_, err = pool.Exec(c, `UPDATE books SET copies = $1 WHERE id = $2`, copies-1, id)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, gin.H{"id": id, "title": title, "copies": copies - 1})
	})
}
