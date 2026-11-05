package l02nethttp

import (
	"fmt"
	"net/http"
)

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
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "home")
	})
	mux.HandleFunc("GET /users/{id}", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "user ", r.PathValue("id"))
	})
	mux.HandleFunc("GET /users/me", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "me")
	})
	mux.HandleFunc("POST /users", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		fmt.Fprint(w, "created")
	})
	mux.HandleFunc("DELETE /users/{id}", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("GET /files/{path...}", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "file ", r.PathValue("path"))
	})
	return mux
}
