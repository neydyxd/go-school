package l05productionserver

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"testing"
	"time"
)

func TestNewServer(t *testing.T) {
	h := http.NewServeMux()
	s := NewServer(":8081", h)
	if s.Addr != ":8081" || s.Handler != h {
		t.Errorf("NewServer: Addr=%q Handler=%v, want :8081 и переданный хендлер", s.Addr, s.Handler)
	}
	want := map[string][2]time.Duration{
		"ReadHeaderTimeout": {s.ReadHeaderTimeout, 5 * time.Second},
		"ReadTimeout":       {s.ReadTimeout, 15 * time.Second},
		"WriteTimeout":      {s.WriteTimeout, 30 * time.Second},
		"IdleTimeout":       {s.IdleTimeout, 2 * time.Minute},
	}
	for name, v := range want {
		if v[0] != v[1] {
			t.Errorf("%s = %v, want %v", name, v[0], v[1])
		}
	}
}

func startRun(t *testing.T, h http.Handler, timeout time.Duration) (string, context.CancelFunc, chan error) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- Run(ctx, &http.Server{Handler: h}, ln, timeout) }()
	return "http://" + ln.Addr().String(), cancel, done
}

func waitRefused(t *testing.T, addr string) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		c, err := net.DialTimeout("tcp", addr[len("http://"):], 100*time.Millisecond)
		if err != nil {
			return
		}
		c.Close()
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("через 2с после отмены ctx сервер всё ещё принимает новые соединения: Shutdown не вызван?")
}

func TestRunGraceful(t *testing.T) {
	started := make(chan struct{})
	release := make(chan struct{})
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(started)
		<-release
		io.WriteString(w, "done")
	})
	addr, cancel, done := startRun(t, h, 5*time.Second)

	type resp struct {
		code int
		body string
		err  error
	}
	got := make(chan resp, 1)
	go func() {
		r, err := http.Get(addr)
		if err != nil {
			got <- resp{err: err}
			return
		}
		defer r.Body.Close()
		b, _ := io.ReadAll(r.Body)
		got <- resp{code: r.StatusCode, body: string(b)}
	}()

	select {
	case <-started:
	case <-time.After(2 * time.Second):
		t.Fatal("запрос не дошёл до хендлера: Run не запускает Serve?")
	}
	cancel()
	waitRefused(t, addr)
	select {
	case err := <-done:
		t.Fatalf("Run вернулся (%v), пока запрос ещё в полёте: graceful shutdown должен его дождаться", err)
	default:
	}
	close(release)

	select {
	case r := <-got:
		if r.err != nil || r.code != 200 || r.body != "done" {
			t.Errorf("запрос в полёте получил %d %q %v, want 200 \"done\": его оборвали", r.code, r.body, r.err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("запрос в полёте не получил ответ")
	}
	select {
	case err := <-done:
		if err != nil {
			t.Errorf("Run после чистой остановки = %v, want nil (http.ErrServerClosed не ошибка)", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Run не вернулся после завершения запросов")
	}
}

func TestRunShutdownTimeout(t *testing.T) {
	started := make(chan struct{})
	release := make(chan struct{})
	t.Cleanup(func() { close(release) })
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(started)
		<-release // хендлер, который не слушает контекст
	})
	addr, cancel, done := startRun(t, h, 100*time.Millisecond)
	clientErr := make(chan error, 1)
	go func() {
		r, err := http.Get(addr)
		if err == nil {
			r.Body.Close()
		}
		clientErr <- err
	}()
	<-started
	cancel()
	select {
	case err := <-done:
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Errorf("Run при зависшем запросе = %v, want ошибку с context.DeadlineExceeded", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("Run висит дольше 3с при shutdownTimeout 100мс")
	}
	select {
	case err := <-clientErr:
		if err == nil {
			t.Error("зависший запрос получил ответ, want обрыв соединения после srv.Close")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("соединение зависшего запроса не закрыто: после таймаута нужен srv.Close()")
	}
}

func TestRunServeError(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	ln.Close()
	done := make(chan error, 1)
	go func() { done <- Run(context.Background(), &http.Server{}, ln, time.Second) }()
	select {
	case err := <-done:
		if err == nil || errors.Is(err, http.ErrServerClosed) {
			t.Errorf("Run на закрытом листенере = %v, want ошибку Serve", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Run на закрытом листенере не вернулся: ошибку Serve надо ловить и без отмены ctx")
	}
}
