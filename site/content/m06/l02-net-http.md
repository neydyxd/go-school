# Сервер на net/http

В Go не нужен фреймворк, чтобы написать HTTP-сервер для продакшена. Пакет `net/http` из стандартной библиотеки держит нагрузку крупных сервисов, умеет HTTP/2 и TLS, а с Go 1.22 его роутер `ServeMux` понимает методы и параметры в пути. Gin, chi и echo, которые ты встретишь в вакансиях, построены поверх тех же интерфейсов `http.Handler` и `http.ResponseWriter`. Поняв их, ты поймёшь любой Go-фреймворк за вечер.

Этот урок про три вещи. Первая: интерфейс `Handler`, на котором держится всё. Вторая: маршрутизация через `ServeMux` с паттернами Go 1.22. Третья: middleware, контекст запроса и ловушки `ResponseWriter`, на которых спотыкаются даже опытные разработчики.

На собесе здесь любят спрашивать: «что такое `http.HandlerFunc` и зачем он нужен», «в каком порядке выполняются middleware», «что будет, если вызвать `WriteHeader` после `Write`», «как передать user ID из middleware в хендлер».

## Handler: один метод на всё

Всё в `net/http` крутится вокруг маленького интерфейса:

```go
type Handler interface {
	ServeHTTP(ResponseWriter, *Request)
}
```

Сервер принимает соединение, разбирает запрос в `*http.Request` и вызывает `ServeHTTP`, передав `ResponseWriter`, через который ты пишешь ответ. Всё. Роутер, middleware, файловый сервер: каждый из них просто `Handler`.

```go
type hello struct{ greeting string }

func (h hello) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "%s, %s!", h.greeting, r.URL.Query().Get("name"))
}

func main() {
	http.ListenAndServe(":8080", hello{"Привет"})
}
```

```bash
curl 'localhost:8080/?name=Степан'
# Привет, Степан!
```

### HandlerFunc: функция как интерфейс

Заводить тип ради одного метода неудобно. Поэтому есть адаптер:

```go
type HandlerFunc func(ResponseWriter, *Request)

func (f HandlerFunc) ServeHTTP(w ResponseWriter, r *Request) {
	f(w, r)
}
```

Это **функциональный тип с методом**: любая функция с подходящей сигнатурой, приведённая к `HandlerFunc`, становится `Handler`. Этот приём «функциональный тип реализует интерфейс» спрашивают на собесах:

```go
var h http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("ok"))
})
```

`mux.HandleFunc(pattern, f)` делает это приведение за тебя.

### Что лежит в Request

- `r.Method`, `r.URL.Path`, `r.URL.Query()` (это `url.Values`, `map[string][]string`).
- `r.Header`: заголовки запроса, канонизированные.
- `r.Body`: тело как `io.ReadCloser`. Сервер закроет его сам после хендлера, но читать надо с лимитом (об этом в следующем уроке).
- `r.Context()`: контекст запроса.
- `r.PathValue("id")`: значение wildcard из паттерна (Go 1.22+).
- `r.Pattern`: паттерн `ServeMux`, который сработал (Go 1.23+).
- `r.RemoteAddr`: IP:порт того, кто подключился (за прокси это прокси).

::: hood Горутина на соединение
`http.Server` в цикле делает `Accept` и на каждое соединение запускает горутину `conn.serve`. Внутри она читает запросы по одному (HTTP/1.1 keep-alive) и для каждого вызывает твой `ServeHTTP`. Для HTTP/2 на соединение одна горутина чтения, а каждый поток (запрос) обслуживается в своей горутине.

Отсюда два вывода. Первый: хендлер **вызывается конкурентно** из многих горутин, любое общее состояние (мапа, счётчик) нужно защищать мьютексом или атомиками. Второй: горутины дешёвые, поэтому модель «одна горутина на запрос» и блокирующий код внутри хендлера это нормально, никаких колбэков и async/await. Планировщик Go сам переключит горутину, пока она ждёт базу.
:::

