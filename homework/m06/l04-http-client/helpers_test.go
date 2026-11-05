package l04httpclient

import (
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
)

// trackingTransport считает тела ответов, которые открыли и закрыли.
type trackingTransport struct {
	base   http.RoundTripper
	opened atomic.Int64
	closed atomic.Int64
	trips  atomic.Int64
	read   atomic.Int64 // сколько байт тела прочитал вызывающий код
}

func (t *trackingTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	t.trips.Add(1)
	resp, err := t.base.RoundTrip(r)
	if resp != nil {
		t.opened.Add(1)
		resp.Body = &trackedBody{ReadCloser: resp.Body, t: t}
	}
	return resp, err
}

type trackedBody struct {
	io.ReadCloser
	t    *trackingTransport
	once sync.Once
}

func (b *trackedBody) Read(p []byte) (int, error) {
	n, err := b.ReadCloser.Read(p)
	b.t.read.Add(int64(n))
	return n, err
}

func (b *trackedBody) Close() error {
	b.once.Do(func() { b.t.closed.Add(1) })
	return b.ReadCloser.Close()
}

func (t *trackingTransport) checkAllClosed(tb testing.TB) {
	tb.Helper()
	if o, c := t.opened.Load(), t.closed.Load(); o != c {
		tb.Errorf("открыто тел ответа %d, закрыто %d: каждое тело надо закрыть", o, c)
	}
}

// newTrackedServer поднимает httptest.Server, считает новые TCP-соединения
// и возвращает клиент с отдельным транспортом и трекером тел.
func newTrackedServer(tb testing.TB, h http.Handler) (*httptest.Server, *http.Client, *trackingTransport, *atomic.Int64) {
	tb.Helper()
	var conns atomic.Int64
	srv := httptest.NewUnstartedServer(h)
	srv.Config.ConnState = func(_ net.Conn, s http.ConnState) {
		if s == http.StateNew {
			conns.Add(1)
		}
	}
	srv.Start()
	tb.Cleanup(srv.Close)
	tr := &trackingTransport{base: &http.Transport{}}
	tb.Cleanup(func() { tr.base.(*http.Transport).CloseIdleConnections() })
	return srv, &http.Client{Transport: tr}, tr, &conns
}
