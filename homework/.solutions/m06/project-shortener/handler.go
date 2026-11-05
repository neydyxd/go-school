package shortener

import (
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
)

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
	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/links", func(w http.ResponseWriter, r *http.Request) {
		if mt, _, err := mime.ParseMediaType(r.Header.Get("Content-Type")); err != nil || mt != "application/json" {
			writeError(w, http.StatusUnsupportedMediaType, "unsupported_media_type", "нужен Content-Type: application/json")
			return
		}
		var in struct {
			URL string `json:"url"`
		}
		dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4<<10))
		dec.DisallowUnknownFields()
		err := dec.Decode(&in)
		if err == nil && dec.Decode(&struct{}{}) != io.EOF {
			err = errors.New("лишние данные после JSON")
		}
		if err != nil {
			var tooBig *http.MaxBytesError
			if errors.As(err, &tooBig) {
				writeError(w, http.StatusRequestEntityTooLarge, "body_too_large", "тело больше 4 КБ")
				return
			}
			writeError(w, http.StatusBadRequest, "invalid_json", err.Error())
			return
		}
		if err := ValidateURL(in.URL); err != nil {
			writeError(w, http.StatusUnprocessableEntity, "invalid_url", err.Error())
			return
		}
		l, err := s.Create(in.URL)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal", "internal server error")
			return
		}
		short := baseURL + "/" + l.Code
		w.Header().Set("Location", short)
		writeJSON(w, http.StatusCreated, map[string]string{"code": l.Code, "short_url": short, "url": l.URL})
	})

	mux.HandleFunc("GET /{code}", func(w http.ResponseWriter, r *http.Request) {
		u, err := s.Resolve(r.PathValue("code"))
		if err != nil {
			writeError(w, http.StatusNotFound, "not_found", "short link not found")
			return
		}
		http.Redirect(w, r, u, http.StatusFound)
	})

	mux.HandleFunc("GET /api/links/{code}/stats", func(w http.ResponseWriter, r *http.Request) {
		l, err := s.Stats(r.PathValue("code"))
		if err != nil {
			writeError(w, http.StatusNotFound, "not_found", "short link not found")
			return
		}
		writeJSON(w, http.StatusOK, l)
	})

	return mux
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, code, msg string) {
	writeJSON(w, status, map[string]map[string]string{"error": {"code": code, "message": msg}})
}