## ServeMux: маршрутизация в Go 1.22+

`http.ServeMux` сопоставляет запрос с зарегистрированными паттернами. До Go 1.22 он умел только пути и префиксы, поэтому все брали gorilla/mux или chi. Теперь паттерн выглядит так:

```text
[МЕТОД ][ХОСТ]/ПУТЬ
```

```go
mux := http.NewServeMux()
mux.HandleFunc("GET /{$}", home)                  // ровно "/"
mux.HandleFunc("GET /users/{id}", getUser)        // один сегмент
mux.HandleFunc("GET /users/me", getMe)            // точный путь
mux.HandleFunc("POST /users", createUser)
mux.HandleFunc("DELETE /users/{id}", deleteUser)
mux.HandleFunc("GET /files/{path...}", getFile)   // остаток пути со слешами
mux.HandleFunc("api.example.com/", apiHost)       // по хосту
```

Правила:

- **Метод** в начале ограничивает паттерн. `GET` автоматически обслуживает и `HEAD`. Паттерн без метода ловит все методы.
- **`{name}`** соответствует ровно одному сегменту пути. Значение достаёшь `r.PathValue("name")`, процентная кодировка уже раскодирована: `/users/%D0%B8` даст `"и"`.
- **`{name...}`** в конце ловит весь остаток, включая слеши: `/files/a/b/c.txt` даст `"a/b/c.txt"`.
- **Слеш на конце** делает паттерн префиксным: `/static/` ловит `/static/` и всё под ним. Чтобы совпадение было точным, ставь `{$}`: `/posts/{$}` ловит только `/posts/`. А `/` сам по себе ловит **всё**, поэтому корень пишут как `/{$}`.
- Запрос `/files` при наличии `/files/{path...}` получит редирект на `/files/`. С Go 1.26 этот редирект `307`, раньше был `301`.

### Кто победит: правило специфичности

Если запросу подходят несколько паттернов, выигрывает **более специфичный**: тот, который подходит строго к меньшему набору запросов. `/users/me` специфичнее `/users/{id}`, `GET /x` специфичнее `/x`.

```go
mux.HandleFunc("/users/{id}", func(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "id=", r.PathValue("id"))
})
mux.HandleFunc("GET /users/me", func(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "me")
})
for _, m := range []string{"GET", "POST"} {
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, httptest.NewRequest(m, "/users/me", nil))
	fmt.Println(m, w.Body.String())
}
```

<Predict :options="['GET me / POST me', 'GET me / POST id=me', 'GET id=me / POST id=me', 'GET me / POST 405']" :correct="1">

Для `GET /users/me` подходят оба паттерна, и `GET /users/me` специфичнее (точный путь и метод). Для `POST` паттерн с `GET` не подходит вовсе, остаётся `/users/{id}` без метода, а он ловит любой метод. Поэтому 405 тут не будет: путь обслуживается паттерном «на все методы».

</Predict>

Если ни один из двух паттернов не специфичнее другого, но они пересекаются, `ServeMux` **паникует при регистрации**, а не молча выбирает:

```go
mux.HandleFunc("/posts/{id}", a)
mux.HandleFunc("/{resource}/latest", b)
// panic: pattern "/{resource}/latest" ... conflicts with pattern "/posts/{id}" ...
// both match some paths, like "/posts/latest".
```

Это хорошо: конфликт маршрутов всплывает при старте, а не в проде на странном URL.

### 404 и 405 бесплатно

Если путь не подошёл ни к одному паттерну, `ServeMux` ответит `404`. Если путь подошёл, а метод нет, ответ будет `405 Method Not Allowed` с заголовком `Allow`, где перечислены допустимые методы:

```text
POST /users/42  →  405, Allow: DELETE, GET, HEAD
```

Руками писать `if r.Method != http.MethodPost` больше не нужно.

