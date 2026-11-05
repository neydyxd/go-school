# Конфиг, логи, OpenAPI и версионирование

Сервис, который работает у тебя на ноутбуке, и сервис, который можно отдать в эксплуатацию, отличаются не бизнес-логикой, а «обвязкой». Откуда он берёт адрес базы и секрет для JWT? Что он пишет в лог, когда в три часа ночи что-то сломалось, и можно ли по этим логам найти один конкретный запрос? Как фронтенд узнаёт, какие поля придут в ответе, и что будет с мобильным приложением, если ты переименуешь `title` в `name`?

Эти вопросы обычно не задают на первом собесе в лоб, но они всплывают в каждом обсуждении «расскажи про свой проект». Джун, который отвечает «конфиг из переменных окружения с валидацией при старте, логи в JSON через slog с request_id, контракт в OpenAPI, изменения только совместимые», звучит как человек, который уже работал в команде.

В уроке три части: конфигурация по 12-factor, структурное логирование через `log/slog` и документирование API через OpenAPI вместе с версионированием.

## Конфигурация

### 12-factor: конфиг в окружении

Методология The Twelve-Factor App (её написали в Heroku, и она стала стандартом для облачных сервисов) говорит: всё, что отличается между окружениями (dev, staging, prod), живёт **в переменных окружения**, а не в коде и не в файлах внутри репозитория. Это адреса баз, секреты, порты, уровни логов. Проверка простая: можно ли прямо сейчас выложить репозиторий в открытый доступ, не скомпрометировав ни одного пароля?

Почему окружение, а не `config.yaml`? Переменные окружения понимают все: Docker (`-e`, `env_file`), Kubernetes (`env`, `envFrom`), systemd, CI. Один и тот же образ запускается в разных окружениях с разным конфигом, его не нужно пересобирать.

### Источники и приоритет

Обычный порядок: **флаг командной строки** важнее **переменной окружения**, а та важнее **значения по умолчанию**. Флаги удобны для локальных экспериментов (`go run ./cmd/api -port 9090`), окружение для деплоя.

```go
func Load(args []string, getenv func(string) string) (Config, error) {
	fs := flag.NewFlagSet("api", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	port := fs.String("port", "", "порт HTTP")
	if err := fs.Parse(args); err != nil {
		return Config{}, err
	}
	p := *port
	if p == "" {
		p = getenv("PORT")
	}
	if p == "" {
		p = "8080"
	}
	...
}

// в main:
cfg, err := config.Load(os.Args[1:], os.Getenv)
```

Две детали, которые делают функцию тестируемой. Свой `flag.NewFlagSet`, а не глобальный `flag.CommandLine`: глобальный нельзя разобрать дважды в одном процессе тестов, и при ошибке он вызывает `os.Exit`. И окружение приходит параметром `getenv`, поэтому тест передаёт обычную функцию над map и не трогает настоящее окружение процесса.

Отличить «переменная не задана» от «задана пустой» можно только через `os.LookupEnv`:

```go
os.Setenv("EMPTY", "")
v1 := os.Getenv("EMPTY")
v2 := os.Getenv("MISSING_VAR")
_, ok1 := os.LookupEnv("EMPTY")
_, ok2 := os.LookupEnv("MISSING_VAR")
fmt.Println(v1 == v2, ok1, ok2)
```

<Predict :options="['true true false', 'false true false', 'true false false', 'true true true']" :correct="0">

`Getenv` для незаданной переменной возвращает пустую строку, поэтому обе строки равны. `LookupEnv` различает случаи: `EMPTY` задана (пусть и пустая), `MISSING_VAR` нет. Обычно пустое значение трактуют как «не задано» и берут значение по умолчанию, но если пустая строка допустимое значение, нужен `LookupEnv`.

</Predict>

Для больших конфигов есть библиотеки: `caarlos0/env` и `kelseyhightower/envconfig` заполняют структуру по тегам (`env:"PORT" envDefault:"8080"`), `spf13/viper` умеет файлы, окружение, флаги и горячую перезагрузку. Для сервиса на десяток параметров хватает stdlib, и его поведение очевидно без чтения чужой документации.

### Валидация при старте

Конфиг проверяют **сразу при запуске**, до открытия порта и подключения к базе. Сервис, который стартовал с пустым `JWT_SECRET` и упал на первом логине через час, хуже сервиса, который не запустился вовсе: во втором случае деплой откатится автоматически, а в первом ты узнаешь о проблеме от пользователей. Это принцип **fail fast**.

