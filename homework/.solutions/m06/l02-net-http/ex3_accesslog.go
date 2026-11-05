package l02nethttp

import "net/http"

// Entry: одна строка access-лога.
type Entry struct {
	Method  string
	Path    string // r.URL.Path
	Pattern string // паттерн ServeMux, который сработал ("GET /users/{id}"), "" если ни один
	Status  int
	Bytes   int // сколько байт тела записал хендлер
}

// AccessLog возвращает middleware, который вызывает next, а после него
// отдаёт в log одну Entry про этот запрос.
//
// Подводные камни ResponseWriter, которые надо учесть:
//   - хендлер может не вызвать WriteHeader вовсе: тогда статус 200
//     (и когда он что-то записал через Write, и когда не записал ничего);
//   - повторный WriteHeader игнорируется net/http, значит в лог идёт ПЕРВЫЙ статус;
//   - Bytes: сумма того, что вернули вызовы Write;
//   - обёртка над ResponseWriter должна иметь метод Unwrap() http.ResponseWriter,
//     чтобы через неё работал http.NewResponseController (Flush, дедлайны);
//   - Pattern: ServeMux записывает его в r.Pattern того же *http.Request,
//     так что после next его видно и снаружи.
func AccessLog(log func(Entry)) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rec := &statusRecorder{ResponseWriter: w}
			next.ServeHTTP(rec, r)
			if rec.status == 0 {
				rec.status = http.StatusOK
			}
			log(Entry{Method: r.Method, Path: r.URL.Path, Pattern: r.Pattern, Status: rec.status, Bytes: rec.bytes})
		})
	}
}

type statusRecorder struct {
	http.ResponseWriter
	status int
	bytes  int
}

func (s *statusRecorder) WriteHeader(code int) {
	if s.status == 0 {
		s.status = code
	}
	s.ResponseWriter.WriteHeader(code)
}

func (s *statusRecorder) Write(b []byte) (int, error) {
	if s.status == 0 {
		s.status = http.StatusOK
	}
	n, err := s.ResponseWriter.Write(b)
	s.bytes += n
	return n, err
}

func (s *statusRecorder) Unwrap() http.ResponseWriter { return s.ResponseWriter }