::: trap DefaultServeMux
```go
// плохо: регистрация в глобальном DefaultServeMux
http.HandleFunc("/admin", admin)
http.ListenAndServe(":8080", nil) // nil значит DefaultServeMux
```

`DefaultServeMux` это глобальная переменная, и в неё может что-то зарегистрировать **любой** импортированный пакет. Классика: `import _ "net/http/pprof"` где-то в зависимостях добавляет `/debug/pprof/` с профилями и дампами памяти на публичный порт.

```go
// хорошо: свой mux, явно переданный серверу
mux := http.NewServeMux()
mux.HandleFunc("GET /admin", admin)
http.ListenAndServe(":8080", mux)
```
:::

::: hood Как ServeMux ищет хендлер
С Go 1.22 паттерны хранятся в дереве по сегментам пути (плюс отдельный уровень для хоста и метода). При поиске mux идёт по сегментам запроса, на каждом уровне сначала пробуя точное совпадение, потом wildcard. Специфичность и конфликты считаются **при регистрации**: для новой пары паттернов вычисляется, пересекаются ли множества путей, которые они ловят, и вложено ли одно в другое. Поэтому `Handle` может паниковать, а `ServeHTTP` никогда не сомневается.

Старое поведение (только префиксы, фигурные скобки как обычные символы) можно вернуть `GODEBUG=httpmuxgo121=1`, если обновляешь древний проект.
:::

## Middleware

Middleware это функция, которая принимает хендлер и возвращает новый, добавляя поведение до или после вызова исходного:

```go
func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r) // передать дальше по цепочке
		log.Printf("%s %s %v", r.Method, r.URL.Path, time.Since(start))
	})
}
```

Сигнатура `func(http.Handler) http.Handler` это весь «фреймворк middleware». Логирование, recovery, request ID, аутентификация, CORS, таймауты, сжатие: всё пишется так. Middleware с настройками это функция, возвращающая middleware:

```go
func RequireRole(role string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !hasRole(r, role) {
				http.Error(w, "forbidden", http.StatusForbidden)
				return // next не вызываем: цепочка оборвана
			}
			next.ServeHTTP(w, r)
		})
	}
}
```

### Порядок в цепочке

Middleware вкладываются как матрёшки: `A(B(handler))`. Запрос идёт снаружи внутрь, ответ изнутри наружу.

```text
запрос → A → B → handler
ответ  ← A ← B ← handler
```

Если оборачивать в цикле по списку, легко получить обратный порядок:

```go
handler := http.Handler(h)
for _, m := range []func(http.Handler) http.Handler{mw("A"), mw("B")} {
	handler = m(handler)
}
handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil))
// mw(name) печатает name+">" до next и "<"+name после
```

<Predict :options="['A>B>H <B <A', 'B>A>H <A <B', 'A>B>H <A <B', 'H <A <B']" :correct="1">

Первая итерация обернула `h` в A, вторая обернула результат в B. Самым внешним оказался B, он и выполняется первым. Поэтому функции `Chain` обходят список с конца: тогда первый в списке становится внешним. Порядок важен: recovery должен быть снаружи логирования и всего остального, иначе паника в них не поймается; request ID раньше логирования, иначе в логе не будет ID.

</Predict>

Типичный порядок снаружи внутрь: recovery, request ID, логирование, CORS, аутентификация, лимиты, затем роутер. Для группы маршрутов middleware можно повесить на под-mux:

```go
api := http.NewServeMux()
api.HandleFunc("GET /api/orders", listOrders)

root := http.NewServeMux()
root.Handle("/api/", RequireAuth(api)) // auth только для /api/
root.HandleFunc("GET /healthz", healthz)
```

## Контекст запроса

У каждого запроса есть `context.Context`, который:

- **отменяется**, когда клиент разорвал соединение, когда `ServeHTTP` вернулся, или когда запрос HTTP/2 отменён;
- несёт **значения уровня запроса**: request ID, аутентифицированный пользователь, trace ID.

