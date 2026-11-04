package linkcheck

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"testing/synctest"
	"time"
)

// Первый тест ходит в настоящий httptest.NewServer на loopback. Остальные
// работают внутри synctest.Test с httptest.NewTestServer: сеть там в памяти,
// время фейковое, поэтому «запрос на секунду» выполняется мгновенно и
// детерминированно, а зависшая горутина роняет тест с сообщением про deadlock.

// guardTransport стоит на месте http.DefaultTransport. Тестовые адреса
// *.test существуют только в памяти httptest.NewTestServer, дойти до них
// можно лишь через opts.Client. Если решение пошло мимо него, вместо
// настоящего DNS-запроса (который внутри synctest роняет весь бинарник)
// получится обычная ошибка и понятный FAIL.
type guardTransport struct{ next http.RoundTripper }

func (g guardTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	if strings.HasSuffix(r.URL.Hostname(), ".test") {
		return nil, errors.New("запрос ушёл мимо opts.Client")
	}
	return g.next.RoundTrip(r)
}

func init() { http.DefaultTransport = guardTransport{http.DefaultTransport} }

func TestCheckStatuses(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/ok", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("привет")) })
	mux.HandleFunc("/boom", func(w http.ResponseWriter, r *http.Request) { http.Error(w, "упал", 500) })
	mux.Handle("/old", http.RedirectHandler("/ok", http.StatusMovedPermanently))
	srv := httptest.NewServer(mux)
	defer srv.Close()

	urls := []string{srv.URL + "/ok", srv.URL + "/missing", srv.URL + "/boom", srv.URL + "/old", "://кривой", "ftp://example.test/"}
	want := []struct {
		status int
		err    bool
		ok     bool
	}{
		{200, false, true},
		{404, false, false},
		{500, false, false},
		{200, false, true}, // клиент сам идёт по редиректу
		{0, true, false},
		{0, true, false},
	}
	got := Check(context.Background(), urls, Options{Concurrency: 3}) // Client: nil, значит http.DefaultClient
	if len(got) != len(urls) {
		t.Fatalf("Check вернул %d результатов, want %d", len(got), len(urls))
	}
	for i, w := range want {
		r := got[i]
		if r.URL != urls[i] || r.Status != w.status || (r.Err != nil) != w.err || r.OK() != w.ok {
			t.Errorf("results[%d] = {URL:%q Status:%d Err:%v OK:%v}, want {URL:%q Status:%d ошибка:%v OK:%v}",
				i, r.URL, r.Status, r.Err, r.OK(), urls[i], w.status, w.err, w.ok)
		}
	}
}

// countingTransport считает запросы, прошедшие через клиент.
type countingTransport struct{ n atomic.Int32 }

func (c *countingTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	c.n.Add(1)
	return http.DefaultTransport.RoundTrip(r)
}

func TestCheckUsesClient(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	defer srv.Close()
	tr := &countingTransport{}
	Check(context.Background(), []string{srv.URL, srv.URL}, Options{Client: &http.Client{Transport: tr}})
	if n := tr.n.Load(); n != 2 {
		t.Fatalf("через opts.Client прошло %d запросов из 2: если Client задан, ходи через него (дальше тесты на нём держатся)", n)
	}
}

func TestCheckEmpty(t *testing.T) {
	if got := Check(context.Background(), nil, Options{}); len(got) != 0 {
		t.Errorf("Check(nil) = %v, want пусто", got)
	}
}

// slowServer: каждый запрос длится d (в фейковом времени), сервер считает,
// сколько запросов обрабатывается одновременно.
func slowServer(t *testing.T, d time.Duration) (*httptest.Server, *atomic.Int32) {
	var cur, maxSeen atomic.Int32
	srv := httptest.NewTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c := cur.Add(1)
		defer cur.Add(-1)
		for m := maxSeen.Load(); c > m && !maxSeen.CompareAndSwap(m, c); m = maxSeen.Load() {
		}
		select {
		case <-time.After(d):
		case <-r.Context().Done():
		}
	}))
	return srv, &maxSeen
}

func urlsN(n int) []string {
	urls := make([]string, n)
	for i := range urls {
		urls[i] = "http://site.test/page"
	}
	return urls
}

