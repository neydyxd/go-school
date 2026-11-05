package l01web

// IsSafe сообщает, безопасен ли метод по RFC 9110: семантически только
// читает и не меняет состояние на сервере. Безопасные: GET, HEAD, OPTIONS, TRACE.
// Методы регистрозависимы: "get" это не GET, а неизвестный метод.
func IsSafe(method string) bool {
	switch method {
	case "GET", "HEAD", "OPTIONS", "TRACE":
		return true
	}
	return false
}

// IsIdempotent сообщает, идемпотентен ли метод по RFC 9110: повтор того же
// запроса оставляет сервер в том же состоянии, что и один запрос.
// Идемпотентны все безопасные методы, а ещё PUT и DELETE.
// POST, PATCH, CONNECT и неизвестные методы: нет.
func IsIdempotent(method string) bool {
	return IsSafe(method) || method == "PUT" || method == "DELETE"
}
