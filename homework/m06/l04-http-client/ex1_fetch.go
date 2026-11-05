package l04httpclient

import (
	"context"
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
	panic("TODO: реализуй FetchJSON")
}