func TestCheckConcurrency(t *testing.T) {
	tests := []struct {
		name        string
		urls        int
		concurrency int
		wantMax     int32
		wantDur     time.Duration
	}{
		{"7 ссылок, лимит 3", 7, 3, 3, 3 * time.Second},
		{"лимит 0 значит 1", 3, 0, 1, 3 * time.Second},
		{"лимит больше ссылок", 4, 100, 4, time.Second},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				srv, maxSeen := slowServer(t, time.Second)
				start := time.Now()
				res := Check(context.Background(), urlsN(tt.urls), Options{Concurrency: tt.concurrency, Client: srv.Client()})
				for i, r := range res {
					if !r.OK() {
						t.Fatalf("results[%d] = {Status:%d Err:%v}, want 200 без ошибки (используй opts.Client)", i, r.Status, r.Err)
					}
				}
				if m := maxSeen.Load(); m != tt.wantMax {
					t.Errorf("сервер видел %d одновременных запросов, want %d", m, tt.wantMax)
				}
				if d := time.Since(start); d != tt.wantDur {
					t.Errorf("Check длился %v, want %v", d, tt.wantDur)
				}
			})
		})
	}
}

func TestCheckTimeout(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		srv := httptest.NewTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/fast":
				time.Sleep(time.Second)
			case "/slow":
				select {
				case <-time.After(time.Hour):
				case <-r.Context().Done():
				}
			case "/stall-body":
				// Заголовки отдали сразу, а тело тянем: таймаут должен
				// покрывать и чтение тела.
				w.Write([]byte("начало"))
				w.(http.Flusher).Flush()
				select {
				case <-time.After(time.Hour):
				case <-r.Context().Done():
				}
			}
		}))
		urls := []string{"http://site.test/fast", "http://site.test/slow", "http://site.test/stall-body"}
		start := time.Now()
		res := Check(context.Background(), urls, Options{Concurrency: 3, Timeout: 2 * time.Second, Client: srv.Client()})
		if d := time.Since(start); d != 2*time.Second {
			t.Errorf("Check длился %v, want 2s (таймаут на запрос)", d)
		}
		if !res[0].OK() {
			t.Errorf("/fast: {Status:%d Err:%v}, want 200 без ошибки", res[0].Status, res[0].Err)
		}
		if !errors.Is(res[1].Err, context.DeadlineExceeded) || res[1].Status != 0 {
			t.Errorf("/slow: {Status:%d Err:%v}, want Status 0 и DeadlineExceeded", res[1].Status, res[1].Err)
		}
		if !errors.Is(res[2].Err, context.DeadlineExceeded) || res[2].Status != 200 || res[2].OK() {
			t.Errorf("/stall-body: {Status:%d Err:%v OK:%v}, want Status 200, DeadlineExceeded при чтении тела, OK false",
				res[2].Status, res[2].Err, res[2].OK())
		}
	})
}

func TestCheckCancel(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var started atomic.Int32
		srv := httptest.NewTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			started.Add(1)
			d := time.Second
			if r.URL.Path == "/slow" {
				d = time.Hour
			}
			select {
			case <-time.After(d):
			case <-r.Context().Done():
			}
		}))
		urls := []string{"http://site.test/fast", "http://site.test/fast", "http://site.test/slow",
			"http://site.test/slow", "http://site.test/fast", "http://site.test/fast"}
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		time.AfterFunc(1500*time.Millisecond, cancel)

		start := time.Now()
		res := Check(ctx, urls, Options{Concurrency: 2, Client: srv.Client()})
		if d := time.Since(start); d != 1500*time.Millisecond {
			t.Errorf("Check длился %v после отмены в 1.5s, want ровно 1.5s", d)
		}
		if len(res) != len(urls) {
			t.Fatalf("Check вернул %d результатов, want %d", len(res), len(urls))
		}
		for i := range 2 {
			if !res[i].OK() {
				t.Errorf("results[%d] = {Status:%d Err:%v}, want 200: успели до отмены", i, res[i].Status, res[i].Err)
			}
		}
		for i := 2; i < len(res); i++ {
			if res[i].URL != urls[i] || !errors.Is(res[i].Err, context.Canceled) {
				t.Errorf("results[%d] = {URL:%q Err:%v}, want URL на месте и context.Canceled", i, res[i].URL, res[i].Err)
			}
		}
		synctest.Wait()
		if n := started.Load(); n != 4 {
			t.Errorf("сервер получил %d запросов, want 4: после отмены новые запросы не начинаются", n)
		}
	})
}
