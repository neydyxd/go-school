# Gin: роутинг и первые шаги

После модуля про `net/http` у тебя в руках всё, чтобы написать HTTP-сервис без единой зависимости: с Go 1.22 стандартный `ServeMux` понимает методы и параметры пути, middleware это обычная функция-обёртка, JSON кодирует `encoding/json`. Зачем тогда Gin? Затем, что в реальном сервисе сотня обработчиков, и каждый из них разбирает JSON, валидирует поля, отвечает ошибкой в едином формате, читает параметры. Фреймворк убирает этот повторяющийся код и даёт общие соглашения для команды.

Gin самый популярный веб-фреймворк в Go, его спрашивают в вакансиях и на собесах. Но хороший бэкендер не просто «знает Gin», а понимает, что он делает поверх `net/http`, где у него острые углы (контекст из пула, `Bind` против `ShouldBind`, middleware после маршрутов) и когда он вообще не нужен.

В этом уроке: чем Gin отличается от chi, echo и fiber, как устроен его роутер, параметры пути и запроса, группы маршрутов, `gin.Context` и почему его нельзя просто так отдать горутине.

## Нужен ли фреймворк вообще

Честный ответ: для небольшого сервиса нет. Вот как выглядит маршрут с параметром на чистом stdlib с Go 1.22:

```go
mux := http.NewServeMux()
mux.HandleFunc("GET /users/{id}", func(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"id": id})
})
```

То же на Gin:

```go
r := gin.New()
r.GET("/users/:id", func(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"id": c.Param("id")})
})
```

Разница на одном маршруте мизерная. Она набегает на масштабе: привязка и валидация тела в одну строку (`c.ShouldBindJSON`), группы маршрутов со своими middleware, хранилище значений на время запроса (`c.Set`/`c.Get`), готовые рендеры (JSON, XML, файлы), экосистема middleware (`gin-contrib/cors`, `gin-contrib/timeout` и другие).

Когда фреймворк **не** нужен:

- Маленький сервис или CLI с парой эндпоинтов: stdlib короче, чем импорт фреймворка.
- Библиотека или SDK: не навязывай пользователям зависимость. Принимай и отдавай `http.Handler`.
- Команда пишет на stdlib и chi, и всем удобно. Gin не делает код быстрее в разы, он делает его короче.

Когда нужен: CRUD-сервис на десятки ручек, команда, которой нужен общий стиль, и проект, где Gin уже стоит (таких большинство в вакансиях).

## Gin, chi, echo, fiber: честное сравнение

| | Совместимость с `net/http` | Обработчик | Что даёт |
|---|---|---|---|
| **chi** | полная: маршрутизатор это `http.Handler`, обработчики `http.HandlerFunc` | `func(w, r)` | только роутер и набор middleware; всё остальное stdlib |
| **Gin** | движок реализует `http.Handler`, но обработчики свои | `func(c *gin.Context)` | роутер, binding, валидация, рендеры, middleware |
| **echo** | построен на `net/http`, есть обёртки для stdlib-обработчиков | `func(c *echo.Context) error` (v5) | как Gin, плюс обработчик возвращает `error`, ошибки собирает один `HTTPErrorHandler` |
| **fiber** | построен на **fasthttp**, а не на `net/http` | `func(c fiber.Ctx) error` | API в стиле Express из Node.js, упор на скорость |

Ключевое различие проходит не по удобству, а по фундаменту. chi, Gin и echo живут поверх `net/http`: работают с `http.Server`, `httptest`, `pprof`, OpenTelemetry-инструментацией и любым middleware из экосистемы. Fiber живёт поверх fasthttp, у которого свой тип запроса. `http.Handler` в fiber v3 можно повесить, но через слой совместимости: такие обработчики не видят `fiber.Ctx` и платят за конвертацию. Плюс fiber переиспользует значения из контекста между запросами: строку, полученную из `fiber.Ctx`, нельзя сохранить после выхода из обработчика, об этом прямо предупреждает README. Сами авторы fasthttp пишут, что он сделан для экстремальных нагрузок, а «в большинстве случаев `net/http` намного лучше».

::: trap Выбор фреймворка по бенчмаркам
«Fiber в 10 раз быстрее Gin» это про пустой обработчик, отдающий строку. В реальном запросе время уходит на базу, сеть и сериализацию, роутер занимает доли процента. Выбирай по совместимости и команде, а не по таблицам hello world.
:::

