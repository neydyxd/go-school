package l02nethttp

import "net/http"

// NewRouter собирает маршруты на ServeMux с паттернами Go 1.22+.
// Хендлеры отвечают простым текстом (fmt.Fprint в w):
//
//	GET    /                     200 "home"            (только корень, не "всё подряд")
//	GET    /users/{id}           200 "user <id>"
//	GET    /users/me             200 "me"              (точный путь важнее wildcard)
//	POST   /users                201 "created"
//	DELETE /users/{id}           204 без тела
//	GET    /files/{path...}      200 "file <path>"     (path со слешами: "a/b/c.txt")
//
// Всё остальное разруливает сам ServeMux: неизвестный путь 404, известный путь
// с чужим методом 405 и заголовок Allow. Например, POST /users/42 даёт 405
// и Allow: DELETE, GET, HEAD. HEAD для GET-маршрутов работает сам.
func NewRouter() *http.ServeMux {
	mux := http.NewServeMux()
	// TODO: зарегистрируй маршруты
	return mux
}
