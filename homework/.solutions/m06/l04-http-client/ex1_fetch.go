package l04httpclient

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
)

// StatusError: сервер ответил не 2xx.
type StatusError struct {
	Code int
	Body string // начало тела ответа, не больше 512 байт: для логов и отладки
}

func (e *StatusError) Error() string {
	return "unexpected status " + strconv.Itoa(e.Code) + ": " + e.Body
}

// FetchJSON делает GET url клиентом c с контекстом ctx и декодирует
// JSON-ответ в dst.
//
//   - Статус 2xx: декодировать тело в dst, ошибку декодирования вернуть.
//   - Другой статус: вернуть *StatusError с кодом и первыми 512 байтами
//     тела (io.LimitReader). Тело в мегабайт читать целиком не надо.
//   - Ошибка сети или отмена ctx: вернуть её (обёрнутой можно), чтобы
//     errors.Is(err, context.Canceled) работал.
//   - Тело ответа закрыть ВСЕГДА, на любом пути, где resp != nil.
//     Иначе соединение не вернётся в пул и каждый запрос будет открывать новое.
func FetchJSON(ctx context.Context, c *http.Client, url string, dst any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	resp, err := c.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return &StatusError{Code: resp.StatusCode, Body: string(b)}
	}
	return json.NewDecoder(resp.Body).Decode(dst)
}
