package l05productionserver

import (
	"context"
	"net"
	"net/http"
	"time"
)

// NewServer возвращает http.Server для addr и h с выставленными таймаутами
// (у нулевого http.Server их нет вовсе, и это открытая дверь для slowloris):
//
//	ReadHeaderTimeout 5s   сколько ждать заголовков запроса
//	ReadTimeout       15s  весь запрос с телом
//	WriteTimeout      30s  от конца чтения заголовков до конца ответа
//	IdleTimeout       2m   сколько держать keep-alive соединение без запросов
func NewServer(addr string, h http.Handler) *http.Server {
	panic("TODO: реализуй NewServer")
}

// Run обслуживает srv на уже открытом ln, пока не отменят ctx, а потом
// аккуратно останавливается.
//
//   - srv.Serve(ln) блокируется, поэтому крутится в своей горутине.
//   - Serve упал сам (например, ln закрыт) раньше отмены ctx: вернуть его ошибку.
//   - ctx отменён: srv.Shutdown с НОВЫМ контекстом на shutdownTimeout.
//     Shutdown закрывает листенер (новые соединения не принимаются)
//     и ждёт, пока доработают запросы, которые уже в полёте.
//     ctx брать нельзя: он уже отменён, Shutdown сразу сдастся.
//   - Не уложились в shutdownTimeout: srv.Close() (рвём оставшиеся
//     соединения) и вернуть ошибку, для которой
//     errors.Is(err, context.DeadlineExceeded) истинно.
//   - Уложились: вернуть nil. http.ErrServerClosed от Serve это не ошибка,
//     а штатный сигнал «сервер остановлен».
//   - Run возвращается только после того, как горутина с Serve завершилась.
func Run(ctx context.Context, srv *http.Server, ln net.Listener, shutdownTimeout time.Duration) error {
	panic("TODO: реализуй Run")
}
