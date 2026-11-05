package shortener

import "net/http"

// NewHandler собирает HTTP API сокращателя. baseURL без слеша в конце,
// например "https://sho.rt"; из него строится короткая ссылка.
//
//	POST /api/links                  {"url":"https://go.dev"}
//	  201, Location: <baseURL>/<code>,
//	  тело {"code":"abc","short_url":"https://sho.rt/abc","url":"https://go.dev"}
//	GET /{code}                      302, Location: исходный url, Hits++
//	GET /api/links/{code}/stats      200, тело Link: {"code","url","hits","created_at"}
//
// Все ответы с телом идут с Content-Type: application/json.
// Ошибки в едином формате {"error":{"code":"...","message":"..."}}:
//
//	Content-Type запроса не application/json   415 unsupported_media_type
//	тело больше 4 КБ                           413 body_too_large
//	битый JSON, лишние поля, пустое тело       400 invalid_json
//	ValidateURL не прошёл                      422 invalid_url
//	кода нет                                   404 not_found
//	ErrCodeExhausted и прочее                  500 internal
//
// Редирект именно 302, а не 301: браузер кеширует 301 навсегда, и
// повторные переходы не дойдут до сервера (и до счётчика).
// Чужой метод на известном пути: 405 и Allow, это сделает ServeMux.
func NewHandler(s *Store, baseURL string) http.Handler {
	panic("TODO: реализуй NewHandler")
}