Хорошая валидация сообщает **все** проблемы сразу, а не по одной за перезапуск. Для этого есть `errors.Join` (Go 1.20):

```go
var errs []error
if cfg.DatabaseURL == "" {
	errs = append(errs, errors.New("DATABASE_URL: обязательная переменная не задана"))
}
if len(cfg.JWTSecret) < 32 {
	errs = append(errs, errors.New("JWT_SECRET: want не короче 32 байт"))
}
if p < 1 || p > 65535 {
	errs = append(errs, fmt.Errorf("PORT: want 1-65535, got %q", port))
}
return cfg, errors.Join(errs...) // nil, если ошибок нет; иначе ошибки через перевод строки
```

В ошибке имя переменной, ожидание и полученное значение. Кроме секретов: значение `JWT_SECRET` или адрес базы с паролем в текст ошибки не попадают, ведь ошибка уйдёт в лог.

### Секреты

- Секреты не коммитят. В репозитории лежит `.env.example` с фиктивными значениями, настоящий `.env` в `.gitignore`.
- Переменные окружения не идеальное хранилище секретов: их видно в `docker inspect`, в `/proc/<pid>/environ` для того же пользователя, они наследуются дочерними процессами. В серьёзных системах секреты монтируют файлами (Kubernetes Secrets, Docker secrets) или берут из менеджера секретов (HashiCorp Vault, облачные KMS). Для учебного проекта окружение нормально, но на собесе стоит знать о компромиссе.
- Секрет не должен случайно попасть в лог. Самый частый путь: «залогирую конфиг при старте».

Как именно это происходит, разберём в задаче в конце урока, а защита такая: метод `String()` у `Config` для `fmt` и метод `LogValue()` для slog, оба отдают версию со скрытыми секретами. `LogValue` покажем ниже.

::: trap os.Getenv по всему коду
```go
// плохо: конфиг размазан по пакетам, ошибки находятся в рантайме
func (r *Repo) connect() {
	pool, _ := pgxpool.New(ctx, os.Getenv("DATABASE_URL"))
}
func sign(t Token) string {
	key := os.Getenv("JWT_SECRET") // а если пусто? подпишем пустым ключом
}
```

```go
// хорошо: один Load в main, валидация при старте, дальше значения передаются явно
cfg, err := config.Load(os.Args[1:], os.Getenv)
if err != nil {
	log.Fatalf("конфиг:\n%v", err)
}
tokens := auth.NewTokens([]byte(cfg.JWTSecret), cfg.AccessTTL)
```

Пакеты получают конкретные значения через конструкторы. Их можно протестировать без манипуляций с окружением, и сразу видно, от каких настроек зависит компонент.
:::

## Логирование со slog

### Зачем структурные логи

Строка `user 42 logged in from 1.2.3.4` хороша для человека и плоха для машины. Когда у тебя миллион строк в час, логи читают не глазами, а запросами в Loki, Elasticsearch или облачном логировании: «все ошибки с `user_id=42` за последний час», «все строки с `request_id=abc`». Для этого каждая запись должна быть набором полей, обычно JSON:

```json
{"time":"2026-11-05T10:00:00Z","level":"INFO","msg":"вход","user_id":42,"ip":"1.2.3.4"}
```

С Go 1.21 структурное логирование есть в стандартной библиотеке: пакет `log/slog`. До него использовали `zap` и `zerolog`, их ты встретишь в существующих проектах; идеи те же.

### Logger, Handler, уровни

`slog.Logger` это то, что ты вызываешь (`Info`, `Error`). `slog.Handler` решает, куда и в каком формате писать. В стандартной библиотеке два готовых: `TextHandler` (`key=value`) и `JSONHandler`.

```go
log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
	Level: slog.LevelInfo, // Debug ниже порога отбрасывается
}))
slog.SetDefault(log) // теперь и slog.Info, и старый log.Printf идут через него

log.Info("сервер запущен", "addr", ":8080")
log.Error("не удалось сохранить задачу", "task_id", 7, "err", err)
```

Уровни: `Debug` (-4), `Info` (0), `Warn` (4), `Error` (8). Числа выбраны с зазорами, чтобы можно было вставить свои уровни. В проде обычно `Info`, при разборе инцидента временно `Debug`. Чтобы менять уровень без перезапуска, вместо константы передают `*slog.LevelVar` и меняют его, например, из админского эндпоинта.