На собесе хороший ответ звучит так: «Для нового сервиса взял бы chi или Gin поверх `net/http`; fiber только если есть доказанная потребность в fasthttp и команда готова к его ограничениям».

## Первый сервис

```go
package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})
	r.Run(":8080") // внутри обычный http.ListenAndServe
}
```

- `gin.H` это просто `map[string]any`, короткое имя для ответов на лету.
- `r.Run` удобен для примеров, но в продакшне создают `http.Server` сам (таймауты, graceful shutdown из модуля про `net/http`) и передают `r` как `Handler`: `*gin.Engine` реализует `http.Handler`.

### gin.New против gin.Default

`gin.Default()` это `gin.New()` плюс два middleware: `Logger()` (текстовый лог каждого запроса в stdout) и `Recovery()` (перехват паники и ответ 500). В исходниках Gin это буквально три строки:

```go
func Default(opts ...OptionFunc) *Engine {
	debugPrintWARNINGDefault()
	engine := New()
	engine.Use(Logger(), Recovery())
	return engine.With(opts...)
}
```

В продакшн-сервисах обычно берут `gin.New()` и ставят свои middleware: структурный лог через `slog` с request ID, recovery, который пишет в тот же лог. Это тема урока про middleware. В тестах тоже `gin.New()`: лог каждого запроса там только шумит.

### Режимы debug, release, test

При старте Gin печатает `[GIN-debug] [WARNING] Running in "debug" mode` и список всех маршрутов. Режим задаётся переменной окружения `GIN_MODE` или кодом:

```go
gin.SetMode(gin.ReleaseMode) // "debug", "release", "test"
```

В release Gin не печатает отладочный вывод и предупреждения. Режим `test` выключает отладочный вывод в тестах. В коде `SetMode` есть попытка угадать test по флагу `test.v`, но при инициализации пакета этот флаг ещё не зарегистрирован, и `gin.Mode()` в `go test` возвращает `debug` (проверено на Gin 1.12). Поэтому в тестах режим ставят явно: `gin.SetMode(gin.TestMode)` в `init` или `TestMain`. В продакшне ставь `GIN_MODE=release` в окружении контейнера.

## Роутинг

### Методы и параметры пути

```go
r.GET("/users/:id", getUser)        // /users/42 -> c.Param("id") == "42"
r.POST("/users", createUser)
r.PUT("/users/:id", updateUser)
r.DELETE("/users/:id", deleteUser)
r.GET("/files/*path", serveFile)    // всё после /files
```

- `:id` совпадает ровно с одним сегментом пути, до следующего `/`. Значение всегда строка: перевести в число и проверить твоя задача.
- `*path` (catch-all) забирает весь остаток пути, **включая ведущий слеш**: для `/files/docs/a.txt` будет `/docs/a.txt`. Catch-all может стоять только в конце шаблона.

```go
r.GET("/files/*path", func(c *gin.Context) {
	c.String(http.StatusOK, c.Param("path"))
})
// GET /files/docs/a.txt
```

<Predict :options="['docs/a.txt', '/docs/a.txt', 'a.txt', '/files/docs/a.txt']" :correct="1">

Catch-all включает ведущий слеш. Поэтому `GET /files/` даёт `c.Param("path") == "/"`, а `GET /files` без слеша Gin по умолчанию перенаправляет на `/files/` кодом 301 (это работа `RedirectTrailingSlash`). Если тебе нужен путь без слеша, обрезай его сам: `strings.TrimPrefix(p, "/")`.

</Predict>

### Как устроен роутер

::: hood Radix-дерево из httprouter
Роутер Gin это переработанный код httprouter Джулиана Шмидта (в заголовке `tree.go` его копирайт). Для каждого HTTP-метода строится своё **radix-дерево** (сжатое префиксное дерево): общие префиксы маршрутов хранятся один раз.

```text
GET-дерево для /users, /users/:id, /users/:id/orders, /uploads
/u
├── sers
│   └── /
│       └── :id
│           └── /orders
└── ploads
```

Поиск идёт по символам пути вниз по дереву, без перебора всех маршрутов и без регулярных выражений, поэтому время зависит от длины пути, а не от числа маршрутов. Дочерние узлы упорядочены по приоритету (сколько маршрутов под ними), чтобы популярные ветки проверялись первыми. При совпадении в контекст пишутся параметры, а обработчики (уже склеенная цепочка middleware + handler) берутся прямо из узла.

