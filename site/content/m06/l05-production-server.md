# Сервер к продакшну

`http.ListenAndServe(":8080", mux)` работает, и для учебного примера этого достаточно. В продакшене этот однострочник опасен: у сервера нет ни одного таймаута, при деплое он обрывает запросы на середине, логи пишутся строками, которые не найти в Kibana, а настройки зашиты в код. Ничего из этого не видно на ноутбуке и всё вылезает под первой нагрузкой или на первом деплое в Kubernetes.

Этот урок про то, что отличает учебный сервер от сервиса, который не стыдно выкатить: таймауты `http.Server` и атака slowloris, graceful shutdown через `signal.NotifyContext` и `Shutdown`, структурные логи на `log/slog`, конфиг из переменных окружения и пробы `healthz`/`readyz`. В конце модуля собираешь из всего этого мини-проект: сокращатель ссылок на стандартной библиотеке.

На собесах здесь спрашивают «как сделать graceful shutdown», «зачем `ReadHeaderTimeout`», «чем liveness отличается от readiness», «почему логи в JSON».

## Таймауты http.Server

`http.ListenAndServe` создаёт `http.Server` с нулевыми полями, а ноль в таймаутах значит **«ждать вечно»**. Каждое соединение это горутина и файловый дескриптор, и сервер, который ждёт вечно, можно исчерпать без всякой нагрузки.

**Slowloris** это атака, которая держит соединения открытыми, отправляя заголовки запроса по байту раз в несколько секунд. Сервер честно ждёт конца заголовков, ресурсов на каждое соединение мало, но их тысячи, и новые клиенты не могут подключиться. Не нужно ни ботнета, ни трафика: хватит одного ноутбука.

```go
srv := &http.Server{
	Addr:              ":8080",
	Handler:           mux,
	ReadHeaderTimeout: 5 * time.Second,  // заголовки запроса
	ReadTimeout:       15 * time.Second, // весь запрос, включая тело
	WriteTimeout:      30 * time.Second, // запись ответа
	IdleTimeout:       2 * time.Minute,  // keep-alive без запросов
}
log.Fatal(srv.ListenAndServe())
```

Что каждый из них ограничивает:

```text
соединение   читаем заголовки   читаем тело    хендлер пишет ответ   ждём следующий запрос
   |--ReadHeaderTimeout--|
   |---------- ReadTimeout ---------|
                       |------------- WriteTimeout --------------|
                                                                 |------ IdleTimeout ------|
```

- **`ReadHeaderTimeout`** закрывает slowloris. Его ставят всегда. Если он ноль, используется `ReadTimeout`.
- **`ReadTimeout`** ограничивает чтение всего запроса с телом. Для загрузки больших файлов он мешает, поэтому многие ставят только `ReadHeaderTimeout`, а размер тела режут `MaxBytesReader`.
- **`WriteTimeout`** отсчитывается с момента, когда прочитаны заголовки запроса, то есть включает и **время работы хендлера**. Хендлер, который думает 40 секунд при `WriteTimeout` 30 секунд, закончит работу, но клиент ответа не получит. Для долгих стримов (SSE, выгрузки) дедлайн продлевают из хендлера: `http.NewResponseController(w).SetWriteDeadline(...)`.
- **`IdleTimeout`**: сколько держать простаивающее keep-alive соединение. Если ноль, используется `ReadTimeout`.
- **`MaxHeaderBytes`**: лимит размера заголовков, по умолчанию `http.DefaultMaxHeaderBytes` (1 МБ). С Go 1.27 есть и `MaxHeaderValueCount`: лимит числа значений заголовков.

Таймауты сервера рвут соединение, а не отвечают красивой ошибкой. Если нужно «хендлер работает не дольше 2 секунд, иначе 503», это делают контекстом с дедлайном внутри хендлера или обёрткой `http.TimeoutHandler(h, 2*time.Second, "timeout")`.