Логи пишут в stdout, а не в файл: по 12-factor сбор и ротация логов это задача окружения (Docker, Kubernetes), а не приложения.

### Атрибуты, With и группы

Пары после сообщения превращаются в атрибуты. Есть два стиля:

```go
log.Info("вход", "user_id", 42, "ip", ip)                    // удобно, но нечётное число аргументов даст !BADKEY
log.LogAttrs(ctx, slog.LevelInfo, "вход",
	slog.Int("user_id", 42), slog.String("ip", ip))           // типизировано и без лишних аллокаций
```

`go vet` проверяет вызовы slog и ругается на непарные ключи и значения (`call to slog.Info missing a final value`), это ещё одна причина не выключать vet.

`With` создаёт логгер с постоянными атрибутами, `WithGroup` и `slog.Group` вкладывают атрибуты в подобъект:

```go
reqLog := log.With("request_id", id, "user_id", uid)
reqLog.Info("задача создана", slog.Group("task", "id", 7, "project_id", 3))
// {"level":"INFO","msg":"задача создана","request_id":"...","user_id":42,"task":{"id":7,"project_id":3}}
```

```go
h := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
	ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
		if a.Key == slog.TimeKey {
			return slog.Attr{} // убираем время, чтобы вывод был стабильным
		}
		return a
	},
})
log := slog.New(h).With("user_id", 1)
log.Debug("кеш промахнулся")
log.Info("вход", "user_id", 2)
```

<Predict :options="['одна строка INFO, где только user_id 2', 'две строки: DEBUG и INFO', 'одна строка INFO, где user_id встречается дважды: 1 и 2', 'паника: дублирующийся ключ']" :correct="2">

Уровень по умолчанию `Info`, поэтому `Debug` отброшен. А дальше сюрприз: `JSONHandler` не схлопывает одинаковые ключи. Строка будет `{"level":"INFO","msg":"вход","user_id":1,"user_id":2}`. Формально это допустимый JSON, но разные системы сбора логов прочитают из него разное значение. Поэтому имена атрибутов держат уникальными: `user_id` у логгера запроса и, например, `target_user_id` у записи.

</Predict>

::: hood Как запись проходит через slog
```text
log.Info("msg", "k", v)
  -> Handler.Enabled(ctx, LevelInfo)?  нет: выходим, запись даже не собирается
  -> собираем Record{Time, Level, Message, PC} и добавляем атрибуты
  -> Handler.Handle(ctx, record)
       JSONHandler: время, уровень, сообщение,
       затем атрибуты от With (уже отформатированы заранее),
       затем атрибуты записи; для каждого вызывается ReplaceAttr
  -> один Write в io.Writer (запись строки атомарна относительно других логов этого хендлера)
```

`With` у встроенных хендлеров форматирует атрибуты **один раз** при создании логгера, а не при каждой записи. Поэтому логгер запроса с `request_id` создают один раз в middleware и передают дальше. Если значение дорого вычислять, оборачивают его в тип с методом `LogValue()`: slog вызовет его, только если запись действительно пишется.
:::

### Логгер запроса через контекст

Главное, что нужно при разборе инцидента: собрать все строки одного запроса. Для этого у каждого запроса есть **request_id** (из заголовка `X-Request-ID`, если его поставил балансировщик, иначе сгенерированный), и он должен быть в каждой строке, которую пишет любой слой. Тащить логгер параметром через все функции неудобно, поэтому его кладут в `context.Context`:

```go
type loggerKey struct{}

func WithLogger(ctx context.Context, l *slog.Logger) context.Context {
	return context.WithValue(ctx, loggerKey{}, l)
}

func FromContext(ctx context.Context) *slog.Logger {
	if l, ok := ctx.Value(loggerKey{}).(*slog.Logger); ok {
		return l
	}
	return slog.Default() // никогда не nil
}

func RequestLogger(base *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		id := c.GetHeader("X-Request-ID")
		if !validID(id) { // клиент может прислать что угодно, включая перевод строки
			id = newID()
		}
		c.Header("X-Request-ID", id)
		l := base.With("request_id", id)
		c.Request = c.Request.WithContext(WithLogger(c.Request.Context(), l))

		c.Next()

		l.Info("http request",
			"method", c.Request.Method,
			"path", c.Request.URL.Path, // без query
			"status", c.Writer.Status(),
			"duration_ms", time.Since(start).Milliseconds())
	}
}

// в сервисе, двумя слоями ниже:
func (s *TaskService) Create(ctx context.Context, ...) {
	logx.FromContext(ctx).Info("создаю задачу", "project_id", projectID)
}
```

