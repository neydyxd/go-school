package http

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"goschool/diploma/internal/service"
)

// ErrorBody это единый формат ошибок из урока про binding (m08/l02):
// {"error":{"code":"validation_failed","message":"...","fields":{"email":"..."}}}.
type ErrorBody struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Fields  map[string]string `json:"fields,omitempty"`
}

func fail(c *gin.Context, status int, code, msg string, fields map[string]string) {
	c.AbortWithStatusJSON(status, gin.H{"error": ErrorBody{Code: code, Message: msg, Fields: fields}})
}

// writeError переводит ошибку сервиса в HTTP-ответ. Подробности (текст ошибки базы)
// клиенту не уходят, только в лог. Новая доменная ошибка: новая ветка здесь.
func writeError(c *gin.Context, err error) {
	var ve *service.ValidationError
	switch {
	case errors.As(err, &ve):
		fail(c, http.StatusUnprocessableEntity, "validation_failed", "проверь поля запроса",
			map[string]string{ve.Field: ve.Message})
	case errors.Is(err, service.ErrNotFound):
		fail(c, http.StatusNotFound, "not_found", "не найдено", nil)
	case errors.Is(err, service.ErrConflict):
		fail(c, http.StatusConflict, "conflict", "конфликт с текущим состоянием", nil)
	default:
		loggerFrom(c).Error("внутренняя ошибка", "err", err)
		fail(c, http.StatusInternalServerError, "internal", "внутренняя ошибка", nil)
	}
}
