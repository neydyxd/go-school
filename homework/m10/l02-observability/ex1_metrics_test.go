package l02observability

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

// family находит метрику по имени в реестре.
func family(t *testing.T, reg *prometheus.Registry, name string) *dto.MetricFamily {
	t.Helper()
	mfs, err := reg.Gather()
	if err != nil {
		t.Fatalf("Gather: %v", err)
	}
	for _, mf := range mfs {
		if mf.GetName() == name {
			return mf
		}
	}
	return nil
}

// key склеивает лейблы в порядке method|route|code (каких нет, пропускает).
func key(m *dto.Metric) string {
	l := map[string]string{}
	for _, p := range m.GetLabel() {
		l[p.GetName()] = p.GetValue()
	}
	parts := []string{l["method"], l["route"]}
	if c, ok := l["code"]; ok {
		parts = append(parts, c)
	}
	return strings.Join(parts, "|")
}

func counters(t *testing.T, reg *prometheus.Registry) map[string]float64 {
	t.Helper()
	out := map[string]float64{}
	if mf := family(t, reg, "http_requests_total"); mf != nil {
		for _, m := range mf.GetMetric() {
			out[key(m)] = m.GetCounter().GetValue()
		}
	}
	return out
}

func histograms(t *testing.T, reg *prometheus.Registry) map[string]*dto.Histogram {
	t.Helper()
	out := map[string]*dto.Histogram{}
	if mf := family(t, reg, "http_request_duration_seconds"); mf != nil {
		for _, m := range mf.GetMetric() {
			out[key(m)] = m.GetHistogram()
		}
	}
	return out
}

func inFlight(t *testing.T, reg *prometheus.Registry) float64 {
	t.Helper()
	mf := family(t, reg, "http_requests_in_flight")
	if mf == nil || len(mf.GetMetric()) != 1 {
		t.Fatal("метрика http_requests_in_flight не найдена в реестре (или у неё есть лейблы)")
	}
	return mf.GetMetric()[0].GetGauge().GetValue()
}

func newTestMux(m *Metrics) *http.ServeMux {
	mux := http.NewServeMux()
	handle := func(pattern string, h http.HandlerFunc) { mux.Handle(pattern, m.Middleware(h)) }
	handle("GET /users/{id}", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("user " + r.PathValue("id"))) })
	handle("POST /users", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusCreated) })
	handle("GET /missing", func(w http.ResponseWriter, r *http.Request) { http.NotFound(w, r) })
	handle("GET /double", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
		w.WriteHeader(http.StatusInternalServerError)
	})
	handle("GET /silent", func(w http.ResponseWriter, r *http.Request) {})
	handle("GET /write-then-header", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
		w.WriteHeader(http.StatusInternalServerError) // поздно: клиент уже получил 200
	})
	return mux
}

func do(h http.Handler, method, path string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(method, path, nil))
	return rec
}

func TestMetricsCounter(t *testing.T) {
	reg := prometheus.NewRegistry()
	mux := newTestMux(NewMetrics(reg))
	for _, req := range [][2]string{
		{"GET", "/users/1"}, {"GET", "/users/2"}, {"GET", "/users/3"},
		{"POST", "/users"}, {"GET", "/missing"}, {"GET", "/double"}, {"GET", "/silent"},
		{"GET", "/write-then-header"},
	} {
		do(mux, req[0], req[1])
	}
	want := map[string]float64{
		"GET|GET /users/{id}|200":        3,
		"POST|POST /users|201":           1,
		"GET|GET /missing|404":           1,
		"GET|GET /double|202":            1,
		"GET|GET /silent|200":            1,
		"GET|GET /write-then-header|200": 1,
	}
	if got := counters(t, reg); !reflect.DeepEqual(got, want) {
		t.Errorf("http_requests_total (method|route|code) = %v, want %v", got, want)
	}
}

func TestMetricsResponseUnchanged(t *testing.T) {
	mux := newTestMux(NewMetrics(prometheus.NewRegistry()))
	rec := do(mux, "GET", "/users/42")
	if rec.Code != 200 || rec.Body.String() != "user 42" {
		t.Errorf("GET /users/42 = %d %q, want 200 \"user 42\": middleware не должна менять ответ", rec.Code, rec.Body.String())
	}
	if rec := do(mux, "POST", "/users"); rec.Code != http.StatusCreated {
		t.Errorf("POST /users = %d, want 201", rec.Code)
	}
}

