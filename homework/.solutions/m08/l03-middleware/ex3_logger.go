package l03middleware

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

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
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		status := c.Writer.Status()
		level := slog.LevelInfo
		switch {
		case status >= 500:
			level = slog.LevelError
		case status >= 400:
			level = slog.LevelWarn
		}
		ctx := c.Request.Context()
		log.LogAttrs(ctx, level, "request",
			slog.String("method", c.Request.Method),
			slog.String("route", c.FullPath()),
			slog.String("path", c.Request.URL.Path),
			slog.Int("status", status),
			slog.Duration("duration", time.Since(start)),
			slog.String("request_id", RequestIDFrom(ctx)),
		)
	}
}

// Формат тела ответа здесь упрощён ради краткости ({"error": "..."}). В настоящем сервисе
// ответы идут в едином формате курса из урока про binding: {"error": {"code", "message", "fields"}}.
//
// Recovery возвращает middleware, которое перехватывает панику дальше по цепочке:
// пишет запись уровня Error с сообщением "panic" и атрибутом panic (fmt.Sprint значения)
// и отвечает 500 {"error": "internal"}. Подробности паники клиенту не отдаём.
func Recovery(log *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if v := recover(); v != nil {
				log.ErrorContext(c.Request.Context(), "panic", slog.String("panic", fmt.Sprint(v)))
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "internal"})
			}
		}()
		c.Next()
	}
}
