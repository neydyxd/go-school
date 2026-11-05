# Middleware

Middleware это код, который выполняется вокруг каждого обработчика: до него и после него. Логирование, перехват паник, request ID, авторизация, CORS, таймауты, ограничение тела запроса: всё это сквозная функциональность, которую не пишут в каждой ручке, а вешают один раз на роутер или группу.

В модуле про `net/http` middleware было функцией `func(http.Handler) http.Handler`, и порядок задавался вложенностью обёрток. В Gin модель другая: плоская цепочка обработчиков и два метода управления, `c.Next()` и `c.Abort()`. Она проще в записи, но у неё свои ловушки: middleware без `Abort` пропускает запрос дальше, логгер до `Next` видит неправильный статус, CORS на группе не ловит preflight, а таймаут «как в туториале» гоняет данные между горутинами.

В этом уроке разберём, как устроена цепочка, напишем набор боевых middleware на `slog` и обсудим, в каком порядке их ставить.

## Как устроена цепочка

При регистрации маршрута Gin склеивает все обработчики в один слайс: глобальные middleware (`r.Use`), middleware групп и сам обработчик. В `gin.Context` хранится этот слайс и индекс текущего обработчика.

```go
func (c *Context) Next() {
	c.index++
	for c.index < safeInt8(len(c.handlers)) {
		if c.handlers[c.index] != nil {
			c.handlers[c.index](c)
		}
		c.index++
	}
}

func (c *Context) Abort() {
	c.index = abortIndex // math.MaxInt8 >> 1, то есть 63
}
```

Это весь механизм. Отсюда следует всё поведение:

- **Middleware, вызвавшее `c.Next()`**, запускает остаток цепочки прямо внутри себя. Код до `Next` выполняется на входе, код после `Next` на выходе, когда обработчик уже отработал. Это и есть «луковица».
- **Middleware без `c.Next()`** просто возвращается, и цикл в `Next` уровнем выше переходит к следующему обработчику. Запрос не останавливается!
- **`c.Abort()`** выставляет индекс за пределы цепочки, и цикл заканчивается. Текущая функция при этом продолжает выполняться до `return`.
- В цепочке может быть не больше 62 обработчиков (`abortIndex` 63 зарезервирован), при превышении Gin паникует при регистрации: `too many handlers`.

```go
func mw(name string) gin.HandlerFunc {
	return func(c *gin.Context) {
		fmt.Println(name, "до")
		c.Next()
		fmt.Println(name, "после")
	}
}

r := gin.New()
r.Use(mw("A"), mw("B"))
r.GET("/x", func(c *gin.Context) { fmt.Println("handler") })
```

<Predict :options="['A до, B до, handler, A после, B после', 'A до, B до, handler, B после, A после', 'A до, A после, B до, B после, handler', 'handler, B после, A после']" :correct="1">

Каждый `Next` вызывает остаток цепочки внутри себя, поэтому выход идёт в обратном порядке: B завершается раньше A, как вложенные вызовы функций. Это свойство используют логгер и recovery: обёртка снаружи видит всё, что произошло внутри.

</Predict>

```go
r := gin.New()
r.Use(func(c *gin.Context) {
	fmt.Println("до")
	fmt.Println("после")
})
r.GET("/", func(c *gin.Context) { fmt.Println("обработчик") })
```

<Predict :options="['до, после, обработчик', 'до, обработчик, после', 'до, после (обработчик не вызван)', 'обработчик, до, после']" :correct="0">

Без `c.Next()` middleware отработало целиком и вернулось, а цикл в `Next` движка вызвал следующий обработчик. Значит, «после» тут на самом деле «до». И значит, забытый `Next` не блокирует запрос, а забытый `Abort` пропускает его дальше.

</Predict>

### Abort и возврат

```go
r.Use(func(c *gin.Context) {
	if !authorized(c) {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	c.Next()
})
```

Тела ответов в примерах и домашке этого урока упрощены до `{"error": "unauthorized"}`, чтобы не отвлекать от механики цепочки. В настоящем сервисе middleware отвечают в том же едином формате, что и обработчики (урок про binding): `{"error": {"code": "unauthorized", "message": "..."}}`, через общую функцию вроде `fail`.

