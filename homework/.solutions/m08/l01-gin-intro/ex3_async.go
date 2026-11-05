package l01ginintro

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ReportHandler обрабатывает POST /reports/:id. Перед ним в цепочке стоит middleware,
// которое кладёт имя пользователя: c.Set("user", "alice").
//
// Обработчик сразу отвечает 202 {"status": "accepted"}, а отчёт готовит в фоне:
// в новой горутине вызывает work() (долгая работа), затем send(user, reportID),
// где user = c.GetString("user"), reportID = c.Param("id").
//
// После возврата из обработчика Gin кладёт *gin.Context обратно в пул и отдаёт
// следующему запросу. Горутина, которая читает исходный c после work(), увидит
// параметры и ключи чужого запроса. Ответ из горутины тоже не пиши.
func ReportHandler(work func(), send func(user, reportID string)) gin.HandlerFunc {
	return func(c *gin.Context) {
		cp := c.Copy() // копию делаем до go, пока c ещё принадлежит нам
		go func() {
			work()
			send(cp.GetString("user"), cp.Param("id"))
		}()
		c.JSON(http.StatusAccepted, gin.H{"status": "accepted"})
	}
}