Ключ контекста это собственный неэкспортируемый тип `loggerKey struct{}`, а не строка: так никакой другой пакет не перезапишет значение случайно. Входящий `X-Request-ID` проверяют: длинная строка или перевод строки в логе это инъекция в логи и способ испортить поиск.

Про контекст как способ передачи значений запроса подробно было в [уроке про context](/m04/l05-context); правило то же: в контекст кладут данные **запроса** (логгер, request_id, пользователь), а не зависимости вроде пула базы.

### Что не логировать

- **Секреты и токены**: пароли, `Authorization`, cookies, refresh-токены, API-ключи, `JWT_SECRET`. Лог читает больше людей, чем базу, и хранится дольше.
- **Персональные данные (PII)** сверх необходимого: email, телефон, паспорт, адрес. В России это 152-ФЗ, в Европе GDPR. Если без идентификатора никак, логируй `user_id`, а не email.
- **Query-строку целиком**: в ней бывают токены (`?token=...` в ссылках сброса пароля).
- **Тела запросов и ответов** целиком: в них пароли и персональные данные.

Защита в два слоя. Первый: дисциплина, логируешь конкретные поля, а не «весь запрос». Второй: страховка в хендлере через `ReplaceAttr`, который вычищает значения по имени ключа, и `LogValuer` на типах с секретами:

```go
func (c Config) LogValue() slog.Value {
	u, _ := url.Parse(c.DatabaseURL)
	return slog.GroupValue(
		slog.Int("port", c.Port),
		slog.String("database_url", u.Redacted()), // postgres://app:xxxxx@db:5432/tasks
		slog.String("jwt_secret", "[REDACTED]"),
	)
}

log.Info("старт", "config", cfg) // slog сам вызовет LogValue
```

::: trap Логирование токена «для отладки»
```go
// плохо: Authorization и cookie уйдут в лог целиком
dump, _ := httputil.DumpRequest(c.Request, true)
log.Debug("запрос", "dump", string(dump))

// плохо: query с токеном сброса пароля
log.Info("request", "url", c.Request.URL.String())
```

```go
// хорошо: только нужные поля, path без query
log.Info("request", "method", c.Request.Method, "path", c.Request.URL.Path)
```

«Это же только на Debug» не спасает: уровень включат при разборе инцидента в проде, и токены окажутся в системе логов, у которой своя политика доступа и хранения.
:::

::: trap Залогировать и вернуть
```go
// плохо: одна ошибка в логе трижды, на каждом слое
if err != nil {
	log.Error("не удалось получить задачу", "err", err)
	return err
}
```

Ошибку либо **обрабатывают** (и тогда логируют), либо **возвращают с контекстом** (`fmt.Errorf("получить задачу %d: %w", id, err)`), но не то и другое. Логирует её тот, кто принимает решение: обычно HTTP-слой, когда превращает ошибку в 500.
:::

## OpenAPI

### Зачем контракт

OpenAPI (бывший Swagger) это машиночитаемое описание HTTP API в YAML или JSON: пути, методы, параметры, тела запросов и ответов, коды ошибок, схема аутентификации. Ты, скорее всего, видел его со стороны фронтенда как Swagger UI или генерировал по нему TypeScript-типы через `openapi-typescript`. Что он даёт:

- **Документацию**, которая не врёт, если генерируется или проверяется автоматически.
- **Генерацию клиентов**: TypeScript для фронтенда, Go-клиент для соседнего сервиса.
- **Договор между командами**: фронтенд начинает работу по спецификации и моку, пока бэкенд ещё пишется.
- **Валидацию запросов** по схеме и контрактные тесты.

Кусок спецификации:

```yaml
openapi: 3.1.0
info: { title: Tasktracker API, version: 1.0.0 }
security:
  - bearerAuth: []
paths:
  /v1/tasks/{id}:
    parameters:
      - { name: id, in: path, required: true, schema: { type: integer, minimum: 1 } }
    get:
      responses:
        "200":
          content:
            application/json:
              schema: { $ref: "#/components/schemas/Task" }
        "401": { $ref: "#/components/responses/Error" }
        "404": { $ref: "#/components/responses/Error" }
components:
  securitySchemes:
    bearerAuth: { type: http, scheme: bearer, bearerFormat: JWT }
  schemas:
    Task:
      type: object
      required: [id, title, status]
      properties:
        id: { type: integer }
        title: { type: string }
        status: { type: string, enum: [todo, in_progress, done] }
```