`AbortWithStatusJSON` = `Abort` + `JSON`. `return` после него нужен, чтобы не выполнить код ниже в этой же функции (`Abort` его не остановит). Варианты: `AbortWithStatus(code)` без тела, `AbortWithError(code, err)` с записью ошибки в `c.Errors`.

::: trap Отказ без Abort
Плохо:

```go
func RequireToken(c *gin.Context) {
	if c.GetHeader("Authorization") == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
}
```

Клиент получает `401` и **следом** в том же теле ответ защищённой ручки: `{"error":"unauthorized"}{"secret":42}`. Статус 401 обманчиво выглядит правильным в логах, а данные утекли. Хорошо: `c.AbortWithStatusJSON(...)` и `return`. На это стоит написать тест, который проверяет, что обработчик не вызывался.
:::

### Где вешать middleware

```go
r := gin.New()
r.Use(RequestID(), Logger(log), Recovery(log)) // на всё, включая 404

api := r.Group("/api", RequireToken(tokens))   // на группу
api.GET("/me", me)

admin := r.Group("/admin")
admin.Use(RequireRole("admin"))                // то же самое, второй способ
admin.GET("/stats", stats)

r.GET("/report", RateLimit(5), report)         // на один маршрут
```

::: trap Use после маршрутов
Цепочка склеивается **в момент регистрации маршрута**. Маршруты, объявленные до `r.Use(mw)`, его не получат:

```go
r.GET("/before", h)
r.Use(func(c *gin.Context) { c.Header("X-Mw", "yes") })
r.GET("/after", h)
// /before: заголовка нет, /after: X-Mw: yes
```

Так легко оставить без авторизации или логирования пару ручек, объявленных «наверху для удобства». Хорошо: все `r.Use` и `Group(..., mw)` сразу после создания роутера, маршруты после.
:::

## Request ID

Request ID связывает все записи лога одного запроса и позволяет найти их по номеру из жалобы пользователя. Если перед сервисом стоит балансировщик или шлюз, он уже присылает ID в заголовке (часто `X-Request-ID`), и его нужно продолжить; иначе генерируем свой.

```go
const HeaderRequestID = "X-Request-ID"

type requestIDKey struct{}

var validID = regexp.MustCompile(`^[A-Za-z0-9_-]{1,64}$`)

func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader(HeaderRequestID)
		if !validID.MatchString(id) {
			b := make([]byte, 16)
			_, _ = rand.Read(b) // crypto/rand.Read с Go 1.24 никогда не возвращает ошибку
			id = hex.EncodeToString(b)
		}
		c.Request = c.Request.WithContext(context.WithValue(c.Request.Context(), requestIDKey{}, id))
		c.Header(HeaderRequestID, id)
		c.Next()
	}
}

func RequestIDFrom(ctx context.Context) string {
	id, _ := ctx.Value(requestIDKey{}).(string)
	return id
}
```

Три важные детали:

- **Входящий ID проверяем.** Заголовок присылает клиент, а мы пишем его в лог и в ответ. 10 КБ мусора или перевод строки в нём это мусор в логах и попытка подделать строки лога (log injection).
- **ID кладём в `context.Context` запроса**, а не только в `c.Set`. `c.Set` видно только тем, у кого есть `*gin.Context`, то есть HTTP-слою. Сервис и репозиторий получают `context.Context`, и через него ID дойдёт до логов базы.
- **Ключ контекста приватного типа**, чтобы чужой пакет не перетёр его строкой `"request_id"` (тот же приём, что в уроке про context).

## Логирование через slog

Встроенный `gin.Logger()` пишет текст в stdout. Для продакшна нужен структурный лог: JSON-строка на запрос, которую соберёт Loki, ELK или любой другой агрегатор, с request ID, шаблоном маршрута и длительностью.

```go
func Logger(log *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next() // всё самое интересное происходит здесь

		status := c.Writer.Status()
		level := slog.LevelInfo
		switch {
		case status >= 500:
			level = slog.LevelError
		case status >= 400:
			level = slog.LevelWarn
		}
		ctx := c.Request.Context()
		log.LogAttrs(ctx, level, "request",
			slog.String("method", c.Request.Method),
			slog.String("route", c.FullPath()), // "/users/:id", а не "/users/42"
			slog.String("path", c.Request.URL.Path),
			slog.Int("status", status),
			slog.Duration("duration", time.Since(start)),
			slog.String("request_id", RequestIDFrom(ctx)),
		)
	}
}
```

