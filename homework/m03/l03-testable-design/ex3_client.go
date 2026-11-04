package l03testabledesign

import (
	"errors"
	"net/http"
)

// Profile это профиль из внешнего API.
type Profile struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// ErrNoProfile означает, что API ответил 404.
var ErrNoProfile = errors.New("profile not found")

// FetchProfile запрашивает профиль у внешнего API клиентом c.
// Клиент и базовый адрес приходят снаружи: в тестах это httptest.Server.
//
// Требования:
//   - запрос GET на baseURL + "/profiles/" + id, например http://127.0.0.1:5555/profiles/42;
//   - заголовок запроса Accept: application/json (http.NewRequest + req.Header.Set);
//   - 200: распарсить JSON-тело в Profile (encoding/json);
//   - 404: вернуть ErrNoProfile (можно обёрнутую);
//   - любой другой статус: ошибка, в тексте которой есть код, например "unexpected status 500";
//   - битый JSON и сетевые ошибки: вернуть ошибку;
//   - тело ответа закрыть всегда: defer resp.Body.Close().
func FetchProfile(c *http.Client, baseURL string, id int) (Profile, error) {
	panic("TODO: реализуй FetchProfile")
}
