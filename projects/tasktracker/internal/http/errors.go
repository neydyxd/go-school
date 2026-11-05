package http

import (
	"errors"
	"net/http"
	"reflect"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"

	"goschool/tasktracker/internal/service"
)

var errPanic = errors.New("panic")

// ErrorBody это единый формат ошибок из урока про binding (m08/l02):
// {"error":{"code":"validation_failed","message":"...","fields":{"email":"..."}}}.
// fields есть только у ошибок валидации.
type ErrorBody struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Fields  map[string]string `json:"fields,omitempty"`
}

func fail(c *gin.Context, status int, code, msg string, fields map[string]string) {
	c.AbortWithStatusJSON(status, gin.H{"error": ErrorBody{Code: code, Message: msg, Fields: fields}})
}

// bindError: ошибка ShouldBind*. Валидатор отработал: 422 validation_failed с fields,
// тело или параметр не разобрать: 400 bad_request.
func bindError(c *gin.Context, err error) {
	if fields := fieldErrors(err); fields != nil {
		fail(c, http.StatusUnprocessableEntity, "validation_failed", "проверь поля запроса", fields)
		return
	}
	fail(c, http.StatusBadRequest, "bad_request", "не удалось разобрать запрос", nil)
}

// writeError переводит ошибку сервиса в HTTP-ответ. Внутренние подробности
// (текст ошибки базы) клиенту не уходят, только в лог.
func writeError(c *gin.Context, err error) {
	var ve *service.ValidationError
	switch {
	case errors.As(err, &ve):
		fail(c, http.StatusUnprocessableEntity, "validation_failed", "проверь поля запроса",
			map[string]string{ve.Field: ve.Message})
	case errors.Is(err, service.ErrUnauthorized):
		fail(c, http.StatusUnauthorized, "unauthorized", "неверные учётные данные", nil)
	case errors.Is(err, service.ErrNotFound):
		fail(c, http.StatusNotFound, "not_found", "не найдено", nil)
	case errors.Is(err, service.ErrConflict):
		fail(c, http.StatusConflict, "conflict", "конфликт с текущим состоянием", nil)
	case errors.Is(err, service.ErrNotImplemented):
		fail(c, http.StatusNotImplemented, "not_implemented", "ещё не реализовано", nil)
	default:
		loggerFrom(c).Error("внутренняя ошибка", "err", err)
		fail(c, http.StatusInternalServerError, "internal", "внутренняя ошибка", nil)
	}
}

var setupValidator sync.Once

// useJSONNames: в fields поле называется как в JSON ("email"), а не как в Go ("Email").
func useJSONNames() {
	setupValidator.Do(func() {
		if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
			v.RegisterTagNameFunc(func(f reflect.StructField) string {
				name, _, _ := strings.Cut(f.Tag.Get("json"), ",")
				if name == "-" {
					return ""
				}
				return name
			})
		}
	})
}

func fieldErrors(err error) map[string]string {
	var ve validator.ValidationErrors
	if !errors.As(err, &ve) {
		return nil
	}
	out := make(map[string]string, len(ve))
	for _, fe := range ve {
		_, key, _ := strings.Cut(fe.Namespace(), ".") // без имени структуры
		switch fe.Tag() {
		case "required":
			out[key] = "обязательное поле"
		case "oneof":
			out[key] = "одно из: " + strings.Join(strings.Fields(fe.Param()), ", ")
		default:
			out[key] = "не прошло проверку " + fe.Tag() + " " + fe.Param()
		}
	}
	return out
}