- Статус читаем **после** `c.Next()`. До него `c.Writer.Status()` возвращает 200 по умолчанию, потому что ответ ещё не записан.
- `c.FullPath()` даёт шаблон маршрута. По нему удобно группировать метрики и логи: у `/users/42` и `/users/43` один шаблон. Для 404 он пустой.
- `LogAttrs` с типизированными атрибутами быстрее, чем `log.Info("request", "status", status, ...)`: нет упаковки в `any` и разбора пар.

::: trap Логи с секретами
Не пиши в лог тело запроса, заголовок `Authorization`, cookie и полный query: в них пароли, токены и персональные данные, а доступ к логам обычно шире, чем к базе. Если очень нужно, логируй белый список полей или маскируй значения. По той же причине в лог идёт `c.Request.URL.Path`, а не `RequestURI` с query-строкой.
:::

## Recovery

Паника в обработчике без recovery завершит не весь процесс (`net/http` ловит паники в обработчиках, пишет стек в лог сервера и рвёт соединение), но клиент получит оборванный ответ, а ты не узнаешь, на каком запросе. Свой recovery отвечает 500 в едином формате и пишет в тот же структурный лог:

```go
func Recovery(log *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if v := recover(); v != nil {
				log.ErrorContext(c.Request.Context(), "panic",
					slog.String("panic", fmt.Sprint(v)),
					slog.String("stack", string(debug.Stack())))
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "internal"})
			}
		}()
		c.Next()
	}
}
```

`recover` работает только в той горутине, где случилась паника. Паника в горутине, которую запустил обработчик, этот middleware не поймает и уронит весь процесс: такие горутины защищай своим `defer recover()`.

### Порядок: кто снаружи

Порядок middleware в `r.Use` это порядок слоёв луковицы, первый снаружи. Рабочий порядок для типового API:

```go
r.Use(
	RequestID(),       // 1. ID нужен всем, кто ниже, и уходит в ответ
	Logger(log),       // 2. снаружи Recovery: увидит итоговый статус 500 после паники
	Recovery(log),     // 3. ловит панику всего, что ниже
	CORS(origins...),  // 4. preflight отвечается до авторизации
)
api := r.Group("/api", RequireToken(tokens)) // 5. авторизация только на API
```

::: hood Почему Logger обязан быть снаружи Recovery
Паника раскручивает стек вызовов. Если порядок `Recovery, Logger`, то паника из обработчика пролетает сквозь `Logger` (его код после `c.Next()` не выполняется, у него нет `defer`), и её ловит `Recovery` снаружи. Запись «request» не появится вовсе: в логе будет только «panic» без метода, маршрута и длительности.

При порядке `Logger, Recovery` паника останавливается на `Recovery`, тот пишет 500, возвращается в `Next` логгера, и логгер штатно пишет запись со статусом 500.

А вот `Logger` до `RequestID` сработает, хотя выглядит неправильно: `RequestID` присваивает новый запрос в `c.Request`, а логгер читает `c.Request.Context()` уже после `c.Next()`, когда замена произошла. Полагаться на это не стоит: любое чтение ID в логгере до `Next` (например, для лога «запрос начат») получит пустую строку.
:::

## CORS

Браузер не даёт JavaScript со страницы `https://app.example.com` читать ответы API на `https://api.example.com`, пока API явно не разрешит это заголовками CORS. Для «непростых» запросов (JSON-тело, заголовок `Authorization`, методы PUT/DELETE) браузер сначала шлёт **preflight**: `OPTIONS` с заголовками `Origin` и `Access-Control-Request-Method`, и только при правильном ответе отправляет настоящий запрос.

```go
func CORS(origins ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin == "" {
			c.Next()
			return
		}
		ok := slices.Contains(origins, origin)
		if ok {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Vary", "Origin")
		}
		if c.Request.Method == http.MethodOptions && c.GetHeader("Access-Control-Request-Method") != "" {
			if !ok {
				c.AbortWithStatus(http.StatusForbidden)
				return
			}
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE")
			c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type, X-Request-ID")
			c.Header("Access-Control-Max-Age", "600")
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
```