Отмена это главное. Передавай `r.Context()` во все вызовы базы и внешних сервисов: если пользователь закрыл вкладку, запрос к базе тоже отменится, и сервер не будет тратить ресурсы на ответ, который никто не прочитает.

```go
func getOrder(w http.ResponseWriter, r *http.Request) {
	order, err := repo.Get(r.Context(), r.PathValue("id"))
	if errors.Is(err, context.Canceled) {
		return // клиент ушёл, отвечать некому
	}
	// ...
}
```

Значения кладут в middleware, а читают в хендлере. Ключ должен быть **своего неэкспортируемого типа**, чтобы не столкнуться с ключами других пакетов:

```go
type userKey struct{}

func WithUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u, err := authenticate(r)
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), userKey{}, u)
		next.ServeHTTP(w, r.WithContext(ctx)) // новый *Request с новым контекстом
	})
}

func UserFrom(ctx context.Context) (User, bool) {
	u, ok := ctx.Value(userKey{}).(User)
	return u, ok
}
```

`r.WithContext` возвращает **копию** запроса, исходный не меняется. Значит, middleware снаружи не увидит контекст, который положили внутри.

::: trap Строковый ключ контекста и бизнес-данные в нём
```go
// плохо: строковый ключ и параметр функции, спрятанный в контекст
ctx := context.WithValue(r.Context(), "user", u)
ctx = context.WithValue(ctx, "limit", 20)
```

Строковый ключ `"user"` может перезаписать любой другой пакет, а `staticcheck` на такое ругается (SA1029). Хуже того, через контекст начинают передавать обычные параметры, и сигнатуры функций врут о том, что им нужно.

```go
// хорошо: свой тип ключа, в контексте только данные уровня запроса
ctx := context.WithValue(r.Context(), userKey{}, u)
orders, err := svc.List(ctx, u.ID, 20) // параметры явно
```
:::

## ResponseWriter и его ловушки

`ResponseWriter` это интерфейс из трёх методов:

```go
type ResponseWriter interface {
	Header() Header          // заголовки ответа, менять ДО WriteHeader
	WriteHeader(statusCode int)
	Write([]byte) (int, error)
}
```

Ответ уходит в строгом порядке: строка статуса, заголовки, тело. Отсюда правила:

1. **Сначала заголовки**, потом `WriteHeader`, потом `Write`.
2. **`Write` без `WriteHeader`** неявно вызывает `WriteHeader(200)`.
3. **Повторный `WriteHeader`** игнорируется, а сервер пишет в лог `http: superfluous response.WriteHeader call`.
4. **Изменения заголовков после `WriteHeader`** не попадают в ответ.
5. Если `Content-Type` не выставлен, `net/http` угадает его по первым 512 байтам (`http.DetectContentType`). JSON он распознает как `text/plain; charset=utf-8`.

```go
srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	io.WriteString(w, `{"ok":true}`)
}))
defer srv.Close()

resp, _ := http.Get(srv.URL)
resp.Body.Close()
fmt.Println(resp.StatusCode, resp.Header.Get("Content-Type"))
```

<Predict :options="['201 application/json', '201 text/plain; charset=utf-8', '200 application/json', '201 (пустая строка)']" :correct="1">

`WriteHeader(201)` зафиксировал статус и снимок заголовков. `Set` после этого меняет мапу, которая уже никуда не попадёт. Когда пришёл `Write`, `Content-Type` в снимке не было, и сервер угадал его по содержимому: текст, значит `text/plain; charset=utf-8`. Фронт, который проверяет `Content-Type`, такой ответ как JSON не разберёт.

</Predict>

::: trap Статус после тела
```go
// плохо: сначала тело, потом статус ошибки
func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "не нашли")
	w.WriteHeader(http.StatusNotFound) // поздно: уже ушло 200
}
```

