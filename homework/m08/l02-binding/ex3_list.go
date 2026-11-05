package l02binding

import "github.com/gin-gonic/gin"

// ListQuery это query-параметры GET /products. Допиши теги form и binding:
//
//	page      по умолчанию 1, не меньше 1
//	per_page  по умолчанию 20, от 1 до 100
//	category  необязателен, одно из books, games, music
//	sort      необязателен, одно из price, -price, name, -name (минус: по убыванию)
//
// json-теги уже стоят: по ним SetupValidator называет поля в ошибках.
// Подсказка: значение по умолчанию задаётся прямо в теге: form:"page,default=1".
type ListQuery struct {
	Page     int    `json:"page"`
	PerPage  int    `json:"per_page"`
	Category string `json:"category"`
	Sort     string `json:"sort"`
}

// ListProducts обрабатывает GET /products: фильтр по category, сортировка, страница.
//
//   - Без sort порядок как в Store.All() (по ID). С sort сортировка устойчивая
//     (slices.SortStableFunc), равные элементы остаются в порядке ID.
//   - Ответ 200 {"data": [...], "meta": {"page", "per_page", "total", "total_pages"}},
//     total считается после фильтра, total_pages = ceil(total / per_page) (0 для пустого).
//   - Страница за пределами списка: 200 и "data": [] (не null).
//   - page=abc (не разбирается в число): 400 bad_request; page=0, per_page=500,
//     sort=price%3B DROP: 422 validation_failed с fields.
func ListProducts(store *Store) gin.HandlerFunc {
	panic("TODO: реализуй ListProducts")
}

// ProductURI это параметры пути GET /products/:id. Допиши теги: uri "id", не меньше 1.
type ProductURI struct {
	ID int64 `json:"id"`
}

// GetProduct обрабатывает GET /products/:id через ShouldBindUri.
//
//   - id не число: 400 bad_request; id меньше 1: 422 validation_failed;
//   - нет товара: 404 not_found; есть: 200 {"data": <Product>}.
func GetProduct(store *Store) gin.HandlerFunc {
	panic("TODO: реализуй GetProduct")
}