- Origin сравнивается **точно**. Проверка `strings.HasPrefix(origin, allowed)` пропустит `https://app.example.com.evil.example`.
- `Vary: Origin` обязателен, когда ответ зависит от Origin: иначе CDN или прокси закеширует ответ с заголовком одного сайта и отдаст его другому.
- `*` в `Access-Control-Allow-Origin` нельзя сочетать с `Access-Control-Allow-Credentials: true` (куки, HTTP-аутентификация): браузер отвергнет такой ответ.
- CORS это защита пользователей браузера, а не API. `curl` и сервер злоумышленника на CORS не смотрят, авторизация всё равно нужна.

В реальном проекте можно взять `github.com/gin-contrib/cors`, но понимать, что он делает, нужно всё равно.

::: trap CORS на группе
Плохо:

```go
api := r.Group("/api", RequireToken(tokens), CORS(origins...))
api.GET("/me", me)
```

Preflight `OPTIONS /api/me` не совпадает ни с одним маршрутом (зарегистрирован только GET), поэтому в цепочку группы он не попадает: Gin ответит 404, и браузер заблокирует запрос. Даже если бы попал, `RequireToken` стоит раньше и отвечает 401: браузер не шлёт `Authorization` в preflight.

Хорошо: `r.Use(CORS(origins...))` глобально. Для ненайденных маршрутов Gin собирает цепочку из глобальных middleware и `NoRoute`, так что preflight дойдёт до CORS и получит 204.
:::

## Таймауты и размер тела

### Почему timeout-middleware в Gin коварен

Хочется написать middleware «если обработчик работает дольше 2 секунд, ответить 503». Чтобы ответить, пока обработчик ещё работает, его придётся запустить **в другой горутине**. А дальше всё плохо:

- `*gin.Context` не потокобезопасен, а теперь его трогают две горутины: обработчик пишет ответ и вызывает `c.Next`, middleware пишет 503 и вызывает `Abort`. Популярный `gin-contrib/timeout` собрал несколько issues про data race, в том числе прямо на `c.Next()` и `c.Abort()`, и гонки в нём чинили не одним PR;
- обработчик после таймаута **продолжает работать**: держит соединение с базой, может записать в ответ (его нужно буферизовать и выбрасывать), может запаниковать мимо recovery;
- после возврата middleware контекст уходит в пул, а горутина обработчика всё ещё его использует (см. урок про Gin и `Copy`).

Надёжная схема другая: **дедлайн через `context.Context`**. Middleware кладёт в запрос контекст с таймаутом, обработчик, сервис и драйвер базы следят за `ctx.Done()` и сами бросают работу. Всё в одной горутине, гонок нет:

```go
func Timeout(d time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), d)
		defer cancel()
		c.Request = c.Request.WithContext(ctx)

		c.Next()

		if errors.Is(ctx.Err(), context.DeadlineExceeded) && !c.Writer.Written() {
			c.AbortWithStatusJSON(http.StatusGatewayTimeout, gin.H{"error": "timeout"})
		}
	}
}
```

Цена: таймаут кооперативный. Обработчик, который не смотрит на контекст (`time.Sleep`, CPU-цикл, вызов без `ctx`), отработает до конца. Поэтому весь код ниже HTTP-слоя принимает `ctx` первым аргументом и передаёт его в `pool.Query(ctx, ...)`. Жёсткий потолок ставят на уровне сервера: `http.Server{ReadTimeout, WriteTimeout}` или `http.TimeoutHandler` вокруг всего движка (он запускает обработчик в отдельной горутине, но весь Gin целиком живёт в ней, и общий контекст никто не делит).

### Ограничение тела

Без ограничения клиент пришлёт гигабайт JSON, и `ShouldBindJSON` попробует его прочитать в память.

```go
func BodyLimit(n int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.ContentLength > n {
			c.AbortWithStatusJSON(http.StatusRequestEntityTooLarge, gin.H{"error": "body too large"})
			return
		}
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, n)
		c.Next()
	}
}
```

Проверки `Content-Length` мало: при `Transfer-Encoding: chunked` длина неизвестна (`-1`), а клиент может и соврать. `http.MaxBytesReader` обрывает чтение после `n` байт с ошибкой `*http.MaxBytesError`. `c.BindJSON` сам превращает её в 413, а с `ShouldBindJSON` проверяй `errors.As(err, &mbe)` в своём обработчике ошибок binding.

## Что выведет программа?

