package l01web

import (
	"net/http"
	"slices"
)

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
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin == "" {
			next.ServeHTTP(w, r)
			return
		}
		h := w.Header()
		h.Add("Vary", "Origin")
		ok := slices.Contains(allowed, origin)
		if r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != "" {
			if !ok {
				w.WriteHeader(http.StatusForbidden)
				return
			}
			h.Set("Access-Control-Allow-Origin", origin)
			h.Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE")
			h.Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			h.Set("Access-Control-Max-Age", "600")
			w.WriteHeader(http.StatusNoContent)
			return
		}
		if ok {
			h.Set("Access-Control-Allow-Origin", origin)
		}
		next.ServeHTTP(w, r)
	})
}
