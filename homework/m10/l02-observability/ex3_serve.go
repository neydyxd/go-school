package l02observability

import (
	"context"
	"net"
	"net/http"
	"sync/atomic"
	"time"
)

// Serve обслуживает srv на ln, пока не отменят ctx (SIGTERM в main), и
// корректно останавливается. ready это флаг, который читает /readyz.
//
//  1. Запустить srv.Serve(ln) и сразу выставить ready в true.
//  2. Если сервер упал сам (Serve вернул ошибку не http.ErrServerClosed):
//     ready в false и вернуть эту ошибку.
//  3. Когда ctx отменён: ready в false, затем подождать drain. Всё это
//     время сервер ещё принимает запросы: балансировщик замечает 503 на
//     /readyz не мгновенно и успевает прислать ещё немного трафика.
//  4. srv.Shutdown с таймаутом timeout: новые соединения не принимаются,
//     активные запросы дообслуживаются. Не успели: вернуть ошибку
//     Shutdown (context.DeadlineExceeded) и закрыть оставшиеся соединения
//     через srv.Close. Успели: вернуть nil.
func Serve(ctx context.Context, srv *http.Server, ln net.Listener, ready *atomic.Bool, drain, timeout time.Duration) error {
	panic("TODO: реализуй Serve")
}
