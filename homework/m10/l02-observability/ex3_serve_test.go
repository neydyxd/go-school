package l02observability

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"sync"
	"sync/atomic"
	"testing"
	"testing/synctest"
	"time"
)

// Тесты Serve идут внутри synctest.Test с listener в памяти поверх net.Pipe:
// ни портов, ни реального времени. Пауза drain и таймауты отсчитываются по
// фейковым часам, поэтому «запрос посреди drain» гарантированно попадает
// в нужное окно, а длительности проверяются точно.

// pipeListener: net.Listener, соединения которого создаёт Dial через net.Pipe.
type pipeListener struct {
	conns  chan net.Conn
	closed chan struct{}
	once   sync.Once
}

func newPipeListener() *pipeListener {
	return &pipeListener{conns: make(chan net.Conn), closed: make(chan struct{})}
}

func (l *pipeListener) Accept() (net.Conn, error) {
	select {
	case c := <-l.conns:
		return c, nil
	case <-l.closed:
		return nil, net.ErrClosed
	}
}

func (l *pipeListener) Close() error {
	l.once.Do(func() { close(l.closed) })
	return nil
}

func (l *pipeListener) Addr() net.Addr { return pipeAddr{} }

// Dial соединяется с сервером; после закрытия listener это «connection refused».
func (l *pipeListener) Dial(ctx context.Context, _, _ string) (net.Conn, error) {
	client, server := net.Pipe()
	select {
	case l.conns <- server:
		return client, nil
	case <-l.closed:
		return nil, errors.New("connection refused: listener закрыт")
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

type pipeAddr struct{}

func (pipeAddr) Network() string { return "pipe" }
func (pipeAddr) String() string  { return "pipe" }

type serveRun struct {
	srv     *http.Server
	client  *http.Client
	ready   *atomic.Bool
	cancel  context.CancelFunc
	done    chan error
	stopped chan struct{} // закрывается, когда Serve вернулась
}

// startServe запускает Serve в горутине. Вызывать внутри synctest.Test.
func startServe(h http.Handler, drain, timeout time.Duration) *serveRun {
	ln := newPipeListener()
	ctx, cancel := context.WithCancel(context.Background())
	r := &serveRun{
		srv:     &http.Server{Handler: h},
		client:  &http.Client{Transport: &http.Transport{DialContext: ln.Dial}},
		ready:   new(atomic.Bool),
		cancel:  cancel,
		done:    make(chan error, 1),
		stopped: make(chan struct{}),
	}
	go func() {
		defer close(r.stopped)
		r.done <- Serve(ctx, r.srv, ln, r.ready, drain, timeout)
	}()
	return r
}

func (r *serveRun) get() (string, error) {
	resp, err := r.client.Get("http://pipe/")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	return string(b), err
}

// wait ждёт возврата Serve не дольше limit фейкового времени. Если Serve
// зависла, закрывает сервер (чтобы пузырь synctest мог завершиться) и валит тест.
func (r *serveRun) wait(t *testing.T, limit time.Duration) error {
	t.Helper()
	select {
	case err := <-r.done:
		return err
	case <-time.After(limit):
		r.srv.Close()
		t.Fatalf("Serve не вернулась за %v после отмены ctx", limit)
		return nil
	}
}

// finish останавливает сервер и дожидается возврата Serve, чтобы пузырь
// synctest завершился штатно даже после t.Fatal на неверном решении
// (иначе Serve, спящая в паузе drain, даёт panic: deadlock и скрывает
// результаты остальных тестов).
func (r *serveRun) finish() {
	r.cancel()
	select {
	case <-r.stopped:
	case <-time.After(time.Hour): // фейковое время: дожидаемся паузы drain и таймаутов
		r.srv.Close()
		<-r.stopped
	}
	r.srv.Close()
	r.client.CloseIdleConnections()
}

var okHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { io.WriteString(w, "ok") })

func TestServeReadyAndDrain(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		r := startServe(okHandler, 400*time.Millisecond, time.Second)
		defer r.finish()
		synctest.Wait()
		if !r.ready.Load() {
			t.Fatal("ready = false после старта, want true")
		}
		if body, err := r.get(); err != nil || body != "ok" {
			t.Fatalf("GET до остановки = %q, %v, want ok", body, err)
		}

		start := time.Now()
		r.cancel()
		synctest.Wait() // Serve дошла до паузы drain
		if r.ready.Load() {
			t.Fatal("ready = true после отмены ctx, want false сразу, до паузы drain")
		}
		time.Sleep(399 * time.Millisecond) // конец drain, но ещё не Shutdown
		if body, err := r.get(); err != nil || body != "ok" {
			t.Errorf("GET во время drain = %q, %v, want ok: сервер закрылся раньше, чем балансировщик убрал его из ротации", body, err)
		}
		if err := r.wait(t, 5*time.Second); err != nil {
			t.Errorf("Serve = %v, want nil при чистой остановке", err)
		}
		if d := time.Since(start); d != 400*time.Millisecond {
			t.Errorf("Serve остановилась через %v после отмены, want ровно drain 400ms (активных запросов нет)", d)
		}
		r.client.CloseIdleConnections()
		if _, err := r.get(); err == nil {
			t.Error("GET после остановки прошёл, want ошибку: сервер должен закрыть listener")
		}
	})
}

