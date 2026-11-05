package l01docker

import (
	"context"
	"fmt"
	"net/http"
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
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("probe %s: статус %d", url, resp.StatusCode)
	}
	return nil
}
