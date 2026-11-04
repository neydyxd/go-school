package l03testabledesign

import (
	"encoding/json"
	"net/http"
	"time"
)

// StatusHandler возвращает HTTP-обработчик для проверки живости сервиса.
func StatusHandler(c Clock) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.Header().Set("Allow", http.MethodGet)
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		// Ошибку записи игнорируем сознательно: заголовок уже ушёл, клиент
		// скорее всего отключился, и исправить ответ нельзя.
		_ = json.NewEncoder(w).Encode(struct {
			Status string `json:"status"`
			Time   string `json:"time"`
		}{"ok", c.Now().UTC().Format(time.RFC3339)})
	})
}