Из устройства дерева следуют правила, которые проверяются при регистрации, а не при запросе:
- в одной позиции может быть только один параметр: `/users/:id` и `/users/:name` вместе дают панику `':name' in new path '/users/:name' conflicts with existing wildcard ':id'`;
- статический сегмент рядом с параметром можно: `/users/:id` и `/users/new` уживаются (в Gin до 1.7 это было конфликтом, поэтому в старых статьях ты встретишь обратное утверждение);
- рядом с catch-all ничего нельзя: `/files/*path` и `/files/new` дают панику;
- тот же метод и путь дважды: паника `handlers are already registered for path`.

Паника при старте это хорошо: ошибка в маршрутах не доживёт до продакшна.
:::

### Хвостовой слеш, 404 и 405

По умолчанию `RedirectTrailingSlash` включён: на `GET /users/42/` при маршруте `/users/:id` Gin ответит 301 с `Location: /users/42` (для не-GET методов 307). Обычно это то, что нужно.

Если на путь есть маршрут, но с другим методом (`POST /users/42` при одном `GET`), Gin по умолчанию отвечает **404**, а не 405. Чтобы получить честный 405 с заголовком `Allow`, включи флаг:

```go
r := gin.New()
r.HandleMethodNotAllowed = true
r.NoRoute(func(c *gin.Context) {
	c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
})
r.NoMethod(func(c *gin.Context) {
	c.JSON(http.StatusMethodNotAllowed, gin.H{"error": "method not allowed"})
})
```

Без `NoRoute`/`NoMethod` Gin отвечает текстом `404 page not found` с `Content-Type: text/plain`, что ломает клиента, который ждёт JSON на любой ответ.

### Параметры запроса

```go
// GET /search?q=go&limit=5
q := c.Query("q")                      // "" если нет
limit := c.DefaultQuery("limit", "10") // "10" если ключа нет
tags := c.QueryArray("tag")            // ?tag=a&tag=b -> [a b]
v, ok := c.GetQuery("page")            // отличить «нет» от «пусто»
```

Все значения строки. Числа переводи через `strconv` и проверяй диапазон. Привязку query к структуре с валидацией (`ShouldBindQuery`) разберём в следующем уроке.

Чем отличаются «ключа нет» и «ключ пустой», проверь задачей в конце урока.

## Группы маршрутов

Группа это общий префикс плюс общие middleware:

```go
r := gin.New()

api := r.Group("/api")
{
	v1 := api.Group("/v1")
	v1.GET("/users/:id", getUserV1)

	v2 := api.Group("/v2")
	v2.GET("/users/:id", getUserV2)
}

admin := r.Group("/admin", requireAdmin) // middleware только для группы
admin.GET("/stats", stats)
```

Фигурные скобки после `api := r.Group(...)` это обычный блок Go, он ничего не делает, кроме визуальной группировки. Так принято в примерах Gin, можно и без них.

Группы применяют для версий API (`/api/v1`, `/api/v2`), для разделения публичных и закрытых маршрутов (авторизация только на группе), для админки. Middleware, переданные в `Group`, склеиваются с глобальными в момент регистрации маршрута: у `/admin/stats` цепочка будет «глобальные middleware, потом `requireAdmin`, потом `stats`».

## gin.Context

`*gin.Context` это «всё про текущий запрос» в одном объекте:

- `c.Request` это исходный `*http.Request`, `c.Writer` это обёртка над `http.ResponseWriter`, которая помнит статус и размер ответа;
- параметры пути (`c.Params`), кеш query и формы;
- хранилище ключей на время запроса: `c.Set("user", u)`, `c.Get("user")`, `c.GetString("user")`;
- цепочка обработчиков и индекс текущего (`c.Next()`, `c.Abort()`, про них урок о middleware);
- накопленные ошибки `c.Errors`.

Он также реализует интерфейс `context.Context`, но с подвохом: по умолчанию его `Done()` возвращает `nil`, а `Value` не видит значения контекста запроса. В сервис и базу передавай `c.Request.Context()`, подробности в уроке про архитектуру.

### c.JSON не останавливает обработчик

`c.JSON(code, obj)` записывает ответ и **возвращается** в твой код. Это обычный вызов функции, а не `return`:

```go
r.GET("/err", func(c *gin.Context) {
	c.JSON(400, gin.H{"error": "bad"})
	c.JSON(200, gin.H{"ok": true})
})
```

<Predict :options="['200, в теле только ok', '400, в теле только error', '400, в теле оба JSON-объекта подряд', 'паника: заголовки уже записаны']" :correct="2">