```go
r := gin.New()
r.Use(func(c *gin.Context) {
	c.AbortWithStatusJSON(401, gin.H{"error": "unauthorized"})
	fmt.Println("после Abort")
}, func(c *gin.Context) {
	fmt.Println("второе middleware")
})
r.GET("/x", func(c *gin.Context) { fmt.Println("обработчик") })
```

<Predict :options="['ничего не напечатает', 'после Abort', 'после Abort, второе middleware, обработчик', 'второе middleware, обработчик']" :correct="1">

`Abort` не прерывает текущую функцию, поэтому строка «после Abort» печатается. Но индекс цепочки уже за её пределами, и ни второе middleware, ни обработчик не вызываются. Если после отказа в middleware есть код с побочными эффектами, нужен `return`.

</Predict>

```go
log := slog.New(slog.NewJSONHandler(os.Stdout, nil))
r := gin.New()
r.Use(func(c *gin.Context) {
	log.Info("request", "status", c.Writer.Status())
	c.Next()
})
r.GET("/missing", func(c *gin.Context) { c.JSON(404, gin.H{"error": "not found"}) })
// GET /missing
```

<Predict :options="['status 404', 'status 200', 'status 0', 'запись не появится']" :correct="1">

Статус прочитан до `c.Next()`, когда ответ ещё не записан: Gin по умолчанию держит 200. Логгер, который пишет до `Next`, врёт про каждый 4xx и 5xx. Статус и длительность читаются только после `Next`.

</Predict>

::: interview
<details><summary>Как в Gin работают c.Next и c.Abort?</summary>

Контекст хранит слайс обработчиков (глобальные middleware, групповые, сам обработчик) и индекс. `Next` увеличивает индекс и в цикле вызывает оставшиеся обработчики, поэтому код после `Next` в middleware выполняется на обратном пути. `Abort` ставит индекс в `abortIndex` (63), и цикл больше никого не вызывает. Текущая функция после `Abort` продолжает выполняться до `return`.

</details>
<details><summary>Что будет, если middleware не вызывает c.Next?</summary>

Запрос не остановится: middleware вернётся, и цикл `Next` уровнем выше вызовет следующий обработчик. `Next` нужен, только чтобы выполнить код после остальной цепочки (логгер, recovery, таймаут). Остановить цепочку можно только `Abort`.

</details>
<details><summary>В каком порядке ставить RequestID, Logger, Recovery, CORS и авторизацию?</summary>

RequestID первым, чтобы ID был у всех. Logger снаружи Recovery, иначе паника пролетит мимо записи лога и 500 не будет залогирован. Recovery до всего, что может паниковать. CORS глобально и до авторизации: preflight идёт без токена и на маршрут, которого нет (OPTIONS). Авторизация на группе закрытых маршрутов.

</details>
<details><summary>Как сделать таймаут запроса в Gin и чем плох запуск обработчика в горутине?</summary>

Через контекст: middleware кладёт в `c.Request` контекст с `WithTimeout`, отменяет его после `Next`, а весь код ниже передаёт `ctx` в базу и внешние вызовы. Запуск обработчика в отдельной горутине ради ответа 503 даёт гонки на `gin.Context` (он не потокобезопасен и возвращается в пул), обработчик продолжает работать после таймаута и может писать в ответ. Жёсткий предел ставят на сервере: `WriteTimeout` или `http.TimeoutHandler` вокруг движка.

</details>
<details><summary>Зачем request ID и как его протащить до репозитория?</summary>

Чтобы связать все записи лога одного запроса и найти их по номеру из ошибки. Берём из заголовка балансировщика (после проверки формата) или генерируем, кладём в `context.Context` запроса под ключом приватного типа и в заголовок ответа. Сервис и репозиторий получают `ctx` аргументом и достают ID оттуда; логгер может добавлять его сам через свой `slog.Handler`.

</details>
<details><summary>Как работает CORS и почему preflight часто ломается?</summary>

Браузер для непростых запросов шлёт `OPTIONS` с `Origin` и `Access-Control-Request-Method` и отправляет настоящий запрос, только если ответ разрешил origin, метод и заголовки. Ломается, когда CORS повешен на группу (OPTIONS-маршрута нет, ответ 404) или стоит после авторизации (preflight без токена получает 401). Ещё нужны точное сравнение origin и `Vary: Origin`.

</details>
<details><summary>Как ограничить размер тела запроса?</summary>

