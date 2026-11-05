package l02binding

import (
	"cmp"
	"net/http"
	"slices"
	"strings"

	"github.com/gin-gonic/gin"
)

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
	Page     int    `json:"page" form:"page,default=1" binding:"min=1"`
	PerPage  int    `json:"per_page" form:"per_page,default=20" binding:"min=1,max=100"`
	Category string `json:"category" form:"category" binding:"omitempty,oneof=books games music"`
	Sort     string `json:"sort" form:"sort" binding:"omitempty,oneof=price -price name -name"`
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
	return func(c *gin.Context) {
		var q ListQuery
		if err := c.ShouldBindQuery(&q); err != nil {
			bindError(c, err)
			return
		}

		items := slices.DeleteFunc(store.All(), func(p Product) bool {
			return q.Category != "" && p.Category != q.Category
		})
		if q.Sort != "" {
			field, desc := strings.CutPrefix(q.Sort, "-")
			slices.SortStableFunc(items, func(a, b Product) int {
				r := cmp.Compare(a.Price, b.Price)
				if field == "name" {
					r = strings.Compare(a.Name, b.Name)
				}
				if desc {
					return -r
				}
				return r
			})
		}

		total := len(items)
		from := min((q.Page-1)*q.PerPage, total)
		to := min(from+q.PerPage, total)
		page := append([]Product{}, items[from:to]...) // [] вместо null для пустой страницы

		c.JSON(http.StatusOK, gin.H{
			"data": page,
			"meta": Meta{Page: q.Page, PerPage: q.PerPage, Total: total, TotalPages: (total + q.PerPage - 1) / q.PerPage},
		})
	}
}

// ProductURI это параметры пути GET /products/:id. Допиши теги: uri "id", не меньше 1.
type ProductURI struct {
	ID int64 `json:"id" uri:"id" binding:"min=1"`
}

// GetProduct обрабатывает GET /products/:id через ShouldBindUri.
//
//   - id не число: 400 bad_request; id меньше 1: 422 validation_failed;
//   - нет товара: 404 not_found; есть: 200 {"data": <Product>}.
func GetProduct(store *Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		var u ProductURI
		if err := c.ShouldBindUri(&u); err != nil {
			bindError(c, err)
			return
		}
		p, ok := store.Get(u.ID)
		if !ok {
			fail(c, http.StatusNotFound, CodeNotFound, "товар не найден", nil)
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": p})
	}
}
