package l02observability

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
)

// Metrics: RED-метрики HTTP-сервиса (Rate, Errors, Duration).
type Metrics struct {
	// TODO: поля для counter, histogram и gauge
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
	panic("TODO: реализуй NewMetrics")
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
	panic("TODO: реализуй Middleware")
}
