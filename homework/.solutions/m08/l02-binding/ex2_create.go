package l02binding

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

// CreateProductRequest это тело POST /products. Допиши теги binding:
//
//	name      обязателен, длина от 2 до 100 символов
//	slug      обязателен, тег slug из SetupValidator
//	price     обязателен, больше 0 (копейки)
//	category  обязателен, одно из books, games, music
//	tags      необязателен, не больше 5 штук, каждый от 2 до 20 символов (dive)
type CreateProductRequest struct {
	Name     string   `json:"name" binding:"required,min=2,max=100"`
	Slug     string   `json:"slug" binding:"required,slug"`
	Price    int64    `json:"price" binding:"required,gt=0"`
	Category string   `json:"category" binding:"required,oneof=books games music"`
	Tags     []string `json:"tags" binding:"max=5,dive,min=2,max=20"`
}

// CreateProduct обрабатывает POST /products. Валидатор Gin уже настроен через SetupValidator.
//
// Ответы (формат из store.go):
//
//   - тело не JSON, пустое или поле не того типа: 400 {"error": {"code": "bad_request", "message": ...}};
//   - не прошли проверки: 422 {"error": {"code": "validation_failed", "message": ..., "fields": FieldErrors(err)}};
//   - slug занят (ErrSlugTaken): 409 {"error": {"code": "conflict", "message": ...}};
//   - успех: 201 {"data": <сохранённый Product>}; tags не передан, значит в ответе пустой массив [], не null.
//
// Используй ShouldBindJSON: Bind/BindJSON сами пишут 400 и не дадут ответить 422.
func CreateProduct(store *Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req CreateProductRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			bindError(c, err)
			return
		}
		if req.Tags == nil {
			req.Tags = []string{}
		}
		p, err := store.Add(Product{Name: req.Name, Slug: req.Slug, Price: req.Price, Category: req.Category, Tags: req.Tags})
		if errors.Is(err, ErrSlugTaken) {
			fail(c, http.StatusConflict, CodeConflict, "товар с таким slug уже есть", nil)
			return
		}
		if err != nil {
			fail(c, http.StatusInternalServerError, "internal", "внутренняя ошибка", nil)
			return
		}
		c.JSON(http.StatusCreated, gin.H{"data": p})
	}
}

// bindError разделяет «не разобрали» (400) и «разобрали, но не прошли проверки» (422).
func bindError(c *gin.Context, err error) {
	if fields := FieldErrors(err); fields != nil {
		fail(c, http.StatusUnprocessableEntity, CodeValidation, "проверь поля запроса", fields)
		return
	}
	fail(c, http.StatusBadRequest, CodeBadRequest, "не удалось разобрать запрос", nil)
}

func fail(c *gin.Context, status int, code, msg string, fields map[string]string) {
	c.AbortWithStatusJSON(status, gin.H{"error": ErrorBody{Code: code, Message: msg, Fields: fields}})
}