Первый `c.JSON` записал статус 400 и тело. Второй не может сменить статус (заголовки ушли), Gin в debug-режиме печатает `[WARNING] Headers were already written. Wanted to override status code 400 with 200`, но **дописывает** второй JSON в тело: клиент получит `{"error":"bad"}{"ok":true}` и не сможет его разобрать. После ответа с ошибкой всегда `return`.

</Predict>

### c.JSON против c.AbortWithStatusJSON

`c.AbortWithStatusJSON(code, obj)` это `c.Abort()` плюс `c.JSON(code, obj)`. `Abort` выставляет индекс цепочки так, что **следующие** обработчики не вызовутся. Текущий при этом продолжает выполняться до `return`.

- В обычном обработчике (последнем в цепочке) разницы нет, после него и так никого.
- В middleware и в «проверках» внутри групп нужен именно `Abort`: иначе после ответа 403 выполнится защищённый обработчик и допишет свои данные.

```go
admin := r.Group("/admin", func(c *gin.Context) {
	if c.GetHeader("X-Admin-Token") != token {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}
	c.Next()
})
```

::: trap Проверка без Abort
Плохо:

```go
func requireAdmin(c *gin.Context) {
	if c.GetHeader("X-Admin-Token") != token {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return // выходим из middleware, но цепочка идёт дальше!
	}
}
```

Клиент получит 403 и **следом** секретную статистику в том же теле. `return` выходит только из текущей функции, а Gin после неё вызывает следующий обработчик. Хорошо: `c.AbortWithStatusJSON(...)` и `return`.
:::

### Контекст из пула и горутины

::: hood Откуда берётся gin.Context
`Engine.ServeHTTP` не создаёт контекст на каждый запрос, а берёт его из `sync.Pool`:

```go
func (engine *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// ... (сокращено: однократная подготовка деревьев маршрутов)
	c := engine.pool.Get().(*Context)
	c.writermem.reset(w)
	c.Request = req
	c.reset()
	engine.handleHTTPRequest(c)
	engine.pool.Put(c)
}
```

Так Gin экономит аллокации: под нагрузкой одни и те же объекты контекста ходят по кругу. Цена: как только обработчик вернул управление, его `*gin.Context` возвращается в пул, и **следующий запрос** перезапишет в нём параметры, ключи, `Request` и `Writer`.
:::

Отсюда главное правило: горутина, которая живёт дольше обработчика, не должна трогать исходный `c`. Для этого есть `c.Copy()`: он делает копию с собственными параметрами и ключами (`Keys` клонируются под мьютексом), а писать ответ через копию нельзя, у неё отвязан `ResponseWriter`.

::: trap Контекст в горутине
Плохо:

```go
r.POST("/reports/:id", func(c *gin.Context) {
	go func() {
		buildReport()                               // долго
		send(c.GetString("user"), c.Param("id"))    // c уже чужой!
	}()
	c.JSON(http.StatusAccepted, gin.H{"status": "accepted"})
})
```

Пока строится отчёт, контекст ушёл другому запросу: отчёт улетит не тому пользователю, а детектор гонок (`go test -race`) покажет data race. Плохо и так: `go func() { cp := c.Copy(); ... }()`, копия делается слишком поздно, уже параллельно с чужим запросом.

Хорошо: копия или нужные значения **до** `go`:

```go
r.POST("/reports/:id", func(c *gin.Context) {
	cp := c.Copy()
	go func() {
		buildReport()
		send(cp.GetString("user"), cp.Param("id"))
	}()
	c.JSON(http.StatusAccepted, gin.H{"status": "accepted"})
})
```

Ещё проще передать в горутину только значения: `user, id := c.GetString("user"), c.Param("id")`. А если горутине нужен `context.Context` для базы, бери `context.WithoutCancel(c.Request.Context())` со своим таймаутом (урок про контекст в модуле конкурентности): контекст запроса отменится, как только клиент получит ответ.
:::

## Что выведет программа?

```go
r := gin.New()
r.GET("/users/:id", func(c *gin.Context) { c.String(200, "id="+c.Param("id")) })
r.GET("/users/new", func(c *gin.Context) { c.String(200, "new") })

w := httptest.NewRecorder()
r.ServeHTTP(w, httptest.NewRequest("GET", "/users/new", nil))
fmt.Println(w.Code, w.Body.String())
```

<Predict :options="['200 id=new', '200 new', 'паника при регистрации второго маршрута', '404 page not found']" :correct="1">

