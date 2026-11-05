package l03middleware

import (
	"log/slog"

	"github.com/gin-gonic/gin"
)

// Logger возвращает middleware, которое пишет одну запись "request" на каждый запрос,
// после того как отработала вся цепочка. Атрибуты:
//
//	method      метод запроса
//	route       шаблон маршрута (c.FullPath(), например "/users/:id"; для 404 пусто)
//	path        фактический путь (c.Request.URL.Path)
//	status      итоговый код ответа (c.Writer.Status())
//	duration    длительность, time.Duration (slog.Duration)
//	request_id  RequestIDFrom(контекст запроса)
//
// Уровень по статусу: 5xx Error, 4xx Warn, остальное Info.
func Logger(log *slog.Logger) gin.HandlerFunc {
	panic("TODO: реализуй Logger")
}

// Формат тела ответа здесь упрощён ради краткости ({"error": "..."}). В настоящем сервисе
// ответы идут в едином формате курса из урока про binding: {"error": {"code", "message", "fields"}}.
//
// Recovery возвращает middleware, которое перехватывает панику дальше по цепочке:
// пишет запись уровня Error с сообщением "panic" и атрибутом panic (fmt.Sprint значения)
// и отвечает 500 {"error": "internal"}. Подробности паники клиенту не отдаём.
func Recovery(log *slog.Logger) gin.HandlerFunc {
	panic("TODO: реализуй Recovery")
}