`required` в схеме ответа это обещание клиенту: поле будет всегда. `$ref` позволяет не повторять схемы. Ошибки 401 и 404 описаны так же тщательно, как 200: фронтенду нужно знать их формат.

### Code-first: swaggo/swag

Пишешь код, над хендлером аннотации в комментариях, утилита `swag` генерирует спецификацию.

```go
// GetTask godoc
// @Summary  Задача по id
// @Param    id  path  int  true  "ID задачи"
// @Success  200 {object} TaskResponse
// @Failure  404 {object} ErrorResponse
// @Security BearerAuth
// @Router   /v1/tasks/{id} [get]
func (h *Handlers) GetTask(c *gin.Context) { ... }
```

Плюсы: быстро начать на готовом проекте, документация рядом с кодом, есть `gin-swagger` для Swagger UI. Минусы: аннотации это комментарии, компилятор их не проверяет, и они расходятся с кодом так же, как любая документация; формат комментариев свой, его надо учить. Важная деталь из README: `swag` генерирует **Swagger 2.0**, а не OpenAPI 3.x. Для многих генераторов клиентов это нормально, но возможностей 3.x (например, `oneOf`, нескольких серверов) не будет.

### Spec-first: oapi-codegen

Сначала пишешь `openapi.yaml`, потом `oapi-codegen` генерирует Go-типы и интерфейс сервера. Ты реализуешь интерфейс, а компилятор проверяет, что реализованы все операции с правильными типами.

```yaml
# oapi-codegen.yaml
package: api
output: api.gen.go
generate:
  models: true
  gin-server: true
  strict-server: true
```

По README он поддерживает OpenAPI 3.0 и 3.1, генерирует серверы для Gin, Chi, Echo, Fiber и стандартного `net/http`. Режим `strict-server` даёт обработчики вида `func(ctx, request) (response, error)` без ручного разбора JSON. Плюсы: контракт первичен, расхождение кода и спецификации ловится компиляцией, фронтенд может стартовать по спецификации. Минусы: YAML пишется руками (и он многословен), сгенерированный код надо принимать как есть, нестандартные вещи (особые middleware, потоковые ответы) иногда неудобно выразить.

Генераторы не ставят глобально через `go install`: разные проекты держат разные версии. С Go 1.24 инструмент закрепляется в `go.mod` проекта:

```bash
go get -tool github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest
go tool oapi-codegen -config oapi-codegen.yaml openapi.yaml
```

| | Code-first (swag) | Spec-first (oapi-codegen) |
|---|---|---|
| Источник правды | код и комментарии | YAML-спецификация |
| Расхождение с кодом | ловится только глазами | ловится компиляцией |
| Версия формата | Swagger 2.0 | OpenAPI 3.0/3.1 |
| Порог входа | низкий | нужно писать спецификацию |
| Когда брать | внутренний API, существующий код | публичный API, несколько команд-потребителей |

Честно: в маленькой команде часто начинают с code-first, а переходят на spec-first, когда у API появляется больше одного потребителя. Ручная спецификация без генерации и без тестов, которые её проверяют, почти всегда со временем врёт.

## Версионирование API

Мобильное приложение, выпущенное полгода назад, всё ещё ходит в твой API, и обновить его ты не можешь. Поэтому изменения API делят на два вида.

**Совместимые** (не ломают существующих клиентов):

- новое поле в ответе (клиенты, написанные правильно, игнорируют неизвестные поля);
- новый необязательный параметр или поле запроса;
- новый эндпоинт;
- ослабление требований к запросу (обязательное поле стало необязательным).

**Несовместимые** (ломают):

- удаление или переименование поля ответа;
- смена типа поля (`id` был числом, стал строкой);
- поле ответа, которое было всегда, стало иногда отсутствовать;
- новое обязательное поле в запросе, ужесточение валидации;
- смена смысла: `status=done` раньше значил одно, теперь другое;
- новый код ответа там, где клиент его не ждёт (это серая зона, договаривайся заранее).

