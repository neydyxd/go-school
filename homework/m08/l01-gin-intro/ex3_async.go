package l01ginintro

import "github.com/gin-gonic/gin"

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
	panic("TODO: реализуй ReportHandler")
}