Клиент получит `200` с текстом «не нашли», а в логе сервера будет `superfluous response.WriteHeader call`. Типичная версия той же ошибки: начали стримить JSON через `json.NewEncoder(w).Encode`, кодирование упало на середине, вызвали `http.Error`, а статус 200 уже отправлен.

```go
// хорошо: решить, что отвечать, до первого байта тела
func handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusNotFound)
	fmt.Fprint(w, "не нашли")
}
```

Если ответ сериализуется в JSON, надёжнее закодировать его в `[]byte` (через `json.Marshal`), проверить ошибку и только потом писать заголовки и тело.
:::

### Обёртка над ResponseWriter

Чтобы middleware узнал статус и размер ответа (для логов и метрик), `ResponseWriter` оборачивают:

```go
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (s *statusRecorder) WriteHeader(code int) {
	if s.status == 0 {
		s.status = code
	}
	s.ResponseWriter.WriteHeader(code)
}

func (s *statusRecorder) Unwrap() http.ResponseWriter { return s.ResponseWriter }
```

Встраивание интерфейса отдаёт `Header` и `Write` бесплатно. Но исходный `ResponseWriter` часто умеет больше: `http.Flusher` (для стриминга), `http.Hijacker` (для WebSocket). Обёртка эти методы прячет. Поэтому с Go 1.20 есть `http.NewResponseController(w)`: он ищет нужный метод, разворачивая обёртки через `Unwrap()`. Метод `Unwrap` в обёртке обязателен, иначе `Flush` вернёт ошибку `http.ErrNotSupported`. Это задание из домашки.

### Паники в хендлерах

Если хендлер паникует, `net/http` **перехватывает** панику, пишет стектрейс в лог и **закрывает соединение**. Сервер не падает, но клиент вместо ответа получает обрыв (`EOF`), а не 500. Поэтому recovery-middleware всё равно нужен: он отвечает клиенту нормальной 500 и пишет панику в твой логгер с request ID.

Особый случай `panic(http.ErrAbortHandler)`: штатный способ оборвать ответ на середине, `net/http` не логирует такую панику. Recovery-middleware должен её пропускать дальше.

`http.Error(w, msg, code)` удобен для простых ответов: выставляет `Content-Type: text/plain; charset=utf-8`, `X-Content-Type-Options: nosniff`, статус и тело `msg` с переводом строки. Для JSON API в следующем уроке напишем свой аналог.

## Что выведет программа?

```go
srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "не нашли")
	w.WriteHeader(http.StatusNotFound)
}))
defer srv.Close()

resp, _ := http.Get(srv.URL)
resp.Body.Close()
fmt.Println(resp.StatusCode)
```

<Predict :options="['404', '200', '500', 'паника: superfluous WriteHeader']" :correct="1">

`Fprint` вызвал `Write`, а он неявно отправил `200`. Последующий `WriteHeader(404)` уже ничего не изменит, сервер лишь напишет в лог `http: superfluous response.WriteHeader call from main.main.func1 (main.go:...)`. Паники нет, это предупреждение.

</Predict>

```go
mux := http.NewServeMux()
mux.HandleFunc("GET /users/{id}", getUser)
mux.HandleFunc("DELETE /users/{id}", deleteUser)

w := httptest.NewRecorder()
mux.ServeHTTP(w, httptest.NewRequest("PUT", "/users/42", nil))
fmt.Println(w.Code, w.Header().Get("Allow"))
```

<Predict :options="['404 (пусто)', '405 GET, DELETE', '405 DELETE, GET, HEAD', '200 (пусто)']" :correct="2">

Путь `/users/42` подошёл к паттернам, но метода `PUT` среди них нет, поэтому `405` и заголовок `Allow` со всеми методами, которые есть для этого пути. `HEAD` попал туда, потому что паттерн с `GET` обслуживает и `HEAD`. Методы перечислены по алфавиту.

</Predict>

