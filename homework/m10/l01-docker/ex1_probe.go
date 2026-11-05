package l01docker

import (
	"context"
	"time"
)

// Probe проверяет здоровье сервиса: делает GET на url и ждёт ответа не дольше
// timeout (и не дольше, чем живёт ctx). Это то, что выполняет HEALTHCHECK
// в distroless-образе, где нет ни curl, ни wget: бинарник проверяет сам себя.
//
//   - Ответ 2xx: вернуть nil.
//   - Другой код: вернуть ошибку, в тексте которой есть сам код (например, "503").
//   - Сервер не отвечает, соединение отклонено, истёк таймаут или ctx: ошибка.
//
// Тело ответа нужно закрыть, а клиент использовать с общим транспортом
// (http.DefaultClient или свой http.Client без своего Transport), чтобы
// повторные проверки шли по одному keep-alive соединению.
func Probe(ctx context.Context, url string, timeout time.Duration) error {
	panic("TODO: реализуй Probe")
}