Правило: несовместимые изменения делают только в **новой версии**. Самый распространённый способ: версия в пути, `/v1/tasks` и `/v2/tasks`. Альтернативы: заголовок (`Accept: application/vnd.tasktracker.v2+json`) или query-параметр. Версия в пути проще всего для кеширования, логов и curl, поэтому её выбирают чаще.

В Gin версии это просто группы роутов:

```go
v1 := r.Group("/v1")
v1.GET("/tasks/:id", h.getTaskV1)
v2 := r.Group("/v2")
v2.GET("/tasks/:id", h.getTaskV2) // новый формат ответа
```

Сервис и репозиторий общие, отличаются только слои HTTP: структуры запросов и ответов. Это одна из причин держать транспорт отдельно от бизнес-логики (слои ты разбирал в [уроке про архитектуру сервиса](/m08/l04-architecture)).

Старую версию не выключают внезапно. Объявляют срок, в ответах `/v1` ставят заголовки `Deprecation: @1793880000` (RFC 9745, момент объявления устаревшим) и `Sunset: Sun, 31 Oct 2027 23:59:59 GMT` (RFC 8594, когда выключат), смотрят по логам, кто ещё ходит в старую версию.

::: trap «Это же просто переименование»
```go
// было
type TaskResponse struct {
	Title string `json:"title"`
}
// стало: «так понятнее»
type TaskResponse struct {
	Name string `json:"name"`
}
```

Для Go-кода это рефакторинг, для клиента удаление поля `title`: у всех старых версий приложения пропали названия задач. Хорошо: в `/v1` оставить `title` (можно отдавать оба поля на переходный период), `name` вместо `title` только в `/v2`. Тест на JSON-ответ (golden-файл или проверка обязательных полей) ловит такое переименование до релиза, об этом [следующий урок](/m08/l07-testing-service).
:::

## Что выведет программа?

Программа логирует конфиг при старте:

```go
type Config struct {
	Port      int
	JWTSecret string
}

cfg := Config{Port: 8080, JWTSecret: "s3cr3t"}
fmt.Printf("старт с конфигом %+v\n", cfg)
```

<Predict :options="['старт с конфигом {Port:8080 JWTSecret:s3cr3t}', 'старт с конфигом {8080 s3cr3t}', 'старт с конфигом {Port:8080 JWTSecret:***}', 'ошибка компиляции: неэкспортируемое поле']" :correct="0">

`%+v` печатает имена и значения всех полей, секрет уходит в лог как есть. Защита: метод `String()` у `Config` (его вызовет `fmt`) и метод `LogValue()` для slog, оба возвращают версию со скрытыми секретами, как в следующей задаче и в разделе про `LogValue`.

</Predict>

```go
type Config struct {
	Port      int
	JWTSecret string
}

func (c Config) String() string {
	return fmt.Sprintf("Config{Port:%d JWTSecret:[REDACTED]}", c.Port)
}

func main() {
	cfg := Config{Port: 8080, JWTSecret: "s3cr3t"}
	fmt.Printf("%v\n", cfg)
	fmt.Printf("%+v\n", &cfg)
	fmt.Printf("%#v\n", cfg)
}
```

<Predict :options="['все три строки со скрытым секретом', 'первые две скрыты, в третьей секрет виден', 'только первая скрыта', 'все три показывают секрет']" :correct="1">

`%v` и `%+v` вызывают метод `String()`, и для указателя тоже: метод с получателем-значением есть и в наборе методов `*Config`. А `%#v` печатает Go-синтаксис значения и использует метод `GoString()`, а не `String()`, поэтому выводит `main.Config{Port:8080, JWTSecret:"s3cr3t"}`. Полностью закрыть утечку через `fmt` можно, добавив и `GoString()`, но надёжнее не печатать конфиг через `fmt` вовсе, а логировать через `LogValue`.

</Predict>

```go
var lv slog.LevelVar // по умолчанию Info
log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
	Level: &lv,
	ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
		if a.Key == slog.TimeKey {
			return slog.Attr{}
		}
		return a
	},
}))
log.Debug("раз")
lv.Set(slog.LevelDebug)
log.Debug("два")
lv.Set(slog.LevelError)
log.Warn("три")
```

<Predict :options="['level=DEBUG msg=два', 'level=DEBUG msg=раз и level=DEBUG msg=два', 'level=DEBUG msg=два и level=WARN msg=три', 'ничего: уровень задаётся только при создании']" :correct="0">

