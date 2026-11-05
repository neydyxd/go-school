package l04httpclient

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// scripted отвечает статусами по очереди; последний повторяется.
func scripted(codes []int, retryAfter string, bodies *[]string) (http.Handler, *atomic.Int64) {
	var n atomic.Int64
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		i := int(n.Add(1)) - 1
		if bodies != nil {
			b, _ := io.ReadAll(r.Body)
			*bodies = append(*bodies, string(b))
		}
		code := codes[min(i, len(codes)-1)]
		if retryAfter != "" {
			w.Header().Set("Retry-After", retryAfter)
		}
		w.WriteHeader(code)
		io.WriteString(w, http.StatusText(code))
	}), &n
}

func fakeSleep(log *[]time.Duration) func(context.Context, time.Duration) error {
	return func(ctx context.Context, d time.Duration) error {
		*log = append(*log, d)
		return ctx.Err()
	}
}

func TestDoRetries(t *testing.T) {
	const ms = time.Millisecond
	tests := []struct {
		name         string
		method       string
		idemKey      bool
		codes        []int
		retryAfter   string
		attempts     int
		wantCode     int
		wantAttempts int64
		wantSleeps   []time.Duration
	}{
		{"GET: 503, 503, 200", "GET", false, []int{503, 503, 200}, "", 5, 200, 3, []time.Duration{250 * ms, 500 * ms}},
		{"экспонента упирается в потолок", "GET", false, []int{502}, "", 6, 502, 6, []time.Duration{250 * ms, 500 * ms, 1000 * ms, 2000 * ms, 2000 * ms}},
		{"504 и 429 тоже временные", "DELETE", false, []int{504, 429, 200}, "", 5, 200, 3, []time.Duration{250 * ms, 500 * ms}},
		{"PUT идемпотентен", "PUT", false, []int{503, 200}, "", 3, 200, 2, []time.Duration{250 * ms}},
		{"POST без ключа не повторяем", "POST", false, []int{503, 200}, "", 5, 503, 1, nil},
		{"PATCH без ключа не повторяем", "PATCH", false, []int{502, 200}, "", 5, 502, 1, nil},
		{"POST с Idempotency-Key повторяем", "POST", true, []int{503, 201}, "", 5, 201, 2, []time.Duration{250 * ms}},
		{"500 не временная", "GET", false, []int{500, 200}, "", 5, 500, 1, nil},
		{"404 не повторяем", "GET", false, []int{404, 200}, "", 5, 404, 1, nil},
		{"успех сразу", "GET", false, []int{200}, "", 5, 200, 1, nil},
		{"MaxAttempts 0 значит одна попытка", "GET", false, []int{503, 200}, "", 0, 503, 1, nil},
		{"попытки кончились: последний ответ как есть", "GET", false, []int{503}, "", 2, 503, 2, []time.Duration{250 * ms}},
		{"Retry-After в секундах", "GET", false, []int{429, 200}, "1", 5, 200, 2, []time.Duration{time.Second}},
		{"Retry-After больше потолка", "GET", false, []int{503, 200}, "120", 5, 200, 2, []time.Duration{2 * time.Second}},
		{"Retry-After датой не понимаем: обычная пауза", "GET", false, []int{503, 200}, "Wed, 21 Oct 2026 07:28:00 GMT", 5, 200, 2, []time.Duration{250 * ms}},
		{"Retry-After у 502 не учитываем", "GET", false, []int{502, 200}, "1", 5, 200, 2, []time.Duration{250 * ms}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h, count := scripted(tt.codes, tt.retryAfter, nil)
			srv, client, tr, _ := newTrackedServer(t, h)
			req, _ := http.NewRequest(tt.method, srv.URL, nil)
			if tt.idemKey {
				req.Header.Set("Idempotency-Key", "k-1")
			}
			var sleeps []time.Duration
			p := Policy{MaxAttempts: tt.attempts, BaseDelay: 250 * ms, MaxDelay: 2 * time.Second, Sleep: fakeSleep(&sleeps)}
			resp, err := Do(client, req, p)
			if err != nil {
				t.Fatalf("Do: %v", err)
			}
			if resp.StatusCode != tt.wantCode {
				t.Errorf("Do: код %d, want %d", resp.StatusCode, tt.wantCode)
			}
			if got := count.Load(); got != tt.wantAttempts {
				t.Errorf("Do: попыток %d, want %d", got, tt.wantAttempts)
			}
			if !slices.Equal(sleeps, tt.wantSleeps) {
				t.Errorf("Do: паузы %v, want %v", sleeps, tt.wantSleeps)
			}
			if o, c := tr.opened.Load(), tr.closed.Load(); o-c != 1 {
				t.Errorf("Do: открыто тел %d, закрыто %d; want закрыты все, кроме последнего", o, c)
			}
			if b, _ := io.ReadAll(resp.Body); string(b) != http.StatusText(tt.wantCode) {
				t.Errorf("Do: тело последнего ответа %q, want %q (его нельзя закрывать)", b, http.StatusText(tt.wantCode))
			}
			resp.Body.Close()
		})
	}
}

