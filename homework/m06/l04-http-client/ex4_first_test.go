package l04httpclient

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strconv"
	"sync/atomic"
	"testing"
	"time"
)

// inflight считает RoundTrip, которые ещё не вернулись.
type inflight struct {
	base http.RoundTripper
	n    atomic.Int64
}

func (t *inflight) RoundTrip(r *http.Request) (*http.Response, error) {
	t.n.Add(1)
	defer t.n.Add(-1)
	return t.base.RoundTrip(r)
}

// mirrors: /fast отвечает сразу, а /fast?wait=N сначала дожидается, пока
// N медленных запросов дошли до сервера (так отмена гарантированно застаёт
// их на сервере, тест детерминирован). /slow ждёт отмены запроса и сообщает
// о ней, /fail отвечает 500.
func mirrors(t *testing.T) (string, *http.Client, *trackingTransport, *inflight, chan string) {
	canceled := make(chan string, 10)
	arrived := make(chan struct{}, 10)
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/fast":
			n, _ := strconv.Atoi(r.URL.Query().Get("wait"))
			for range n {
				select {
				case <-arrived:
				case <-time.After(2 * time.Second):
					t.Error("медленные запросы не дошли до сервера за 2с: First запускает их не параллельно?")
					return
				}
			}
			io.WriteString(w, "mirror-fast")
		case "/fail":
			http.Error(w, "boom", 500)
		case "/slow", "/slow2":
			arrived <- struct{}{}
			select {
			case <-r.Context().Done():
				canceled <- r.URL.Path
			case <-time.After(5 * time.Second):
				io.WriteString(w, "too late")
			}
		}
	})
	srv, client, tr, _ := newTrackedServer(t, h)
	inf := &inflight{base: tr}
	client.Transport = inf
	return srv.URL, client, tr, inf, canceled
}

func runFirst(t *testing.T, ctx context.Context, c *http.Client, urls []string) (string, error) {
	t.Helper()
	type res struct {
		s   string
		err error
	}
	done := make(chan res, 1)
	go func() {
		s, err := First(ctx, c, urls)
		done <- res{s, err}
	}()
	select {
	case r := <-done:
		return r.s, r.err
	case <-time.After(3 * time.Second):
		t.Fatal("First не вернулся за 3с: проигравшие запросы не отменены или горутины ждут друг друга")
		return "", nil
	}
}

func TestFirstWinnerCancelsOthers(t *testing.T) {
	base, client, tr, inf, canceled := mirrors(t)
	got, err := runFirst(t, context.Background(), client, []string{base + "/slow", base + "/fast?wait=2", base + "/slow2"})
	if err != nil || got != "mirror-fast" {
		t.Fatalf("First = %q, %v; want \"mirror-fast\", nil", got, err)
	}
	if n := inf.n.Load(); n != 0 {
		t.Errorf("после возврата First ещё %d запросов в полёте: First должен дождаться всех горутин", n)
	}
	tr.checkAllClosed(t)
	for range 2 {
		select {
		case <-canceled:
		case <-time.After(2 * time.Second):
			t.Fatal("медленный сервер не увидел отмену запроса: проигравших надо отменять")
		}
	}
}

func TestFirstSkipsFailures(t *testing.T) {
	base, client, tr, _, _ := mirrors(t)
	got, err := runFirst(t, context.Background(), client, []string{base + "/fail", base + "/fast"})
	if err != nil || got != "mirror-fast" {
		t.Errorf("First = %q, %v; want \"mirror-fast\", nil: 500 не победитель", got, err)
	}
	tr.checkAllClosed(t)
}

func TestFirstAllFail(t *testing.T) {
	base, client, tr, _, _ := mirrors(t)
	_, err := runFirst(t, context.Background(), client, []string{base + "/fail", base + "/fail"})
	if err == nil {
		t.Error("First(все 500) = nil, want ошибку")
	}
	tr.checkAllClosed(t)
	if _, err := runFirst(t, context.Background(), client, nil); err == nil {
		t.Error("First(пустой список) = nil, want ошибку")
	}
}

func TestFirstParentCanceled(t *testing.T) {
	base, client, _, inf, _ := mirrors(t)
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	_, err := runFirst(t, ctx, client, []string{base + "/slow", base + "/slow2"})
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("First с истёкшим ctx = %v, want ошибку с context.DeadlineExceeded", err)
	}
	if n := inf.n.Load(); n != 0 {
		t.Errorf("после возврата First ещё %d запросов в полёте", n)
	}
}