func TestServeWaitsInFlight(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		release := make(chan struct{})
		h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			select {
			case <-release:
				io.WriteString(w, "done")
			case <-r.Context().Done(): // соединение оборвали (неверное решение)
			}
		})
		r := startServe(h, 0, 3*time.Second)
		defer r.finish()
		synctest.Wait()

		type res struct {
			body string
			err  error
		}
		got := make(chan res, 1)
		go func() {
			b, err := r.get()
			got <- res{b, err}
		}()
		synctest.Wait() // обработчик ждёт release

		start := time.Now()
		r.cancel()
		time.AfterFunc(200*time.Millisecond, func() { close(release) })
		if err := r.wait(t, 10*time.Second); err != nil {
			t.Errorf("Serve = %v, want nil: запрос успел завершиться за timeout", err)
		}
		// Shutdown опрашивает активные соединения с растущим интервалом
		// (со случайным разбросом), поэтому точного момента нет: важно, что не
		// раньше конца запроса и заметно раньше timeout.
		if d := time.Since(start); d < 200*time.Millisecond || d >= time.Second {
			t.Errorf("Serve вернулась через %v, want от 200ms (конец запроса) до 1s", d)
		}
		g := <-got
		if g.err != nil || g.body != "done" {
			t.Errorf("активный запрос получил %q, %v, want \"done\": его оборвали", g.body, g.err)
		}
	})
}

func TestServeShutdownTimeout(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			<-r.Context().Done() // висит, пока соединение не закроют
		})
		r := startServe(h, 0, 300*time.Millisecond)
		defer r.finish()
		synctest.Wait()
		go r.get()
		synctest.Wait() // запрос висит в обработчике

		start := time.Now()
		r.cancel()
		err := r.wait(t, 10*time.Second)
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Errorf("Serve = %v, want context.DeadlineExceeded: запрос не успел за timeout", err)
		}
		if d := time.Since(start); d != 300*time.Millisecond {
			t.Errorf("Serve вернулась через %v, want ровно timeout 300ms", d)
		}
	})
}

func TestServeListenerBroken(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ln := newPipeListener()
		ln.Close()
		ready := new(atomic.Bool)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		done := make(chan error, 1)
		go func() {
			done <- Serve(ctx, &http.Server{Handler: okHandler}, ln, ready, 0, time.Second)
		}()
		select {
		case err := <-done:
			if err == nil || errors.Is(err, http.ErrServerClosed) {
				t.Errorf("Serve(закрытый listener) = %v, want ошибку Serve", err)
			}
			if ready.Load() {
				t.Error("ready = true после падения сервера, want false")
			}
		case <-time.After(time.Minute):
			cancel()
			<-done
			t.Fatal("Serve(закрытый listener) не вернулась: падение сервера нужно отслеживать, а не только ctx")
		}
	})
}
