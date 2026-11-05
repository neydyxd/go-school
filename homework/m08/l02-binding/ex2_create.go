package l02binding

import "github.com/gin-gonic/gin"

// CreateProductRequest это тело POST /products. Допиши теги binding:
//
//	name      обязателен, длина от 2 до 100 символов
//	slug      обязателен, тег slug из SetupValidator
//	price     обязателен, больше 0 (копейки)
//	category  обязателен, одно из books, games, music
//	tags      необязателен, не больше 5 штук, каждый от 2 до 20 символов (dive)
type CreateProductRequest struct {
	Name     string   `json:"name"`
	Slug     string   `json:"slug"`
	Price    int64    `json:"price"`
	Category string   `json:"category"`
	Tags     []string `json:"tags"`
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
	panic("TODO: реализуй CreateProduct")
}