::: hood Как таймауты устроены внутри
`net/http` реализует их через дедлайны на соединении: `conn.SetReadDeadline` и `conn.SetWriteDeadline`. Дедлайн это момент времени, после которого любой `Read` или `Write` на сокете вернёт ошибку таймаута. Сетевой поллер рантайма (epoll на Linux, kqueue на macOS) следит за дедлайнами и будит горутину, которая заблокирована на чтении. Поэтому таймаут не требует отдельной горутины-таймера на каждое соединение: горутина соединения просто просыпается с ошибкой, сервер закрывает сокет и выходит.

Перед чтением нового запроса сервер ставит дедлайн чтения на `now + ReadHeaderTimeout`, после заголовков переставляет на `ReadTimeout` (или снимает), а дедлайн записи ставит на `now + WriteTimeout`. Ожидание следующего запроса на keep-alive соединении идёт с дедлайном `IdleTimeout`.
:::

## Graceful shutdown

При деплое старый процесс получает сигнал `SIGTERM` (в Kubernetes, systemd, Docker) или `SIGINT` (Ctrl+C). Если просто умереть, запросы, которые сейчас обрабатываются, оборвутся: клиент получит обрыв соединения, а то и половину записанных в базу данных. Graceful shutdown это «перестать принимать новое, доделать начатое, выйти».

```go
func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	srv := &http.Server{Addr: ":8080", Handler: mux, ReadHeaderTimeout: 5 * time.Second}

	serveErr := make(chan error, 1)
	go func() { serveErr <- srv.ListenAndServe() }()

	select {
	case err := <-serveErr: // не смогли стартовать: порт занят и т.п.
		slog.Error("server failed", "err", err)
		os.Exit(1)
	case <-ctx.Done(): // пришёл сигнал
	}
	stop() // повторный Ctrl+C убьёт процесс сразу

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("graceful shutdown failed", "err", err)
		srv.Close()
	}
	slog.Info("server stopped")
}
```

Разберём:

- **`signal.NotifyContext`** (Go 1.16) возвращает контекст, который отменяется при получении одного из сигналов. Удобнее старого `signal.Notify` с каналом. `stop()` после первого сигнала возвращает поведение по умолчанию: второй Ctrl+C завершит процесс немедленно, если shutdown завис.
- **`ListenAndServe` блокируется**, поэтому живёт в горутине. После вызова `Shutdown` он **сразу** возвращает `http.ErrServerClosed`: это штатный сигнал, а не ошибка.
- **`Shutdown(ctx)`** закрывает листенеры (новые соединения не принимаются), закрывает простаивающие соединения и ждёт, пока активные запросы доработают. Если `ctx` истёк раньше, возвращает его ошибку, и тогда `Close()` рвёт оставшееся.
- `main` должен **дождаться** возврата `Shutdown`. Если `main` выйдет раньше, процесс завершится вместе со всеми горутинами, и смысл пропадёт.
- `Shutdown` **не ждёт** захваченных (hijacked) соединений вроде WebSocket. Их уведомляют через `srv.RegisterOnShutdown`.

::: trap Shutdown с уже отменённым контекстом
```go
// плохо
<-ctx.Done()
srv.Shutdown(ctx) // ctx уже отменён сигналом
```

`Shutdown` видит, что контекст уже отменён, и сразу возвращает ошибку, не дождавшись ни одного запроса. Формально graceful shutdown есть, фактически запросы рвутся так же, как без него. Это одна из самых частых ошибок в коде на собесах и в проде.

```go
// хорошо: новый контекст с таймаутом на сам shutdown
shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()
srv.Shutdown(shutdownCtx)
```
:::

Таймаут shutdown должен укладываться в то, сколько оркестратор ждёт после `SIGTERM`. В Kubernetes это `terminationGracePeriodSeconds`, по умолчанию 30 секунд, после чего приходит `SIGKILL`, который не перехватить.

### Порядок остановки в Kubernetes

Есть тонкость: под получает `SIGTERM` примерно одновременно с тем, как его убирают из балансировки, и какое-то время на него ещё приходит трафик. Если сразу закрыть листенер, эти запросы получат `connection refused`. Поэтому аккуратный порядок такой:

1. Получили сигнал: переключили `readyz` в «не готов» (503).
2. Подождали несколько секунд, пока балансировщик это заметит и перестанет слать трафик.
3. `Shutdown` с таймаутом: доделали начатые запросы.
4. Закрыли пулы БД, отправили последние метрики, вышли.

