package l02nethttp

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
	panic("TODO: реализуй AccessLog")
}