В Gin 1.12 статический сегмент и параметр на одном уровне уживаются, и статический побеждает: `/users/new` уходит в свой обработчик, `/users/42` в обработчик с параметром. В Gin до 1.7 здесь была бы паника о конфликте маршрутов. А вот два разных параметра (`:id` и `:name`) в одной позиции по-прежнему дают панику.

</Predict>

```go
r := gin.New()
r.GET("/users/:id", func(c *gin.Context) { c.String(200, "ok") })

w := httptest.NewRecorder()
r.ServeHTTP(w, httptest.NewRequest("POST", "/users/42", nil))
fmt.Println(w.Code, w.Body.String())
```

<Predict :options="['405 405 method not allowed', '404 404 page not found', '200 ok', '400 bad request']" :correct="1">

По умолчанию `HandleMethodNotAllowed = false`, и Gin не проверяет другие методы: путь не найден в POST-дереве, значит 404 с текстовым телом. Для 405 и заголовка `Allow: GET` включи `r.HandleMethodNotAllowed = true`.

</Predict>

```go
// GET /search?page=
_, ok := c.GetQuery("page")
fmt.Printf("[%s] %v\n", c.DefaultQuery("page", "1"), ok)
```

<Predict :options="['[1] false', '[] true', '[1] true', '[] false']" :correct="1">

`DefaultQuery` подставляет значение по умолчанию, только если ключа **нет совсем**. Ключ `page=` с пустым значением есть, поэтому вернётся пустая строка, и `strconv.Atoi("")` упадёт. Надёжный приём: брать значение, и если оно пустое, тоже подставлять дефолт, или привязывать query к структуре с тегом `default`.

</Predict>

::: interview
<details><summary>Зачем Gin, если в Go 1.22 ServeMux научился методам и параметрам?</summary>

ServeMux закрыл главную причину брать сторонний роутер, но Gin даёт больше: привязку и валидацию тела и query к структурам (`ShouldBindJSON` с тегами validator), группы маршрутов с middleware, хранилище значений на время запроса, рендеры, готовые middleware. На маленьком сервисе выгоды нет, на CRUD из десятков ручек меньше повторяющегося кода и единый стиль. Важно, что Gin остаётся поверх `net/http`: движок это `http.Handler`.

</details>
<details><summary>Чем Gin отличается от chi и от fiber?</summary>

chi это только роутер и middleware, обработчики остаются `http.HandlerFunc`, полная совместимость со stdlib. Gin поверх `net/http`, но со своим типом обработчика `func(*gin.Context)` и встроенными binding, валидацией, рендерами. Fiber построен на fasthttp, а не на `net/http`: другой тип запроса, `http.Handler` подключаются через слой совместимости, значения контекста переиспользуются между запросами и их нельзя хранить после обработчика. Fiber оправдан при доказанной потребности в fasthttp.

</details>
<details><summary>Чем gin.New отличается от gin.Default?</summary>

`Default` это `New` плюс `Logger()` и `Recovery()`. В продакшне обычно берут `New` и ставят свои middleware: структурный лог с request ID, recovery с записью в тот же лог. В тестах `New`, чтобы не шуметь.

</details>
<details><summary>Как устроен роутер Gin и почему поиск маршрута быстрый?</summary>

Это radix-дерево из httprouter, отдельное на каждый HTTP-метод. Общие префиксы хранятся один раз, поиск идёт по символам пути вниз по дереву без перебора маршрутов и регулярок. Дочерние узлы отсортированы по приоритету. Конфликты (два разных параметра в одной позиции, что-то рядом с catch-all) ловятся паникой при регистрации.

</details>
<details><summary>Чем :param отличается от *param?</summary>

`:id` совпадает с одним сегментом до следующего `/`. `*path` забирает весь остаток пути, включая ведущий слеш, и может быть только в конце шаблона. Рядом с catch-all на том же уровне нельзя регистрировать другие маршруты.

</details>
<details><summary>Почему gin.Context нельзя передавать в горутину и что делает Copy?</summary>

Контексты берутся из `sync.Pool` и после возврата обработчика уходят следующему запросу: параметры, ключи, `Request` и `Writer` перезаписываются. Горутина, читающая `c` позже, получит чужие данные и гонку. `c.Copy()` делает копию со своими `Params` и `Keys` и отвязанным `ResponseWriter`. Копию нужно сделать до запуска горутины.

</details>
<details><summary>Чем c.JSON отличается от c.AbortWithStatusJSON?</summary>