## Логи: log/slog

Логи читают не люди глазами, а системы сбора (Loki, Elasticsearch, Cloud Logging), и в них ищут: «все ошибки по заказу 7 за час», «все запросы с request_id abc». Строка `2026/11/05 order 7 failed: timeout` для этого плохо подходит. **Структурный лог** это запись с полями:

```json
{"time":"2026-11-05T12:00:00Z","level":"ERROR","msg":"payment failed","order_id":7,"request_id":"abc","err":"timeout"}
```

С Go 1.21 в стандартной библиотеке есть `log/slog`:

```go
logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
slog.SetDefault(logger)

slog.Info("server started", "addr", ":8080")
slog.Error("payment failed", "order_id", 7, "err", err)
logger.With("component", "billing").Warn("retrying", "attempt", 2)
```

Устройство:

- **`Logger`** это фасад с методами `Debug`, `Info`, `Warn`, `Error` и их версиями с контекстом (`InfoContext`). Атрибуты передаются парами ключ-значение или типизированно: `slog.Int("order_id", 7)`.
- **`Handler`** решает, куда и в каком формате писать: `TextHandler` (`key=value`, удобно локально) или `JSONHandler` (для продакшена). Можно писать свои и оборачивать чужие.
- **`Record`**: одна запись (время, уровень, сообщение, атрибуты), которую логгер отдаёт хендлеру.
- **Уровни** это числа: `Debug` -4, `Info` 0, `Warn` 4, `Error` 8. Хендлер отбрасывает всё ниже своего уровня. `slog.LevelVar` позволяет менять уровень на лету.
- **`With`** создаёт дочерний логгер с постоянными атрибутами, **`WithGroup`** вкладывает атрибуты в группу (`"http":{"status":200}`).
- `slog.SetDefault` заодно перенаправляет в этот хендлер старый пакет `log`, так что `log.Printf` из библиотек попадёт в тот же поток.
- С Go 1.26 `slog.NewMultiHandler` пишет одну запись сразу в несколько хендлеров (например, JSON в stdout и ошибки ещё куда-то).

```go
log := slog.New(slog.NewTextHandler(os.Stdout, nil))
log = log.With("svc", "shop")
log.Debug("кеш прогрет")
log.Info("заказ", "id", 7, "user")
```

<Predict :options="['две строки: DEBUG и INFO', 'одна строка INFO c id=7 user=', 'одна строка INFO c id=7 !BADKEY=user', 'паника: нечётное число аргументов']" :correct="2">

Вывод: `time=... level=INFO msg=заказ svc=shop id=7 !BADKEY=user`. `Debug` не напечатан: уровень хендлера по умолчанию `Info`. Атрибуты из `With` идут перед атрибутами вызова. Последний аргумент `"user"` остался без пары, и `slog` не паникует, а помечает его ключом `!BADKEY`. `go vet` ловит такие вызовы, так что гоняй его в CI.

</Predict>

### request_id в каждой строке

Задача: в каждой строке лога, относящейся к запросу, должен быть его ID, но тащить логгер через все функции неудобно. Решение: middleware кладёт ID в контекст, а **свой `slog.Handler`-обёртка** достаёт его из контекста при каждой записи. Код сервисов пишет `logger.InfoContext(ctx, ...)`, и ID появляется сам.

```go
type ContextHandler struct{ slog.Handler }

func (h ContextHandler) Handle(ctx context.Context, r slog.Record) error {
	if id, ok := ctx.Value(requestIDKey{}).(string); ok {
		r.AddAttrs(slog.String("request_id", id))
	}
	return h.Handler.Handle(ctx, r)
}
```

Встраивание даёт остальные методы интерфейса, но `WithAttrs` и `WithGroup` встроенного хендлера возвращают **внутренний** хендлер без обёртки, и `logger.With(...)` молча потеряет request_id. Их нужно переопределить, это задание в домашке.

::: trap Секреты и персональные данные в логах
```go
// плохо
slog.Info("login", "request", req)            // вся структура с паролем
slog.Debug("db config", "dsn", cfg.DatabaseURL) // postgres://user:password@...
```