::: interview
<details><summary>Что такое http.Handler и http.HandlerFunc? Зачем нужен второй?</summary>

`Handler` это интерфейс с одним методом `ServeHTTP(ResponseWriter, *Request)`. `HandlerFunc` это тип `func(ResponseWriter, *Request)` с методом `ServeHTTP`, который просто вызывает саму функцию. Он позволяет использовать обычную функцию там, где ждут интерфейс, без объявления типа. Пример общего приёма «функциональный тип реализует интерфейс».

</details>
<details><summary>Что нового в ServeMux в Go 1.22?</summary>

Методы в паттерне (`GET /users/{id}`), wildcard на сегмент (`{id}`) и на остаток пути (`{path...}`), точное совпадение `{$}`, получение значений через `r.PathValue`. Выбор по специфичности (более узкий паттерн побеждает), паника при регистрации конфликтующих паттернов, автоматические 405 с заголовком `Allow`, `GET` обслуживает `HEAD`. В 1.23 добавилось поле `r.Pattern`. Для простых сервисов сторонний роутер больше не нужен.

</details>
<details><summary>Как написать middleware и в каком порядке они выполняются?</summary>

Функция `func(http.Handler) http.Handler`, которая возвращает `HandlerFunc`, делающий что-то до и/или после `next.ServeHTTP(w, r)` или не вызывающий его вовсе (отказ). Вложение `A(B(h))`: запрос проходит A, потом B, потом h; ответ обратно. Recovery ставят самым внешним, request ID раньше логирования, аутентификацию ближе к роутеру.

</details>
<details><summary>Как передать данные из middleware в хендлер?</summary>

Через контекст запроса: `ctx := context.WithValue(r.Context(), key, value)` и `next.ServeHTTP(w, r.WithContext(ctx))`. Ключ своего неэкспортируемого типа (`type userKey struct{}`), чтобы не было коллизий; рядом типизированный геттер `UserFrom(ctx)`. В контексте только данные уровня запроса (пользователь, request ID, trace), не параметры бизнес-логики.

</details>
<details><summary>Когда отменяется r.Context()?</summary>

Когда клиент закрыл соединение, когда отменён поток HTTP/2 и когда `ServeHTTP` вернулся. Поэтому контекст передают во все вызовы БД и внешних сервисов, а горутины, запущенные из хендлера и живущие дольше него, не должны использовать `r.Context()` (он отменится сразу после ответа), им нужен свой контекст, например `context.WithoutCancel(r.Context())`.

</details>
<details><summary>Что будет, если поменять заголовок после WriteHeader? А вызвать WriteHeader дважды?</summary>

Изменение заголовков после `WriteHeader` не попадёт в ответ: статус и заголовки уже зафиксированы. Повторный `WriteHeader` игнорируется, сервер пишет в лог `superfluous response.WriteHeader call`. `Write` без `WriteHeader` неявно отправляет 200. Отсюда порядок: заголовки, статус, тело, и решение о статусе до первого байта тела.

</details>
<details><summary>Что происходит при панике в хендлере?</summary>

`net/http` ловит панику в горутине соединения, логирует стектрейс и закрывает соединение; сервер продолжает работать, но клиент получает обрыв, а не 500. Поэтому пишут recovery-middleware, который отвечает 500 и логирует через общий логгер. `http.ErrAbortHandler` это штатный способ прервать ответ, его не логируют и пробрасывают дальше. Паника в горутине, которую хендлер сам запустил, не перехватывается никем и роняет весь процесс.

</details>
<details><summary>Зачем http.ResponseController и Unwrap?</summary>

Обёртки над `ResponseWriter` (для логов, метрик, сжатия) прячут дополнительные интерфейсы исходного writer: `Flusher`, `Hijacker`. `http.NewResponseController(w)` (Go 1.20+) находит нужную возможность, разворачивая цепочку через метод `Unwrap() http.ResponseWriter`, и даёт `Flush`, `Hijack`, `SetReadDeadline`, `SetWriteDeadline`. Без `Unwrap` в обёртке эти вызовы вернут `http.ErrNotSupported`.