`c.JSON` только пишет ответ. `AbortWithStatusJSON` ещё вызывает `Abort()`, после чего следующие обработчики в цепочке не выполняются. Текущий обработчик в обоих случаях продолжает работать до `return`. В middleware для отказа нужен именно Abort, иначе защищённый обработчик выполнится и допишет ответ.

</details>
<details><summary>Что вернёт Gin на запрос с правильным путём, но неверным методом?</summary>

По умолчанию 404, потому что `HandleMethodNotAllowed` выключен. Если включить, Gin вернёт 405 с заголовком `Allow`, в котором перечислены методы, зарегистрированные на этот путь. Тело для обоих случаев настраивается через `NoRoute` и `NoMethod`.

</details>
:::

## Источники

- `RU` [Документация Gin на русском](https://gin-gonic.com/ru/docs/): официальная, разделы «Маршрутизация» и «Middleware» по теме урока.
- `RU` [Группировка маршрутов](https://gin-gonic.com/ru/docs/routing/grouping-routes/) и [Параметры в пути](https://gin-gonic.com/ru/docs/routing/param-in-path/): короткие официальные примеры.
- `RU` [Горутины внутри middleware](https://gin-gonic.com/ru/docs/middleware/goroutines-inside-a-middleware/): официальный пример с `c.Copy()`.
- `RU` [Разработка REST-серверов на Go. Часть 3: Gin](https://habr.com/ru/articles/562878/): перевод серии Эли Бендерски, тот же сервис на stdlib и на Gin, хорошо видно, что даёт фреймворк.
- `EN` [Routing Enhancements for Go 1.22](https://go.dev/blog/routing-enhancements): что умеет новый ServeMux и почему сторонний роутер стал нужен реже.
- `EN` [go-chi/chi](https://github.com/go-chi/chi) и [gofiber/fiber](https://github.com/gofiber/fiber): README обоих, у fiber прочитай раздел Zero Allocation.
- `EN` [fasthttp: fasthttp might not be for you](https://github.com/valyala/fasthttp): авторы fasthttp о том, когда `net/http` лучше.
- `EN` [Пакет gin на pkg.go.dev](https://pkg.go.dev/github.com/gin-gonic/gin): API `Engine`, `RouterGroup`, `Context`.

## Домашка

<Hw id="ex1" title="Маршруты, параметры, query" level="база">

Файл `homework/m08/l01-gin-intro/ex1_routes.go`, функция `NewRouter`: `/ping`, `/users/:id` с проверкой числа, `/files/*path` без ведущего слеша, `/search` с обязательным `q` и `limit` в диапазоне. Тест проверяет, что на ошибку в теле ровно один JSON-объект: забытый `return` после ответа с ошибкой он поймает.

```bash
cd homework && go test -v ./m08/l01-gin-intro/ -run TestNewRouter
```

</Hw>

<Hw id="ex2" title="Группы, 404 и 405" level="база">

Файл `ex2_groups.go`, функция `NewAPI`: версии `/api/v1` и `/api/v2` с разной формой ответа, группа `/admin` с проверкой токена прямо в `Group`, JSON-ответы на 404 и 405 с заголовком `Allow`. Тест поймает проверку без `Abort`: после 403 в теле не должно оказаться статистики.

```bash
cd homework && go test -v ./m08/l01-gin-intro/ -run TestNewAPI
```

</Hw>

<Hw id="ex3" title="Отчёт в фоне" level="посложнее">

Файл `ex3_async.go`, функция `ReportHandler`: ответить 202 сразу, а отчёт собрать в горутине. Тест после возврата обработчика подменяет параметры и ключи контекста, как это делает пул Gin, и проверяет, что в `send` ушли исходные `user` и `id`. Запускай с детектором гонок:

```bash
cd homework && go test -v -race ./m08/l01-gin-intro/ -run TestReport
```

</Hw>

<Hw id="ex4" title="Тот же сервис на stdlib" level="со звёздочкой">

Вне репозитория курса перепиши `NewRouter` из первого задания на `http.ServeMux` (Go 1.22: `"GET /users/{id}"`, `r.PathValue`, `{path...}` для остатка пути). Прогони те же запросы через `httptest` и сравни: сколько строк вышло, что пришлось писать руками (JSON-ответы, 405 в ServeMux есть из коробки?), что оказалось проще. Запиши вывод в двух-трёх предложениях: пригодится на собесе в ответе «зачем вам Gin».

</Hw>