Логи хранятся долго, доступ к ним есть у многих, а утечка пароля из лога это такая же утечка. Не логируй тела запросов целиком, токены, пароли, DSN с паролями, номера карт. Для типов с секретами реализуй `slog.LogValuer`, который отдаёт безопасное представление:

```go
type Password string

func (Password) LogValue() slog.Value { return slog.StringValue("***") }
```
:::

## Конфиг из окружения

Принцип 12-factor: конфигурация, которая отличается между окружениями (адреса баз, ключи, порты, уровни логов), живёт в **переменных окружения**, а не в коде и не в закоммиченном файле. Один и тот же образ Docker запускается в dev, staging и prod с разными переменными.

Правила:

- **Читай и проверяй всё при старте**, в одном месте, и падай сразу с понятной ошибкой. Хуже всего, когда сервис стартовал, а через час упал, потому что `TIMEOUT=30` без единиц не парсится в коде, который впервые выполнился.
- **Собирай все ошибки** конфига разом (`errors.Join`), чтобы не чинить по одной переменной за деплой.
- **Разумные умолчания** для необязательного, обязательное без умолчаний (`DATABASE_URL`).
- **Не логируй секреты** при выводе конфига на старте.
- **Тестируемость**: функция принимает `getenv func(string) string`, а не зовёт `os.Getenv` напрямую. В тесте подставляешь карту.

`os.Getenv` возвращает `""` и для «не задано», и для «задано пустым». Если разница важна, есть `os.LookupEnv`, который возвращает ещё и `bool`.

```go
d, err := time.ParseDuration("30")
fmt.Println(d, err)
```

<Predict :options="['30s <nil>', '30ns <nil>', '0s time: missing unit in duration \'30\'', '30ms <nil>']" :correct="2">

`time.ParseDuration` требует единицы: `30s`, `500ms`, `1m30s`. Голое число это ошибка `time: missing unit in duration "30"` (в выводе там двойные кавычки), а `d` остаётся нулём. Без проверки ошибки получишь таймаут 0, то есть «ждать вечно». Исключение: строка `"0"` парсится без единиц.

</Predict>

## healthz и readyz

Оркестратору нужно знать две разные вещи:

- **Liveness (`/healthz`)**: процесс жив и не завис? Если нет, его **перезапустят**. Проверка должна быть тривиальной: ответить 200. Не проверяй здесь базу: если база упала, перезапуск всех подов её не починит, зато устроит каскад рестартов.
- **Readiness (`/readyz`)**: готов принимать трафик? Если нет, под **уберут из балансировки**, но не убьют. Здесь проверяют зависимости (пинг базы, Redis) с коротким таймаутом, а при shutdown отвечают 503 заранее.

```go
mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("ok"))
})
mux.Handle("GET /readyz", Readyz(time.Second, map[string]Check{
	"db": func(ctx context.Context) error { return pool.Ping(ctx) },
}))
```

Проверки готовности запускают параллельно и с общим таймаутом: пробу дёргают часто, и она не должна сама становиться медленной. Ответ не кешируют (`Cache-Control: no-store`). Эти ручки обычно не пускают в access-лог, иначе они забьют его шумом.

## Что выведет программа?

```go
srv := &http.Server{Addr: "127.0.0.1:0"}
srv.Shutdown(context.Background())
err := srv.ListenAndServe()
fmt.Println(err)
```

<Predict :options="['<nil>', 'http: Server closed', 'listen tcp 127.0.0.1:0: bind: address already in use', 'программа зависнет']" :correct="1">

После `Shutdown` сервер нельзя переиспользовать: любой `Serve` и `ListenAndServe` сразу возвращает `http.ErrServerClosed` с текстом `http: Server closed`. В штатной остановке ты увидишь эту же ошибку из горутины с `ListenAndServe`, поэтому её отфильтровывают: `if !errors.Is(err, http.ErrServerClosed)`.

</Predict>

::: interview
<details><summary>Как сделать graceful shutdown HTTP-сервера в Go?</summary>