func TestMetricsHistogram(t *testing.T) {
	reg := prometheus.NewRegistry()
	m := NewMetrics(reg)
	mux := newTestMux(m)
	mux.Handle("GET /slow", m.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(30 * time.Millisecond)
	})))
	do(mux, "GET", "/users/1")
	do(mux, "GET", "/users/2")
	do(mux, "GET", "/missing")
	do(mux, "GET", "/slow")

	hs := histograms(t, reg)
	wantCounts := map[string]uint64{"GET|GET /users/{id}": 2, "GET|GET /missing": 1, "GET|GET /slow": 1}
	gotCounts := map[string]uint64{}
	for k, h := range hs {
		gotCounts[k] = h.GetSampleCount()
	}
	if !reflect.DeepEqual(gotCounts, wantCounts) {
		t.Fatalf("http_request_duration_seconds _count (method|route) = %v, want %v (лейблы только method и route)", gotCounts, wantCounts)
	}
	if s := hs["GET|GET /slow"].GetSampleSum(); s < 0.03 || s > 5 {
		t.Errorf("сумма длительностей GET /slow = %vs, want от 0.03 до 5 (секунды, а не миллисекунды)", s)
	}
	var bounds []float64
	for _, b := range hs["GET|GET /slow"].GetBucket() {
		bounds = append(bounds, b.GetUpperBound())
	}
	if !reflect.DeepEqual(bounds, prometheus.DefBuckets) {
		t.Errorf("границы бакетов = %v, want prometheus.DefBuckets %v", bounds, prometheus.DefBuckets)
	}
}

func TestMetricsInFlight(t *testing.T) {
	reg := prometheus.NewRegistry()
	m := NewMetrics(reg)
	entered, release := make(chan struct{}), make(chan struct{})
	h := m.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(entered)
		<-release
	}))
	done := make(chan struct{})
	go func() {
		defer close(done)
		do(h, "GET", "/")
	}()
	select {
	case <-entered:
	case <-time.After(3 * time.Second):
		t.Fatal("обработчик не вызван за 3s")
	}
	if v := inFlight(t, reg); v != 1 {
		t.Errorf("in_flight во время запроса = %v, want 1", v)
	}
	close(release)
	<-done
	if v := inFlight(t, reg); v != 0 {
		t.Errorf("in_flight после запроса = %v, want 0", v)
	}
}

func TestMetricsInFlightPanic(t *testing.T) {
	reg := prometheus.NewRegistry()
	h := NewMetrics(reg).Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("обработчик упал")
	}))
	func() {
		defer func() {
			if recover() == nil {
				t.Error("паника обработчика не дошла наружу: middleware не должна её глотать")
			}
		}()
		do(h, "GET", "/boom")
	}()
	if v := inFlight(t, reg); v != 0 {
		t.Errorf("in_flight после паники в обработчике = %v, want 0 (уменьшай через defer)", v)
	}
}

func TestMetricsUnmatched(t *testing.T) {
	reg := prometheus.NewRegistry()
	h := NewMetrics(reg).Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	}))
	do(h, "DELETE", "/whatever/123")
	want := map[string]float64{"DELETE|unmatched|418": 1}
	if got := counters(t, reg); !reflect.DeepEqual(got, want) {
		t.Errorf("запрос без шаблона ServeMux: http_requests_total = %v, want %v", got, want)
	}
}

func TestMetricsOwnRegistry(t *testing.T) {
	defer func() {
		if p := recover(); p != nil {
			t.Fatalf("второй NewMetrics с новым реестром паникует: %v (регистрируй в reg, а не глобально)", p)
		}
	}()
	NewMetrics(prometheus.NewRegistry())
	NewMetrics(prometheus.NewRegistry())
}

func TestMetricsFlushThroughWrapper(t *testing.T) {
	var flushErr error
	h := NewMetrics(prometheus.NewRegistry()).Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("chunk"))
		flushErr = http.NewResponseController(w).Flush()
	}))
	rec := do(h, "GET", "/stream")
	if flushErr != nil || !rec.Flushed {
		t.Errorf("Flush через обёртку: err = %v, flushed = %v, want nil, true (нужен метод Unwrap)", flushErr, rec.Flushed)
	}
}