</details>
:::

## Источники

- `RU` [Разбираемся в новом роутинге в Go 1.22, Хабр (Авито)](https://habr.com/ru/companies/avito/articles/805097/): паттерны, `{$}`, приоритеты и конфликты на примерах.
- `RU` [Как я пишу HTTP-сервисы спустя 13 лет работы с Go, Хабр](https://habr.com/ru/articles/793166/): перевод статьи Мэта Райера; структура сервера, хендлеры как функции, возвращающие `http.Handler`, middleware.
- `EN` [Routing Enhancements for Go 1.22](https://go.dev/blog/routing-enhancements): официальный пост про новый `ServeMux` и правило специфичности.
- `EN` [Better HTTP server routing in Go 1.22](https://eli.thegreenplace.net/2023/better-http-server-routing-in-go-122/): Эли Бендерски сравнивает со сторонними роутерами.
- `EN` [Making and Using HTTP Middleware](https://www.alexedwards.net/blog/making-and-using-middleware): Алекс Эдвардс про паттерн `func(http.Handler) http.Handler` и порядок цепочки.
- `EN` [Документация net/http](https://pkg.go.dev/net/http): разделы `ServeMux` (паттерны и специфичность), `ResponseWriter`, `ResponseController`.

## Домашка

<Hw id="ex1" title="Маршруты на ServeMux" level="база">

Файл `homework/m06/l02-net-http/ex1_routes.go`, функция `NewRouter`: зарегистрируй маршруты из таблицы в комментарии. Тесты проверяют приоритет точного пути над wildcard, `{path...}`, 404, 405 с `Allow` и то, что `/` не ловит всё подряд.

```bash
cd homework && go test -v ./m06/l02-net-http/ -run TestRouter
```

</Hw>

<Hw id="ex2" title="Цепочка и request ID" level="база">

Файл `ex2_middleware.go`: `Chain` (первый middleware самый внешний), `RequestID` (сквозной ID из заголовка или сгенерированный, валидация входящего, заголовок ответа до вызова `next`) и `RequestIDFrom`.

```bash
cd homework && go test -v ./m06/l02-net-http/ -run 'TestChain|TestRequestID'
```

</Hw>

<Hw id="ex3" title="Access-лог и ловушки ResponseWriter" level="посложнее">

Файл `ex3_accesslog.go`, функция `AccessLog`: обёртка над `ResponseWriter`, которая запоминает первый статус (и 200, если `WriteHeader` не вызывали), считает байты, пробрасывает `Unwrap` для `ResponseController` и достаёт `r.Pattern` после работы роутера.

```bash
cd homework && go test -v ./m06/l02-net-http/ -run TestAccessLog
```

</Hw>

<Hw id="ex4" title="Recovery" level="база">

Файл `ex4_recover.go`, функция `Recover`: превратить панику в 500 без утечки деталей, залогировать значение, а `http.ErrAbortHandler` пробросить дальше.

```bash
cd homework && go test -v ./m06/l02-net-http/ -run TestRecover
```

</Hw>

<Hw id="ex5" title="Собери сервер" level="посложнее">

Вне пакета домашки (в своём модуле) собери `main.go`: `NewRouter` из ex1, обёрнутый в `Chain(router, Recover(...), RequestID(...), AccessLog(...))`, логи через `log.Printf`. Запусти и проверь curl-ом: `curl -i localhost:8080/users/42`, `curl -i -X POST localhost:8080/users/42` (405 и `Allow`), `curl -i -H 'X-Request-Id: abc' localhost:8080/` (ID вернулся в ответе). Добавь маршрут, который паникует, и убедись, что клиент получил 500, а сервер жив. Поменяй порядок middleware так, чтобы recovery стал внутренним, и объясни, что сломалось.

</Hw>