`signal.NotifyContext` на `SIGINT` и `SIGTERM`; `ListenAndServe` в горутине; ждать отмены контекста или ошибки старта; затем `srv.Shutdown` с **новым** контекстом с таймаутом (а не с уже отменённым); если таймаут истёк, `srv.Close()`. `ErrServerClosed` это не ошибка. `main` ждёт завершения `Shutdown`. WebSocket и другие hijacked соединения `Shutdown` не ждёт, их уведомляют через `RegisterOnShutdown`. В Kubernetes перед `Shutdown` переключают readiness в 503 и делают паузу, чтобы балансировщик успел убрать под.

</details>
<details><summary>Какие таймауты есть у http.Server и что будет без них?</summary>

`ReadHeaderTimeout` (заголовки), `ReadTimeout` (весь запрос с телом), `WriteTimeout` (от прочтения заголовков до конца ответа, включает работу хендлера), `IdleTimeout` (keep-alive простой), плюс `MaxHeaderBytes`. Без них соединения могут висеть вечно: атака slowloris медленными заголовками исчерпывает горутины и дескрипторы с одной машины. Минимум: `ReadHeaderTimeout`.

</details>
<details><summary>Что такое slowloris?</summary>

Атака на исчерпание соединений: клиент открывает много соединений и шлёт заголовки по чуть-чуть, никогда не заканчивая запрос. Сервер без таймаута на чтение заголовков держит все эти соединения, и для настоящих клиентов места не остаётся. Трафика почти нет, поэтому защита пропускной способностью не помогает. Лечится `ReadHeaderTimeout` и лимитами на балансировщике.

</details>
<details><summary>Чем liveness отличается от readiness?</summary>

Liveness: жив ли процесс; провал ведёт к перезапуску контейнера. Должна быть тривиальной, без проверки зависимостей, иначе падение базы вызовет каскад рестартов. Readiness: можно ли слать трафик; провал убирает под из балансировки без перезапуска. Проверяет зависимости с коротким таймаутом и переключается в «не готов» при graceful shutdown.

</details>
<details><summary>Зачем структурные логи и что умеет slog?</summary>

Структурные логи (обычно JSON) машиночитаемы: системы сбора индексируют поля, и можно искать по `request_id`, `order_id`, уровню. `slog` (Go 1.21+): `Logger` с уровнями и контекстными методами, `Handler` (Text, JSON, свои), атрибуты парами или типизированно, `With` и `WithGroup`, `LevelVar` для смены уровня на лету, `LogValuer` для безопасного представления значений, `SetDefault` перехватывает и пакет `log`. В 1.26 добавлен `NewMultiHandler`.

</details>
<details><summary>Как добавить request_id во все логи запроса?</summary>

Middleware генерирует или берёт ID из заголовка и кладёт в контекст. Свой `slog.Handler`-обёртка в `Handle` достаёт ID из контекста и добавляет атрибут; `WithAttrs` и `WithGroup` переопределены, чтобы возвращать обёртку. Код пишет `logger.InfoContext(ctx, ...)`. Альтернатива: класть в контекст логгер с `With("request_id", id)`, но тогда его нужно доставать из контекста в каждом месте.

</details>
<details><summary>Почему конфиг в переменных окружения и как его читать?</summary>

12-factor: один артефакт для всех окружений, отличия только в окружении, секреты не попадают в git. Читать и валидировать при старте целиком, падать сразу со всеми ошибками, иметь умолчания для необязательного, не логировать секреты. Для тестируемости принимать `getenv` функцией. `os.LookupEnv` отличает «не задано» от «пусто».

</details>
<details><summary>Чем WriteTimeout опасен для долгих хендлеров?</summary>

Он отсчитывается от прочтения заголовков запроса и включает время работы хендлера. Если хендлер работает дольше, соединение закроется на записи ответа: работа сделана, клиент ответа не получил. Для долгих операций и стримов дедлайн продлевают через `http.NewResponseController(w).SetWriteDeadline`, а «хендлер не дольше N» делают контекстом с дедлайном или `http.TimeoutHandler`, который ответит 503.

</details>
:::

## Источники

