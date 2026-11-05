package l02observability

import (
	"context"
	"errors"
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
	errc := make(chan error, 1)
	go func() { errc <- srv.Serve(ln) }()
	ready.Store(true)

	select {
	case err := <-errc:
		ready.Store(false)
		return err
	case <-ctx.Done():
	}

	ready.Store(false)
	time.Sleep(drain)

	sctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	if err := srv.Shutdown(sctx); err != nil {
		_ = srv.Close()
		return err
	}
	if err := <-errc; !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}