`LevelVar` меняет порог на лету для всех логгеров, созданных с ним. «раз» отброшен (порог Info), «два» напечатан после переключения на Debug, «три» отброшен, потому что Warn ниже Error. Так делают переключение уровня через админский эндпоинт или сигнал без перезапуска сервиса.

</Predict>

::: interview
<details><summary>Что такое 12-factor и почему конфиг в переменных окружения?</summary>

Набор принципов для облачных сервисов от Heroku. Про конфиг: всё, что меняется между окружениями, хранится в окружении процесса, а не в коде. Один и тот же собранный образ запускается в dev и prod с разными переменными; секреты не попадают в репозиторий; окружение понимают Docker, Kubernetes и CI. Логи по тем же принципам пишутся в stdout, их собирает окружение.

</details>
<details><summary>Зачем валидировать конфиг при старте?</summary>

Fail fast: сервис с плохим конфигом не должен стартовать и принимать трафик, иначе ошибка проявится позже, на конкретном запросе, и деплой не откатится автоматически. Хорошая валидация сообщает все проблемы сразу (`errors.Join`), с именем переменной и ожиданием, но без значений секретов.

</details>
<details><summary>Как хранить секреты и чем плохи переменные окружения?</summary>

Не в git (`.env` в `.gitignore`, в репозитории `.env.example`). Переменные окружения видны в `docker inspect`, `/proc/<pid>/environ`, наследуются дочерними процессами и могут попасть в дамп при падении. Надёжнее монтировать секреты файлами (Kubernetes Secrets, Docker secrets) или получать из менеджера секретов вроде Vault. Секреты нельзя логировать: защищаются через `LogValue` и `ReplaceAttr`.

</details>
<details><summary>Что такое структурное логирование и как устроен slog?</summary>

Запись лога это набор полей (обычно JSON), по которым система логов ищет и агрегирует. В `slog` есть `Logger` (API: `Info`, `With`, `LogAttrs`) и `Handler` (формат и вывод: `JSONHandler`, `TextHandler` или свой). Уровни Debug, Info, Warn, Error, порог задаётся в `HandlerOptions.Level`, динамически через `LevelVar`. `With` добавляет постоянные атрибуты, `ReplaceAttr` переписывает или удаляет атрибуты, `LogValuer` даёт типу своё представление в логе.

</details>
<details><summary>Как связать все строки лога одного запроса?</summary>

Middleware берёт `X-Request-ID` из запроса (проверив формат) или генерирует новый, отдаёт его в заголовке ответа, создаёт `logger.With("request_id", id)` и кладёт в `context.Context` запроса. Все слои берут логгер из контекста. В итоге поиск по `request_id` в системе логов показывает всю историю запроса; этот же id можно передавать в соседние сервисы.

</details>
<details><summary>Что нельзя писать в логи?</summary>

Пароли, токены, заголовки `Authorization` и `Cookie`, API-ключи, секреты конфига, тела запросов целиком, query-строки (в них бывают токены), персональные данные сверх необходимого (152-ФЗ, GDPR). Вместо email логируют `user_id`. Логи доступны большему числу людей и систем, чем база, и хранятся дольше.

</details>
<details><summary>Code-first или spec-first для OpenAPI?</summary>

Code-first (swaggo/swag): аннотации в комментариях, спецификация генерируется из кода. Быстро начать, но комментарии не проверяются компилятором, и swag генерирует Swagger 2.0. Spec-first (oapi-codegen): сначала спецификация, из неё генерируются типы и интерфейс сервера, расхождения ловит компилятор, поддерживается OpenAPI 3.0/3.1. Spec-first лучше для API с несколькими потребителями и параллельной работы фронтенда и бэкенда.

</details>
<details><summary>Какие изменения API совместимые, а какие нет? Как версионировать?</summary>

Совместимые: новые поля ответа, новые необязательные поля и параметры запроса, новые эндпоинты, ослабление валидации. Несовместимые: удаление, переименование или смена типа поля ответа, новое обязательное поле запроса, смена смысла значений. Несовместимые изменения только в новой версии, чаще всего с версией в пути (`/v2`). Старую версию выключают по объявленному графику с заголовками `Deprecation` и `Sunset`.

</details>
:::

## Источники