- `RU` [Реализация Graceful Shutdown в Go, Хабр](https://habr.com/ru/articles/771626/): порядок работы `Shutdown`, `signal.NotifyContext`, `BaseContext` для долгих хендлеров.
- `RU` [Graceful Shutdown в Go на практике, Хабр](https://habr.com/ru/articles/908344/): readiness перед остановкой и согласование таймаутов с оркестратором.
- `RU` [Структурированное логирование в Go с помощью Slog, Хабр (Слёрм)](https://habr.com/ru/companies/slurm/articles/798207/): Logger, Handler, Record, уровни, `LogValuer` на примерах.
- `RU` [Конфигурация, The Twelve-Factor App](https://12factor.net/ru/config): почему конфиг в окружении, коротко и по делу.
- `RU` [Go в примерах: сигналы](https://gobyexample.ru/signals.html): как процесс получает `SIGINT` и `SIGTERM`.
- `EN` [The complete guide to Go net/http timeouts](https://blog.cloudflare.com/the-complete-guide-to-golang-net-http-timeouts/): серверная половина статьи, схемы таймаутов сервера.
- `EN` [Structured Logging with slog](https://go.dev/blog/slog): официальный пост с мотивацией и устройством пакета.
- `EN` [Документация log/slog](https://pkg.go.dev/log/slog): раздел про написание своих хендлеров и `LogValuer`.

## Домашка

<Hw id="ex1" title="Конфиг из окружения" level="база">

Файл `homework/m06/l05-production-server/ex1_config.go`, функция `LoadConfig`: порт с проверкой диапазона, уровень логов, длительность, обязательная `DATABASE_URL`, все ошибки сразу через `errors.Join`.

```bash
cd homework && go test -v ./m06/l05-production-server/ -run TestLoadConfig
```

</Hw>

<Hw id="ex2" title="Сервер и graceful shutdown" level="посложнее">

Файл `ex2_server.go`: `NewServer` с таймаутами и `Run`, который останавливает сервер по отмене контекста, дожидается запросов в полёте, рвёт зависшие по таймауту и не считает `ErrServerClosed` ошибкой. Тесты проверяют, что запрос в полёте получает ответ, а новые соединения уже не принимаются.

```bash
cd homework && go test -race -v ./m06/l05-production-server/ -run 'TestNewServer|TestRun'
```

</Hw>

<Hw id="ex3" title="request_id в slog" level="посложнее">

Файл `ex3_slog.go`: `ContextHandler`, который добавляет `request_id` из контекста и не теряет обёртку в `With` и `WithGroup`.

```bash
cd homework && go test -v ./m06/l05-production-server/ -run TestContextHandler
```

</Hw>

<Hw id="ex4" title="readyz" level="посложнее">

Файл `ex4_health.go`, функция `Readyz`: проверки параллельно, общий таймаут, ответ не позже таймаута даже при зависшей проверке, JSON со статусом каждой зависимости, 200 или 503.

```bash
cd homework && go test -race -v ./m06/l05-production-server/ -run TestReadyz
```

</Hw>

<Hw id="ex5" title="Slowloris своими руками" level="со звёздочкой">

Подними сервер через `http.ListenAndServe` (без таймаутов) и открой к нему `nc localhost 8080`. Набери `GET / HTTP/1.1` и `Host: x`, но не отправляй пустую строку. Соединение висит сколько угодно. Повтори с `NewServer` из ex2: через 5 секунд сервер закроет соединение. Посмотри на сервере, сколько горутин живёт при 100 таких соединениях (`runtime.NumGoroutine` в отдельной ручке) в обоих вариантах.

</Hw>

## Мини-проект модуля

**URL-shortener на стандартной библиотеке.** Сервис принимает длинную ссылку, выдаёт короткую, редиректит по ней и считает переходы. Всё на `net/http`, `encoding/json`, `log/slog`, без внешних зависимостей. Это первый твой сервис целиком: роутинг, JSON API, валидация, конкурентное хранилище, конфиг, логи, graceful shutdown.

### Ядро: код с тестами

Пакет `homework/m06/project-shortener/`, тесты уже написаны:

- `code.go`: `NewCode(n)` генерирует код из base62 на `crypto/rand`; `ValidateURL` пропускает только `http`/`https` с хостом, без userinfo (`https://bank.ru@evil.example`), не длиннее 2048 байт.
- `store.go`: `Store`, потокобезопасное хранилище в памяти. `Create` с повтором при коллизии кода (не больше 5 попыток), `Resolve` с подсчётом переходов, `Stats`. Генератор кода и часы передаются в конструктор, поэтому тесты детерминированы.
- `handler.go`: `NewHandler(store, baseURL)` с маршрутами `POST /api/links`, `GET /{code}` (302), `GET /api/links/{code}/stats` и ошибками в едином формате.

```bash
cd homework && go test -race -v ./m06/project-shortener/
```

Рекомендуемый порядок: `code.go`, потом `store.go` (тест на 1000 параллельных переходов поймает гонку и потерянные инкременты), потом `handler.go`.

### Сервер: ТЗ

Собери исполняемый файл `homework/m06/project-shortener/cmd/shortener/main.go` (пакет `main` внутри того же модуля, импорт `goschool/homework/m06/project-shortener`). Тестов на него нет, проверяешь сам по чеклисту.

1. **Конфиг из окружения**: `PORT` (по умолчанию 8080), `BASE_URL` (по умолчанию `http://localhost:<PORT>`), `LOG_LEVEL` (info), `SHUTDOWN_TIMEOUT` (10s). Ошибки конфига печатаются все сразу, процесс выходит с кодом 1. Можно взять `LoadConfig` из домашки l05 и убрать `DATABASE_URL`.
2. **Логи**: `slog` с `JSONHandler` в stdout, уровень из конфига, `ContextHandler` из домашки, чтобы у каждой строки запроса был `request_id`.
3. **Middleware** (снаружи внутрь): recovery, request ID, access-лог (метод, путь, паттерн, статус, байты, длительность в мс). `/healthz` и `/readyz` в access-лог не пишутся.
4. **Сервер** с таймаутами из `NewServer` (l05).
5. **Graceful shutdown**: `signal.NotifyContext` на `SIGINT` и `SIGTERM`, `readyz` переключается в 503, затем `Shutdown` с таймаутом из конфига, лог «server stopped».
6. **Пробы**: `GET /healthz` всегда 200, `GET /readyz` 200, пока сервер не останавливается.

### Чеклист проверки

```bash
cd homework && go run ./m06/project-shortener/cmd/shortener
```

- [ ] `curl -i -X POST localhost:8080/api/links -H 'Content-Type: application/json' -d '{"url":"https://go.dev/doc/"}'` даёт `201`, `Location` и JSON с `code` и `short_url`.
- [ ] `curl -i localhost:8080/<code>` даёт `302` и `Location: https://go.dev/doc/`; `curl -L` доходит до go.dev.
- [ ] После двух переходов `curl localhost:8080/api/links/<code>/stats` показывает `"hits":2`.
- [ ] `-d '{"url":"javascript:alert(1)"}'` даёт `422` с `"code":"invalid_url"`; битый JSON `400`; без `Content-Type` `415`; тело 5 КБ `413`.
- [ ] `curl -i -X DELETE localhost:8080/<code>` даёт `405` и `Allow: GET, HEAD`.
- [ ] `curl -i localhost:8080/nope` даёт `404` в JSON-формате ошибки.
- [ ] `curl -i -H 'X-Request-Id: my-id-1' localhost:8080/healthz`: в ответе тот же `X-Request-Id`; у обычного запроса ID генерируется, и он же есть в строке лога.
- [ ] `LOG_LEVEL=loud PORT=abc go run ./m06/project-shortener/cmd/shortener` печатает обе ошибки и выходит с кодом 1.
- [ ] Добавь временную ручку `GET /slow`, которая спит 5 секунд. Запусти `curl localhost:8080/slow`, сразу нажми Ctrl+C в окне сервера: curl получает ответ, сервер пишет «server stopped» и выходит. С `SHUTDOWN_TIMEOUT=1s` curl получает обрыв, а сервер логирует ошибку shutdown.
- [ ] `go vet ./m06/project-shortener/...` чистый, `go test -race ./m06/project-shortener/` зелёный.

Идеи на потом (не обязательно): свой код ссылки в запросе (`"code":"go-docs"` с проверкой занятости и 409), срок жизни ссылки, хранение в PostgreSQL после модуля про базы, rate limit на создание.
