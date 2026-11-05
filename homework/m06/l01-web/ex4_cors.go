package l01web

import "net/http"

// CORS оборачивает next и отвечает на CORS-запросы браузера для origin из
// allowed (полное совпадение, например "https://app.example.com").
//
//   - Нет заголовка Origin: это не CORS (curl, сервер-сервер, same-origin GET).
//     Просто вызвать next, никаких CORS-заголовков, Vary не трогать.
//   - Есть Origin: всегда добавить (Add, не Set) заголовок Vary: Origin,
//     чтобы кеш не отдал ответ для одного origin другому.
//   - Preflight: метод OPTIONS и есть заголовок Access-Control-Request-Method.
//     Origin разрешён: ответить 204 с заголовками
//     Access-Control-Allow-Origin: <тот же origin> (не "*"),
//     Access-Control-Allow-Methods: GET, POST, PUT, PATCH, DELETE,
//     Access-Control-Allow-Headers: Content-Type, Authorization,
//     Access-Control-Max-Age: 600.
//     Origin не разрешён: 403 без Access-Control-*-заголовков.
//     В обоих случаях next не вызывается.
//   - Обычный запрос (включая OPTIONS без Access-Control-Request-Method):
//     если origin разрешён, поставить Access-Control-Allow-Origin: <origin>;
//     если нет, заголовок не ставить. next вызывается в любом случае:
//     CORS защищает пользователя в браузере, а не сервер, отказывает браузер.
func CORS(allowed []string, next http.Handler) http.Handler {
	panic("TODO: реализуй CORS")
}