- `RU` [The Twelve-Factor App: конфигурация](https://12factor.net/ru/config): официальный перевод; прочитай ещё разделы «Логи» и «Одноразовость».
- `RU` [Структурированное логирование в Go с помощью Slog, Хабр](https://habr.com/ru/companies/slurm/articles/798207/): подробный перевод гайда по slog, хендлеры, уровни, `LogValuer`.
- `RU` [Структурированное логирование, документация Gin](https://gin-gonic.com/ru/docs/logging/structured-logging/): как подружить slog с middleware Gin.
- `EN` [Structured Logging with slog, Go Blog](https://go.dev/blog/slog): почему slog устроен именно так, от авторов пакета.
- `EN` [Документация log/slog](https://pkg.go.dev/log/slog): разделы про `HandlerOptions`, `ReplaceAttr`, `LogValuer` и производительность.
- `EN` [OpenAPI Specification](https://swagger.io/specification/): справочник по формату, нужны разделы Paths, Components, Security Scheme.
- `EN` [oapi-codegen](https://github.com/oapi-codegen/oapi-codegen): README со списком серверов (Gin, net/http), strict-server и установкой через `go get -tool`.
- `EN` [swaggo/swag](https://github.com/swaggo/swag): формат аннотаций и интеграция с Gin через gin-swagger.

## Домашка

Пакет `homework/m08/l06-config-logs-docs/`.

<Hw id="ex1" title="Конфиг с валидацией" level="база">

Файл `ex1_config.go`: `Load(args, getenv)` собирает `Config` из флагов, окружения и значений по умолчанию, проверяет все поля и возвращает все ошибки разом через `errors.Join`. Секреты в тексте ошибок появляться не должны.

```bash
cd homework && go test -v ./m08/l06-config-logs-docs/ -run TestLoad
```

</Hw>

<Hw id="ex2" title="Конфиг, безопасный для логов" level="база">

Файл `ex2_logvalue.go`: методы `LogValue` и `String` у `Config`, которые прячут `JWTSecret` и пароль в `DatabaseURL`.

```bash
cd homework && go test -v ./m08/l06-config-logs-docs/ -run TestConfigLogValue
```

</Hw>

<Hw id="ex3" title="Логгер, который не выдаёт секреты" level="база">

Файл `ex3_logger.go`: `NewLogger` на `JSONHandler` с `ReplaceAttr`, который заменяет значения `password`, `token`, `authorization` и других опасных ключей на `[REDACTED]` в любом регистре, в `With` и внутри групп.

```bash
cd homework && go test -v ./m08/l06-config-logs-docs/ -run TestNewLogger
```

</Hw>

<Hw id="ex4" title="Request-scoped логгер" level="посложнее">

Файл `ex4_reqlog.go`: `WithLogger`, `FromContext` и middleware `RequestLogger` с `request_id` (проверка входящего, генерация нового), логгером в контексте и итоговой строкой с уровнем по статусу. Токены из заголовков и query в лог не попадают.

```bash
cd homework && go test -v ./m08/l06-config-logs-docs/ -run 'TestRequestLogger|TestFromContext'
```

</Hw>

<Hw id="ex5" title="Детектор несовместимых изменений" level="посложнее">

Файл `ex5_compat.go`: `BreakingChanges(dir, old, new)` сравнивает две версии схемы и перечисляет изменения, которые сломают существующих клиентов. Для запросов и ответов правила разные: подумай, кто пишет объект, а кто читает.

```bash
cd homework && go test -v ./m08/l06-config-logs-docs/ -run TestBreakingChanges
```

</Hw>

<Hw id="ex6" title="Контракт API в OpenAPI" level="посложнее">

Файл `ex6_openapi.json`: допиши спецификацию API задач (OpenAPI 3.x в JSON). Список требований в начале `ex6_openapi_test.go`: bearer-аутентификация, пять операций, ответы 401 и 404, пагинация, единый формат ошибки, схема `Task`. Тест проверяет смысл, а не оформление.

```bash
cd homework && go test -v ./m08/l06-config-logs-docs/ -run TestOpenAPISpec
```

</Hw>

<Hw id="ex7" title="Генерация сервера по спецификации" level="со звёздочкой">

Без тестов. Вне репозитория курса создай модуль, положи туда свой `ex6_openapi.json`, поставь генератор как инструмент модуля (`go get -tool github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest`) и сгенерируй `models` и `gin-server`. Посмотри на интерфейс `ServerInterface`: какие методы и типы параметров получились? Реализуй один метод заглушкой и запусти сервер. Сравни с тем, сколько кода ты писал руками в домашке m08/l05.

</Hw>
