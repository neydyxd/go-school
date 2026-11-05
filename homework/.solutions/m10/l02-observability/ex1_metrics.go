package l02observability

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// Metrics: RED-метрики HTTP-сервиса (Rate, Errors, Duration).
type Metrics struct {
	requests *prometheus.CounterVec
	duration *prometheus.HistogramVec
	inFlight prometheus.Gauge
}

// NewMetrics создаёт метрики и регистрирует их в reg (не в глобальном
// реестре: тесты и несколько экземпляров сервиса в одном процессе не должны
// конфликтовать). Три метрики:
//
//   - http_requests_total: counter с лейблами method, route, code;
//   - http_request_duration_seconds: histogram с лейблами method, route
//     и бакетами prometheus.DefBuckets;
//   - http_requests_in_flight: gauge без лейблов, сколько запросов
//     обрабатывается прямо сейчас.
func NewMetrics(reg prometheus.Registerer) *Metrics {
	m := &Metrics{
		requests: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Число обработанных HTTP-запросов.",
		}, []string{"method", "route", "code"}),
		duration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Время обработки HTTP-запроса.",
			Buckets: prometheus.DefBuckets,
		}, []string{"method", "route"}),
		inFlight: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "http_requests_in_flight",
			Help: "Запросы, которые обрабатываются прямо сейчас.",
		}),
	}
	reg.MustRegister(m.requests, m.duration, m.inFlight)
	return m
}

// Middleware оборачивает обработчик, зарегистрированный в http.ServeMux
// (шаблоны вида "GET /users/{id}" с Go 1.22, поле r.Pattern с Go 1.23),
// и пишет метрики каждого запроса.
//
//   - route: шаблон, который сопоставил ServeMux, то есть r.Pattern
//     ("GET /users/{id}"), а не r.URL.Path ("/users/42"): иначе каждый id
//     станет отдельным временным рядом. Пустой Pattern: route "unmatched".
//   - code: статус ответа строкой ("200", "404"). Обработчик, который ничего
//     не вызвал или сразу вызвал Write, ответил 200. Повторный WriteHeader
//     ничего не меняет, считается первый; WriteHeader(500) после Write тоже
//     ничего не меняет: клиент уже получил 200.
//   - duration: секунды от входа в Middleware до выхода из next.
//   - in_flight: +1 на входе, -1 на выходе, в том числе если next
//     паникует (панику не перехватывать: её ловит recovery снаружи).
//
// Обёртка над ResponseWriter обязана иметь метод Unwrap() http.ResponseWriter,
// чтобы http.ResponseController мог добраться до Flush исходного writer.
func (m *Metrics) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m.inFlight.Inc()
		defer m.inFlight.Dec()

		route := r.Pattern
		if route == "" {
			route = "unmatched"
		}
		rec := &statusRecorder{ResponseWriter: w}
		start := time.Now()
		next.ServeHTTP(rec, r)

		m.duration.WithLabelValues(r.Method, route).Observe(time.Since(start).Seconds())
		if rec.code == 0 {
			rec.code = http.StatusOK
		}
		m.requests.WithLabelValues(r.Method, route, strconv.Itoa(rec.code)).Inc()
	})
}

type statusRecorder struct {
	http.ResponseWriter
	code int
}

func (s *statusRecorder) WriteHeader(code int) {
	if s.code == 0 {
		s.code = code
	}
	s.ResponseWriter.WriteHeader(code)
}

func (s *statusRecorder) Write(b []byte) (int, error) {
	if s.code == 0 {
		s.code = http.StatusOK
	}
	return s.ResponseWriter.Write(b)
}

func (s *statusRecorder) Unwrap() http.ResponseWriter { return s.ResponseWriter }
