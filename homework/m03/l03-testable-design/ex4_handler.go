package l03testabledesign

import "net/http"

// StatusHandler возвращает HTTP-обработчик для проверки живости сервиса.
// Часы приходят снаружи, поэтому ответ можно проверить точно.
//
// Требования:
//   - GET: статус 200, заголовок Content-Type: application/json, тело
//     {"status":"ok","time":"2026-11-04T09:30:00Z"}, где time это c.Now()
//     в UTC в формате time.RFC3339; удобно собрать через json.NewEncoder(w).Encode
//     со структурой с тегами `json:"status"` и `json:"time"`;
//   - любой другой метод: статус 405 и заголовок Allow: GET
//     (http.Error(w, ..., http.StatusMethodNotAllowed) подойдёт для тела).
//
// Важно: заголовки ставь до первой записи тела. После w.Write или
// w.WriteHeader менять их уже поздно.
func StatusHandler(c Clock) http.Handler {
	panic("TODO: реализуй StatusHandler")
}