func TestDoRetriesBody(t *testing.T) {
	var bodies []string
	h, count := scripted([]int{503, 503, 201}, "", &bodies)
	srv, client, _, _ := newTrackedServer(t, h)
	req, _ := http.NewRequest("POST", srv.URL, strings.NewReader(`{"amount":100}`))
	req.Header.Set("Idempotency-Key", "pay-42")
	var sleeps []time.Duration
	resp, err := Do(client, req, Policy{MaxAttempts: 5, BaseDelay: time.Millisecond, MaxDelay: time.Second, Sleep: fakeSleep(&sleeps)})
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	resp.Body.Close()
	if count.Load() != 3 || resp.StatusCode != 201 {
		t.Fatalf("Do: попыток %d, код %d; want 3 и 201", count.Load(), resp.StatusCode)
	}
	want := []string{`{"amount":100}`, `{"amount":100}`, `{"amount":100}`}
	if !slices.Equal(bodies, want) {
		t.Errorf("тела попыток %q, want %q: тело надо брать заново из GetBody", bodies, want)
	}
}

type onceReader struct{ r io.Reader }

func (o *onceReader) Read(p []byte) (int, error) { return o.r.Read(p) }

func TestDoBodyWithoutGetBody(t *testing.T) {
	h, count := scripted([]int{503, 200}, "", nil)
	srv, client, _, _ := newTrackedServer(t, h)
	req, _ := http.NewRequest("PUT", srv.URL, &onceReader{strings.NewReader("data")})
	var sleeps []time.Duration
	resp, err := Do(client, req, Policy{MaxAttempts: 3, BaseDelay: time.Millisecond, MaxDelay: time.Second, Sleep: fakeSleep(&sleeps)})
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	resp.Body.Close()
	if count.Load() != 1 || resp.StatusCode != 503 {
		t.Errorf("тело без GetBody: попыток %d, код %d; want 1 и 503 (повторить тело нечем)", count.Load(), resp.StatusCode)
	}
}

func TestDoNetworkError(t *testing.T) {
	var n atomic.Int64
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if n.Add(1) == 1 {
			conn, _, _ := http.NewResponseController(w).Hijack()
			conn.Close() // обрыв соединения без ответа
			return
		}
		io.WriteString(w, "ok")
	})
	srv, client, _, _ := newTrackedServer(t, h)
	req, _ := http.NewRequest("GET", srv.URL, nil)
	var sleeps []time.Duration
	resp, err := Do(client, req, Policy{MaxAttempts: 3, BaseDelay: time.Millisecond, MaxDelay: time.Second, Sleep: fakeSleep(&sleeps)})
	if err != nil {
		t.Fatalf("Do после обрыва соединения: %v, want повтор и успех", err)
	}
	resp.Body.Close()
	if resp.StatusCode != 200 || len(sleeps) != 1 {
		t.Errorf("Do: код %d, пауз %d; want 200 и одна пауза", resp.StatusCode, len(sleeps))
	}
}

func TestDoCancelDuringSleep(t *testing.T) {
	h, count := scripted([]int{503}, "", nil)
	srv, client, tr, _ := newTrackedServer(t, h)
	ctx, cancel := context.WithCancel(context.Background())
	req, _ := http.NewRequestWithContext(ctx, "GET", srv.URL, nil)
	sleep := func(ctx context.Context, d time.Duration) error {
		cancel() // пока ждём, пользователь ушёл
		return ctx.Err()
	}
	resp, err := Do(client, req, Policy{MaxAttempts: 5, BaseDelay: time.Second, MaxDelay: time.Second, Sleep: sleep})
	if !errors.Is(err, context.Canceled) {
		t.Errorf("Do = %v, %v; want context.Canceled", resp, err)
	}
	if count.Load() != 1 {
		t.Errorf("после отмены сделано %d попыток, want 1", count.Load())
	}
	tr.checkAllClosed(t)
}

func TestDoCanceledRequestNotRetried(t *testing.T) {
	release := make(chan struct{})
	var n atomic.Int64
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n.Add(1)
		select {
		case <-r.Context().Done():
		case <-release:
		}
	}))
	defer srv.Close()
	defer close(release)
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	req, _ := http.NewRequestWithContext(ctx, "GET", srv.URL, nil)
	var sleeps []time.Duration
	_, err := Do(srv.Client(), req, Policy{MaxAttempts: 5, BaseDelay: time.Millisecond, MaxDelay: time.Second, Sleep: fakeSleep(&sleeps)})
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("Do = %v, want context.DeadlineExceeded", err)
	}
	if len(sleeps) != 0 {
		t.Errorf("после истечения ctx были паузы %v, want ни одной: отменённый запрос не повторяют", sleeps)
	}
}