`http.MaxBytesReader` поверх `c.Request.Body`: чтение больше лимита вернёт `*http.MaxBytesError`, его переводят в 413. Дополнительно можно сразу отказать, если `Content-Length` больше лимита. Одной проверки `Content-Length` мало: при chunked он `-1`, и клиент может соврать.

</details>
<details><summary>Почему recovery не ловит панику в горутине, запущенной обработчиком?</summary>

`recover` работает только в той горутине, где произошла паника, и только в отложенной функции на её стеке. Горутина, запущенная обработчиком, живёт со своим стеком, и паника в ней без собственного `defer recover()` роняет весь процесс.

</details>
:::

## Источники

- `RU` [Пользовательский middleware](https://gin-gonic.com/ru/docs/middleware/custom-middleware/) и [Использование middleware](https://gin-gonic.com/ru/docs/middleware/using-middleware/): официальная документация Gin, `Next` и подключение на группы.
- `RU` [CORS на MDN](https://developer.mozilla.org/ru/docs/Web/HTTP/CORS): как браузер делает preflight и какие заголовки ждёт.
- `RU` [Middleware обработки ошибок](https://gin-gonic.com/ru/docs/middleware/error-handling-middleware/): как собрать ошибки из `c.Errors` в одном месте.
- `EN` [Structured Logging with slog](https://go.dev/blog/slog): блог Go про `log/slog`, `LogAttrs` и атрибуты.
- `EN` [Data race on c.Next() and c.Abort()](https://github.com/gin-contrib/timeout/issues/45): issue в `gin-contrib/timeout`, живой пример того, почему таймаут через горутину опасен.
- `EN` [net/http: TimeoutHandler и MaxBytesReader](https://pkg.go.dev/net/http): официальная документация, описание обеих функций.

## Домашка

<Hw id="ex1" title="Request ID" level="база">

Файл `homework/m08/l03-middleware/ex1_requestid.go`: `RequestID` и `RequestIDFrom`. Тест проверяет, что допустимый входящий ID сохраняется, мусорный заменяется на 32 hex-символа, ID уникальны, лежат в контексте запроса под приватным ключом и уходят в ответ.

```bash
cd homework && go test -v ./m08/l03-middleware/ -run TestRequestID
```

</Hw>

<Hw id="ex2" title="Авторизация по токену" level="база">

Файл `ex2_auth.go`: `RequireToken`. Тест считает вызовы обработчика: после 401 он не должен выполниться, в теле ровно один JSON, есть `WWW-Authenticate: Bearer`.

```bash
cd homework && go test -v ./m08/l03-middleware/ -run TestRequireToken
```

</Hw>

<Hw id="ex3" title="Logger и Recovery на slog" level="база">

Файл `ex3_logger.go`: запись `request` с методом, шаблоном маршрута, статусом, длительностью и request ID, уровень по статусу; recovery с ответом 500 в JSON. Тест разбирает JSON-лог: статус до `c.Next()` он поймает.

```bash
cd homework && go test -v ./m08/l03-middleware/ -run 'TestLogger|TestRecovery'
```

</Hw>

<Hw id="ex4" title="CORS" level="посложнее">

Файл `ex4_cors.go`: точное сравнение origin, `Vary: Origin`, preflight 204 с нужными заголовками для своих и 403 для чужих.

```bash
cd homework && go test -v ./m08/l03-middleware/ -run TestCORS
```

</Hw>

<Hw id="ex5" title="Таймаут и лимит тела" level="посложнее">

Файл `ex5_limits.go`: `Timeout` через контекст и `BodyLimit` через `MaxBytesReader`. Тест таймаута идёт в `synctest` (время фейковое и точное), проверяет 504, отмену контекста после ответа и то, что поверх уже записанного ответа 504 не дописывается. Тест лимита шлёт тело без `Content-Length`.

```bash
cd homework && go test -v -race ./m08/l03-middleware/ -run 'TestTimeout|TestBodyLimit'
```

</Hw>

<Hw id="ex6" title="Сборка сервиса" level="посложнее">

Файл `ex6_router.go`: собери всё в `NewRouter` в правильном порядке. Тесты проверяют, что паника попадает в лог со статусом 500 и тем же request ID, что ушёл клиенту, а preflight на `/api/me` проходит без токена.

```bash
cd homework && go test -v -race ./m08/l03-middleware/
```

</Hw>
