// Package httpapi это транспортный слой: разбор HTTP, вызов сервиса, ответ в JSON.
package httpapi

import (
	"context"

	"github.com/gin-gonic/gin"

	"goschool/homework/m08/l04-architecture/library"
)

// BookService это то, что HTTP-слою нужно от бизнес-логики. *service.Service ему удовлетворяет,
// а тесты подставляют fake и проверяют обработчики без базы.
type BookService interface {
	Create(ctx context.Context, title, author, isbn string, copies int) (library.Book, error)
	Get(ctx context.Context, id int64) (library.Book, error)
	Take(ctx context.Context, id int64) (library.Book, error)
}

// Handler держит зависимости HTTP-слоя.
type Handler struct {
	svc BookService
}

// New собирает обработчики.
func New(svc BookService) *Handler {
	return &Handler{svc: svc}
}

// createBookRequest это тело POST /books. Бизнес-правила проверяет сервис, здесь только форма JSON.
type createBookRequest struct {
	Title  string `json:"title"`
	Author string `json:"author"`
	ISBN   string `json:"isbn"`
	Copies int    `json:"copies"`
}

// bookResponse это книга в ответе API. Отдельный тип, чтобы JSON не зависел от доменной структуры.
type bookResponse struct {
	ID     int64  `json:"id"`
	Title  string `json:"title"`
	Author string `json:"author"`
	ISBN   string `json:"isbn"`
	Copies int    `json:"copies"`
}

func toResponse(b library.Book) bookResponse {
	return bookResponse{ID: b.ID, Title: b.Title, Author: b.Author, ISBN: b.ISBN, Copies: b.Copies}
}

// errorBody это {"error": {...}} в ответе, единый формат курса из урока про binding:
// {"error": {"code", "message", "fields"}}. Fields заполняется только для ошибок валидации.
type errorBody struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Fields  map[string]string `json:"fields,omitempty"`
}

// Register вешает маршруты на r (обычно это группа /api):
//
//	POST /books            201 {"data": bookResponse}
//	GET  /books/:id        200 {"data": bookResponse}
//	POST /books/:id/take   200 {"data": bookResponse}
//
// В сервис передавай c.Request.Context(), а не сам c: *gin.Context тоже context.Context,
// но по умолчанию (ContextWithFallback = false) не отдаёт отмену и значения контекста запроса.
// Тело не JSON или id не число: 400 {"error": {"code": "bad_request", ...}}. Остальные ошибки
// через writeError.
func (h *Handler) Register(r gin.IRouter) {
	panic("TODO: реализуй Register")
}

// writeError это единственное место, где доменные ошибки превращаются в HTTP (errors.Is / errors.As):
//
//	*library.ValidationError  422 code "validation_failed", message "проверь поля запроса",
//	                          fields {<Field ошибки>: <Message ошибки>}
//	library.ErrNotFound       404 code "not_found"
//	library.ErrISBNTaken      409 code "isbn_taken"
//	library.ErrNoCopies       409 code "no_copies"
//	всё остальное             500 code "internal", message "внутренняя ошибка"
//	                          (текст исходной ошибки клиенту не показывать: там бывают SQL и пароли)
func writeError(c *gin.Context, err error) {
	panic("TODO: реализуй writeError")
}
