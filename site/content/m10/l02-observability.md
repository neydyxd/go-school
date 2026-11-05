# Наблюдаемость: логи, метрики, трейсы

Сервис в проде это чёрный ящик, к которому нельзя подключить отладчик. Пользователь пишет «у меня всё тормозит», и у тебя есть только то, что сервис сам рассказал о себе: логи, метрики и трейсы. **Наблюдаемость** (observability) это способность ответить на вопрос «что сейчас происходит внутри» по этим внешним сигналам, не выкатывая новый код с отладочными `Println`.

Три сигнала решают разные задачи:

- **Логи** отвечают на вопрос «что случилось с вот этим запросом». Подробно, дорого хранить, плохо агрегируются.
- **Метрики** отвечают на вопрос «как система ведёт себя в целом»: сколько запросов в секунду, какая доля ошибок, какой 99-й перцентиль задержки. Дёшево, хранятся месяцами, по ним строят графики и алерты.
- **Трейсы** отвечают на вопрос «куда ушло время в этом запросе», когда запрос проходит через пять сервисов, базу и очередь.

Сюда же относятся проверки здоровья (`/healthz`, `/readyz`) и корректная остановка: оркестратор тоже «наблюдает» за сервисом и решает, слать ли ему трафик. Логи, health-проверки и базовый graceful shutdown ты уже делал в модулях 06 и 08, поэтому главное в этом уроке новое: метрики Prometheus, пауза drain перед остановкой и трейсы. На собесе почти всегда спросят про типы метрик, кардинальность, liveness против readiness и graceful shutdown.

## Логи: что уже есть и что добавить

Структурные логи на `log/slog` ты уже настраивал. В [уроке «Сервер к продакшну»](/m06/l05-production-server) был JSON в stdout, `ContextHandler` с `request_id` в каждой строке и правила, что нельзя логировать. В [уроке про конфиг и логи](/m08/l06-config-logs-docs) были уровни, `With`, группы и логгер запроса через контекст, а в [уроке про middleware](/m08/l03-middleware) request id для Gin. Повторять это не будем. Здесь о том, что превращает логи из текста в часть наблюдаемости.

- **Корреляция между сервисами.** `request_id` склеивает строки одного запроса внутри сервиса. Чтобы склеить их по всей цепочке, тот же id пробрасывают в исходящие запросы, а при трейсинге в каждую строку пишут `trace_id` и `span_id`. Тогда из трейса можно перейти к логам запроса и обратно.
- **Логи стоят денег.** Строка на каждый успешный запрос при 5 000 RPS это гигабайты в час. На вопрос «сколько и как быстро» отвечают метрики, а access-лог пишут только для медленных и ошибочных запросов или сэмплируют.
- **Алертят по метрикам, а не по логам.** Алерт «в логе встретилось `payment failed`» перестаёт работать молча, когда кто-то поправил текст сообщения. Счётчик `payments_failed_total` так не сломается.

`trace_id` в каждую строку добавляет та же обёртка над `slog.Handler`, что и `ContextHandler` из m06, только id берётся из спана в контексте (пакет `go.opentelemetry.io/otel/trace`, про спаны ниже). `WithAttrs` и `WithGroup` переопределяются так же, как в m06:

```go
type TraceHandler struct{ slog.Handler }

func (h TraceHandler) Handle(ctx context.Context, r slog.Record) error {
	if sc := trace.SpanContextFromContext(ctx); sc.IsValid() {
		r.AddAttrs(
			slog.String("trace_id", sc.TraceID().String()),
			slog.String("span_id", sc.SpanID().String()),
		)
	}
	return h.Handler.Handle(ctx, r)
}
// slog.InfoContext(ctx, "заказ создан") внутри спана:
// {"level":"INFO","msg":"заказ создан","trace_id":"b8a2a959ec8f2d0133ca865d87f74192","span_id":"6eceda277f1cf318"}
```

::: trap Одна ошибка на пяти уровнях
```go
// плохо: каждый слой логирует и пробрасывает, одна ошибка даёт три строки
// уровня ERROR, и по логу кажется, что сломалось три разных вещи
func (r *Repo) Get(ctx context.Context, id int64) (Task, error) {
	t, err := r.query(ctx, id)
	if err != nil {
		slog.ErrorContext(ctx, "repo get", "err", err)
		return Task{}, err
	}
	return t, nil
}
// ... и то же в сервисе, и то же в хендлере
```

```go
// хорошо: внизу ошибку оборачивают контекстом, логируют один раз там,
// где её обрабатывают (хендлер или middleware на границе сервиса)
if err != nil {
	return Task{}, fmt.Errorf("repo get task %d: %w", id, err)
}
```

Итоговая строка `handler: service: repo get task 7: connection refused` несёт всю цепочку, и алерт по счётчику ошибок посчитает одну ошибку, а не три.
:::

## Метрики Prometheus

**Prometheus** это база временных рядов и стандарт де-факто для метрик. Модель такая: сервис отдаёт текущие значения своих метрик по HTTP на `/metrics`, а Prometheus раз в 15-30 секунд сам приходит и забирает их (**pull-модель**). Каждый сбор называется scrape.

**Временной ряд** это имя метрики плюс набор лейблов:

```text
http_requests_total{method="GET",route="GET /users/{id}",code="200"}  1027
http_requests_total{method="GET",route="GET /users/{id}",code="500"}  3
```

Это два разных ряда. Prometheus хранит для каждого пары (время, значение).

### Четыре типа

**Counter** (счётчик) только растёт: число запросов, ошибок, отправленных писем. Уменьшить нельзя, при перезапуске процесса сбрасывается в ноль. Само значение счётчика почти бесполезно («с момента старта было 1027 запросов»), смотрят на скорость роста: `rate(http_requests_total[5m])` даёт запросы в секунду за последние 5 минут, а `rate` сам корректно обрабатывает сбросы при рестарте. Имя по соглашению заканчивается на `_total`.

**Gauge** (датчик) может расти и падать: текущее число горутин, активных соединений, размер очереди, температура. Смотрят на значение как есть.

**Histogram** (гистограмма) раскладывает наблюдения по корзинам (**бакетам**): сколько запросов уложилось в 5 мс, в 10 мс, в 25 мс и так далее. Из бакетов Prometheus считает перцентили: `histogram_quantile(0.99, rate(http_request_duration_seconds_bucket[5m]))`.

**Summary** считает перцентили прямо в процессе. Звучит удобнее, но такие перцентили **нельзя агрегировать**: средний p99 по десяти подам не равен p99 всего сервиса. Гистограммы агрегируются (бакеты просто складываются), поэтому для задержек почти всегда берут гистограмму.

```go
var ordersCreated = prometheus.NewCounterVec(prometheus.CounterOpts{
	Name: "shop_orders_created_total",
	Help: "Сколько заказов создано.",
}, []string{"payment"})

reg := prometheus.NewRegistry()
reg.MustRegister(
	collectors.NewGoCollector(),      // go_goroutines, go_memstats_*, GC
	collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}), // CPU, RSS, fd
	ordersCreated,
)

ordersCreated.WithLabelValues("card").Inc()
ordersCreated.WithLabelValues("card").Add(2)
ordersCreated.WithLabelValues("sbp").Inc()

mux.Handle("GET /metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
```

`curl localhost:8080/metrics` отдаёт текстовый формат:

```text
go_goroutines 9
process_resident_memory_bytes 1.4319616e+07
# HELP shop_orders_created_total Сколько заказов создано.
# TYPE shop_orders_created_total counter
shop_orders_created_total{payment="card"} 3
shop_orders_created_total{payment="sbp"} 1
```

Можно не создавать реестр и регистрировать всё в глобальном `prometheus.DefaultRegisterer` (через `promauto` это вообще одна строка), а отдавать `promhttp.Handler()`. Для маленького сервиса нормально, но глобальный реестр мешает тестам: второй вызов конструктора с той же метрикой паникует с `duplicate metrics collector registration attempted`. Свой реестр, переданный параметром, лечит это.

::: hood Как устроена гистограмма
Гистограмма с бакетами `0.1, 0.25, 0.5, 1` это на самом деле несколько счётчиков: по одному на бакет, плюс `_sum` (сумма всех наблюдений) и `_count` (их число). Бакеты **кумулятивные**: `le="0.5"` означает «наблюдений меньше или равно 0,5», то есть включает и всё, что попало в меньшие бакеты. Последний бакет `le="+Inf"` всегда равен `_count`.

`histogram_quantile` не знает настоящих значений, только счётчики в корзинах. Он находит бакет, в который попадает нужный перцентиль, и **линейно интерполирует** внутри него. Отсюда два следствия: точность зависит от границ бакетов (если все запросы между 0,5 и 1 секундой, p99 будет «где-то там»), а границы надо выбирать под свой SLO. `prometheus.DefBuckets` (от 5 мс до 10 с) подходят типичному веб-сервису.

Каждый бакет это отдельный временной ряд. Гистограмма с `DefBuckets` даёт 14 рядов на каждую комбинацию лейблов: 11 бакетов, `+Inf`, `_sum` и `_count`. На 20 маршрутов это уже 280 рядов.
:::

```go
h := prometheus.NewHistogram(prometheus.HistogramOpts{
	Name:    "req_seconds",
	Buckets: []float64{0.1, 0.25, 0.5, 1},
})
for _, v := range []float64{0.1, 0.3, 0.3, 2} {
	h.Observe(v)
}
// что в строке req_seconds_bucket{le="0.5"}?
```

<Predict :options="['2', '3', '0', '4']" :correct="1">

Бакеты кумулятивные: `le="0.5"` считает все наблюдения не больше 0,5. Это 0,1 (граница включается) и две 0,3, итого 3. Полностью: `le="0.1"` 1, `le="0.25"` 1, `le="0.5"` 3, `le="1"` 3, `le="+Inf"` 4, `_sum` 2.7, `_count` 4.

</Predict>

### RED и USE: что вообще мерить

Чтобы не изобретать набор метрик каждый раз, есть два метода.

**RED** для сервисов, которые обрабатывают запросы:

- **R**ate: запросов в секунду, `rate(http_requests_total[5m])`;
- **E**rrors: доля ошибок, `rate(http_requests_total{code=~"5.."}[5m]) / rate(http_requests_total[5m])`;
- **D**uration: распределение задержек, гистограмма и её p50/p95/p99.

**USE** для ресурсов (CPU, память, пул соединений к базе, очередь):

- **U**tilization: насколько ресурс занят (пул: 18 из 20 соединений);
- **S**aturation: сколько работы ждёт (запросов в очереди за соединением);
- **E**rrors: ошибки ресурса (таймауты получения соединения).

Для HTTP-сервиса минимальный набор: счётчик запросов с лейблами `method`, `route`, `code` и гистограмма длительности с `method`, `route`. Это ровно первая задача домашки. Плюс метрики пула базы (у `pgxpool.Stat()` есть всё нужное) и бизнес-метрики вроде созданных заказов: по ним видно, что сервис не просто отвечает 200, а делает свою работу.

### Лейблы и кардинальность

**Кардинальность** это число временных рядов. Каждая уникальная комбинация лейблов создаёт новый ряд, а каждый ряд Prometheus держит в памяти. Классическая авария:

```go
// плохо: каждый id пользователя, каждый URL с параметрами это новый ряд
requests.WithLabelValues(r.Method, r.URL.Path, code).Inc()  // /users/1, /users/2, ...
requests.WithLabelValues(userID).Inc()
```

Миллион пользователей = миллион рядов, Prometheus съедает всю память и падает, а вместе с ним все алерты компании. Правило: значения лейблов должны быть из **маленького ограниченного множества**. Метод (десяток), шаблон маршрута (десятки), код ответа (десяток), тип оплаты. Никогда: id, email, полный URL, текст ошибки, IP.

С Go 1.23 `http.ServeMux` кладёт сопоставленный шаблон в `r.Pattern` (сами шаблоны с методом и `{id}` появились в 1.22, поле `Pattern` на год позже): `"GET /users/{id}"` вместо `/users/42`. Это готовое значение лейбла `route`.

::: trap Лейбл с неограниченными значениями
```go
// плохо: ряд на каждый уникальный путь и текст ошибки
requests.WithLabelValues(r.URL.Path).Inc()
failures.WithLabelValues(err.Error()).Inc()
```

```go
// хорошо: шаблон маршрута и класс ошибки из фиксированного набора
requests.WithLabelValues(r.Pattern).Inc()
failures.WithLabelValues("timeout").Inc() // timeout, conflict, validation...
```

Подробности конкретного запроса (id, текст ошибки) пиши в лог или в атрибуты спана. Метрика отвечает на «сколько», лог на «что именно».
:::

### Middleware для RED-метрик

Код ответа хендлер сообщает через `WriteHeader`, а middleware его не видит. Решение: обернуть `http.ResponseWriter` и запомнить код.

```go
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

func (s *statusRecorder) Unwrap() http.ResponseWriter { return s.ResponseWriter }
```

Тонкости, на которых спотыкаются:

- Хендлер может вообще не вызвать `WriteHeader`, а сразу `Write`. Тогда код 200, и обёртка должна это учесть (в `Write` тоже).
- Хендлер может сначала вызвать `Write`, а потом `WriteHeader(500)`. Второй вызов `net/http` проигнорирует (клиент уже получил 200 и пишет в лог `superfluous response.WriteHeader call`), и метрика должна записать 200, а не 500.

::: trap Обёртка прячет Flusher
```go
// плохо: statusRecorder встраивает http.ResponseWriter, но не Flusher.
// SSE-хендлер за такой middleware перестаёт стримить:
f, ok := w.(http.Flusher) // ok == false, приведение к обёртке провалилось
```

```go
// хорошо: обёртка отдаёт исходный writer, а хендлер идёт через ResponseController
func (s *statusRecorder) Unwrap() http.ResponseWriter { return s.ResponseWriter }

// в хендлере:
if err := http.NewResponseController(w).Flush(); err != nil { ... }
```

`http.ResponseController` (Go 1.20) сам проходит по цепочке `Unwrap` до writer, который умеет `Flush`, `Hijack` и дедлайны. Любая middleware-обёртка без `Unwrap` ломает стриминг, WebSocket и `SetWriteDeadline` для всего, что за ней.
:::

## Health и readiness

Разницу ты уже знаешь по [уроку «Сервер к продакшну»](/m06/l05-production-server): если провалилась **liveness** (`/healthz`), Kubernetes перезапускает контейнер, поэтому она проверяет только сам процесс; если провалилась **readiness** (`/readyz`), под убирают из балансировки без перезапуска, и здесь проверяют зависимости. Добавим то, что там осталось за кадром.

**Startup probe** нужна медленно стартующим сервисам (прогрев кеша, миграции). Пока она не прошла, liveness и readiness не проверяются, и сервис не убивают в цикле «не успел стартовать, перезапуск».

**Параметры проб.** По умолчанию kubelet дёргает пробу раз в 10 секунд (`periodSeconds`), ждёт ответа 1 секунду (`timeoutSeconds`) и считает пробу проваленной после 3 неудач подряд (`failureThreshold`). Отсюда требования к `/readyz`:

- проверки зависимостей идут **параллельно** и с **общим таймаутом** меньше секунды: три последовательные проверки по 400 мс дадут «не готов» при живых зависимостях;
- проверка, которая повисла и игнорирует контекст, не должна задерживать ответ: обработчик отвечает по таймауту, а зависшая горутина допишет результат в буферизованный канал и завершится, не утекая;
- ответ показывает, что именно сломалось (`{"status":"fail","checks":{"db":"ok","redis":"timeout"}}`): так дежурный за секунду видит причину.

Это вторая задача домашки.

::: trap Необязательная зависимость в readiness
```go
// плохо: Redis здесь только кеш, но его падение выключает весь сервис:
// все поды разом «не готовы», балансировщику некуда слать трафик
Readiness(map[string]Check{"db": pool.Ping, "redis": redisPing}, 800*time.Millisecond)
```

```go
// хорошо: в readiness только то, без чего запрос гарантированно упадёт;
// состояние кеша видно по метрике, а сервис ходит в базу напрямую
Readiness(map[string]Check{"db": pool.Ping}, 800*time.Millisecond)
```

Та же логика, по которой базу не проверяют в liveness: проверка должна выключать под только тогда, когда выключение помогает. Если сервис умеет деградировать без зависимости (кеш, рекомендации, аналитика), её падение не повод уводить трафик.
:::

## Graceful shutdown с drain

Базовую остановку ты написал в [уроке «Сервер к продакшну»](/m06/l05-production-server): `signal.NotifyContext`, `Shutdown` с отдельным контекстом от `context.Background()` (контекст сигнала уже отменён), `http.ErrServerClosed` как штатное завершение и порядок остановки в Kubernetes. Здесь доводим её до функции, которую можно проверить тестом: `Serve(ctx, srv, ln, ready, drain, timeout)` из третьей задачи домашки. Флаг готовности это `atomic.Bool`, который читает `/readyz`:

```go
var ready atomic.Bool
mux.HandleFunc("GET /readyz", func(w http.ResponseWriter, r *http.Request) {
	if !ready.Load() {
		http.Error(w, "draining", http.StatusServiceUnavailable)
		return
	}
	// ... проверки зависимостей
})
```

::: hood Почему сначала readiness, потом Shutdown
Kubernetes отправляет SIGTERM и одновременно убирает под из Endpoints сервиса. Но это распространяется асинхронно: kube-proxy на каждой ноде, ingress-контроллер, внешний балансировщик узнают об этом с задержкой от долей секунды до нескольких секунд. Всё это время на под ещё приходят новые соединения. Если сервис сразу сделал `Shutdown` и закрыл порт, эти запросы получат `connection refused`.

```text
SIGTERM
  │
  ├─ ready = false         /readyz отвечает 503
  ├─ пауза drain (5-10 с)  трафик ещё принимаем, балансировщики «догоняют»
  ├─ srv.Shutdown(ctx)     новых не принимаем, ждём активные
  ├─ закрыть пул БД, сбросить буферы логов и трейсов
  └─ exit 0

drain + timeout Shutdown < terminationGracePeriodSeconds
```

Сумма паузы и таймаута Shutdown должна укладываться в grace period (30 секунд по умолчанию), иначе SIGKILL прилетит посреди остановки.

Как такое тестировать без настоящей сети и без `time.Sleep` в тесте: listener в памяти поверх `net.Pipe` и весь тест внутри `testing/synctest`. Тогда пауза drain идёт по фейковым часам, а запрос «посреди drain» гарантированно попадает в нужное окно. Так устроены тесты домашки.
:::

::: trap Сервер упал, а процесс живёт
```go
// плохо: если порт занят, ListenAndServe сразу вернёт ошибку, её никто
// не прочитает, а main будет ждать сигнала: процесс жив, но не обслуживает
go srv.ListenAndServe()
<-ctx.Done()
```

```go
// хорошо: ждём либо сигнал, либо падение сервера
errc := make(chan error, 1)
go func() { errc <- srv.Serve(ln) }()
select {
case err := <-errc:
	ready.Store(false)
	return err
case <-ctx.Done():
}
```

Без второй ветки сервис с занятым портом проходит liveness по процессу, но ни на что не отвечает. В тестах домашки это отдельный кейс: закрытый listener.
:::

## Трейсы и OpenTelemetry

Запрос «оформить заказ» идёт в gateway, оттуда в сервис заказов, тот ходит в базу, в сервис оплаты и кладёт задачу в очередь. Запрос занял 3 секунды. Где именно? Логи каждого сервиса по отдельности не ответят, метрики покажут средние. Ответит **распределённый трейс**.

- **Трейс** (trace) это путь одного запроса через все сервисы, у него один `trace_id`.
- **Спан** (span) это одна операция внутри трейса: обработка HTTP-запроса, SQL-запрос, вызов другого сервиса. У спана есть имя, время начала и конца, атрибуты, статус и **родитель**. Спаны складываются в дерево, которое UI (Jaeger, Grafana Tempo) рисует как диаграмму Ганта.

```text
trace 4463...a9
├─ POST /orders                      gateway     3020 ms
│  └─ createOrder                    orders      2990 ms
│     ├─ db.insert order             postgres      12 ms
│     ├─ POST /payments              payments    2950 ms  ← вот оно
│     │  └─ bank.authorize           payments    2940 ms
│     └─ queue.publish               river          4 ms
```

**OpenTelemetry** (OTel) это вендор-нейтральный стандарт и набор SDK для всех трёх сигналов. Код инструментируется один раз через API OTel, а куда отправлять данные (Jaeger, Tempo, облачный APM) решается конфигурацией экспортёра. В Go это отдельные модули `go.opentelemetry.io/otel/...`, в домашку курса они не входят, поэтому ниже обзор.

```go
var tracer = otel.Tracer("shop/orders")

func createOrder(ctx context.Context, userID int) error {
	ctx, span := tracer.Start(ctx, "createOrder")
	defer span.End()
	span.SetAttributes(attribute.Int("user.id", userID))

	if err := saveToDB(ctx); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "save failed")
		return err
	}
	return nil
}

func saveToDB(ctx context.Context) error {
	_, span := tracer.Start(ctx, "db.insert order") // дочерний спан
	defer span.End()
	// ...
}
```

Ключевая идея: **текущий спан живёт в `context.Context`**. `tracer.Start(ctx, ...)` берёт родителя из `ctx` и возвращает новый `ctx` с дочерним спаном. Поэтому в Go контекст передают первым аргументом через все слои: без него трейс рвётся.

Между сервисами контекст едет в заголовке **`traceparent`** (стандарт W3C Trace Context): версия, trace id, id родительского спана и флаги. Готовые обёртки делают это сами: `otelhttp.NewHandler` на сервере достаёт заголовок и начинает спан, `otelhttp.NewTransport` на клиенте добавляет заголовок к исходящим запросам, для gRPC, pgx и Redis есть аналогичные инструментации.

Связка сигналов: если писать `trace_id` в каждую запись лога (`TraceHandler` из раздела про логи), то из трейса можно перейти к логам запроса, а из всплеска на графике метрики (через exemplars) к конкретному медленному трейсу.

Трейсить каждый запрос дорого, поэтому применяют **сэмплирование**: сохраняют, например, 1% трейсов или все трейсы с ошибками.

## Что выведет программа?

```go
c := prometheus.NewCounter(prometheus.CounterOpts{Name: "jobs_total", Help: "h"})
defer func() { fmt.Println("recovered:", recover()) }()
c.Add(-1)
```

<Predict :options="['ничего, значение станет -1', 'recovered: <nil>', 'recovered: counter cannot decrease in value', 'ошибка компиляции']" :correct="2">

Counter по определению только растёт, и `Add` с отрицательным аргументом паникует. Если значение должно уменьшаться (размер очереди, активные соединения), нужен Gauge. Если хочется «счётчик, который можно сбросить», тоже нет: сбросы `rate` обработает сам при рестарте процесса, а вручную счётчики не обнуляют.

</Predict>

```go
mux := http.NewServeMux()
mux.Handle("GET /users/{id}", metrics.Middleware(h))
// запросы: GET /users/1, GET /users/2, GET /users/3
// middleware пишет лейбл route = r.Pattern
```

<Predict :options="['3 ряда http_requests_total, по одному на пользователя', '1 ряд со значением 3', '1 ряд со значением 1', '0 рядов: r.Pattern пуст внутри middleware']" :correct="1">

Middleware обёрнута вокруг хендлера, зарегистрированного в `ServeMux`, поэтому к моменту её вызова mux уже сопоставил шаблон и записал его в `r.Pattern`: `"GET /users/{id}"` для всех трёх запросов. Один ряд, значение 3. Если бы middleware обернула весь mux снаружи, `r.Pattern` до вызова `next` был бы пуст.

</Predict>

::: interview
<details><summary>Чем логи, метрики и трейсы отличаются друг от друга?</summary>

Логи это отдельные события с подробностями, отвечают на «что произошло с этим запросом», дороги в хранении. Метрики это агрегированные числа во времени (счётчики, гистограммы), дёшевы, по ним строят графики и алерты, отвечают на «как система ведёт себя в целом». Трейсы показывают путь одного запроса через сервисы как дерево спанов с длительностями, отвечают на «где ушло время». Связываются через общий trace id.

</details>
<details><summary>Какие типы метрик есть в Prometheus? Когда histogram, а когда summary?</summary>

Counter (только растёт, смотрят через `rate`), Gauge (произвольное текущее значение), Histogram (наблюдения по кумулятивным бакетам плюс `_sum` и `_count`), Summary (перцентили считаются в процессе). Для задержек почти всегда histogram: бакеты можно складывать между подами и считать перцентиль всего сервиса через `histogram_quantile`. Перцентили summary агрегировать нельзя, усреднение p99 по подам даёт бессмыслицу.

</details>
<details><summary>Что такое кардинальность и почему это важно?</summary>

Число уникальных временных рядов: каждая комбинация значений лейблов это отдельный ряд в памяти Prometheus. Лейбл с неограниченными значениями (user id, URL с параметрами, текст ошибки) взрывает число рядов, Prometheus съедает память и падает. Лейблы берут только из маленьких фиксированных множеств: метод, шаблон маршрута, код ответа.

</details>
<details><summary>Что такое RED и USE?</summary>

RED для сервисов: Rate (запросы в секунду), Errors (доля ошибок), Duration (распределение задержек). USE для ресурсов: Utilization (занятость), Saturation (очередь ожидающих), Errors. Вместе дают минимальный набор метрик, по которому видно и поведение сервиса для клиентов, и то, во что он упирается.

</details>
<details><summary>Чем liveness отличается от readiness? Что будет, если проверять базу в liveness?</summary>

Liveness проваливается, когда процесс завис, и Kubernetes перезапускает контейнер. Readiness проваливается, когда сервис временно не готов, и под убирают из балансировки без перезапуска. Если проверять базу в liveness, падение базы вызовет перезапуск всех подов разом, что ничего не чинит и создаёт шквал подключений при восстановлении. Зависимости проверяют в readiness, параллельно и с таймаутом.

</details>
<details><summary>Как сделать graceful shutdown HTTP-сервера в Go?</summary>

Поймать SIGTERM через `signal.NotifyContext`, перевести readiness в 503 и подождать несколько секунд, пока балансировщики уберут под. Затем `srv.Shutdown` с отдельным контекстом-таймаутом от `context.Background()` (контекст сигнала уже отменён): он закрывает слушающие сокеты и ждёт активные запросы. После этого закрыть пул базы и сбросить буферы. `http.ErrServerClosed` из `ListenAndServe` после `Shutdown` это норма. Пауза плюс таймаут должны укладываться в `terminationGracePeriodSeconds`.

</details>
<details><summary>Как передаётся контекст трейса между сервисами?</summary>

Внутри процесса текущий спан лежит в `context.Context`, поэтому ctx передают через все слои. Между сервисами контекст сериализуется в заголовок `traceparent` по стандарту W3C Trace Context (trace id, id родительского спана, флаги сэмплирования). Инструментация OpenTelemetry (`otelhttp`, gRPC-интерцепторы) извлекает его на входе и добавляет на выходе.

</details>
<details><summary>Как сделать, чтобы все логи одного запроса можно было найти вместе?</summary>

Middleware берёт `X-Request-ID` из заголовка или генерирует новый, кладёт в `context` и возвращает в ответе. Логи пишутся через методы с контекстом (`InfoContext`), а обёртка над `slog.Handler` достаёт id из контекста и добавляет атрибутом. Если есть трейсинг, туда же добавляют trace id, чтобы связать логи с трейсом.

</details>
:::

## Источники

- `RU` [Структурированное логирование в Go с помощью Slog](https://habr.com/ru/companies/slurm/articles/798207/): подробный перевод гайда по slog, хендлеры, уровни, `LogValuer`, контекст.
- `RU` [Как работает гистограмма Prometheus?](https://habr.com/ru/companies/otus/articles/501978/): бакеты, `histogram_quantile` и откуда берётся погрешность.
- `RU` [Трейсинг в Go: это просто](https://habr.com/ru/articles/710644/): OpenTelemetry и Jaeger на маленьком сервисе, спаны и экспорт.
- `EN` [Structured Logging with slog](https://go.dev/blog/slog): пост команды Go о дизайне пакета.
- `EN` [Metric and label naming](https://prometheus.io/docs/practices/naming/): как называть метрики, единицы, суффиксы `_total` и `_seconds`, предупреждение про кардинальность.
- `EN` [Histograms and summaries](https://prometheus.io/docs/practices/histograms/): когда что выбирать и почему summary не агрегируется.
- `EN` [The RED Method](https://grafana.com/blog/2018/08/02/the-red-method-how-to-instrument-your-services/): статья автора метода.
- `EN` [Liveness, Readiness, and Startup Probes](https://kubernetes.io/docs/concepts/configuration/liveness-readiness-startup-probes/): что делает Kubernetes с каждой проверкой.

## Домашка

<Hw id="ex1" title="RED-метрики в middleware" level="база">

Файл `homework/m10/l02-observability/ex1_metrics.go`: `NewMetrics(reg)` и `Middleware`. Счётчик с лейблами `method`, `route`, `code`, гистограмма длительности, gauge активных запросов. Тесты ловят лейбл из `r.URL.Path` вместо шаблона, регистрацию в глобальном реестре, потерянный код 200, длительность в миллисекундах и обёртку без `Unwrap`.

```bash
cd homework && go test -v ./m10/l02-observability/ -run TestMetrics
```

</Hw>

<Hw id="ex2" title="Readiness с зависимостями" level="посложнее">

Файл `homework/m10/l02-observability/ex2_readiness.go`, функция `Readiness(checks, timeout)`. Проверки параллельно, общий таймаут, JSON с результатом каждой, 200 или 503. Проверка, которая игнорирует контекст и висит, не должна задерживать ответ и не должна оставлять заблокированную горутину навсегда. Тесты идут в `synctest`, время там фейковое.

```bash
cd homework && go test -v ./m10/l02-observability/ -run TestReadiness
```

</Hw>

<Hw id="ex3" title="Graceful shutdown с drain" level="посложнее">

Файл `homework/m10/l02-observability/ex3_serve.go`, функция `Serve(ctx, srv, ln, ready, drain, timeout)`: флаг готовности, пауза перед `Shutdown`, ожидание активных запросов, ошибка при нехватке времени и при падении самого сервера. Тесты идут в `synctest` с listener в памяти поверх `net.Pipe`: ни портов, ни реального ожидания, длительности проверяются по фейковым часам.

```bash
cd homework && go test -race -v ./m10/l02-observability/ -run TestServe
```

</Hw>

<Hw id="ex4" title="Собери всё в сервис" level="со звёздочкой">

Вне папки домашки (свой модуль) собери сервис из трёх решений: `RequestID`-middleware и `ContextHandler` из урока, метрики на всех маршрутах, `/metrics` через `promhttp.HandlerFor` со своим реестром и Go-коллектором, `/healthz`, `/readyz` и `Serve` с обработкой SIGTERM. Проверь руками:

```bash
curl -i localhost:8080/api/ping            # в ответе X-Request-ID, в логе тот же id
curl -s localhost:8080/metrics | grep http_
kill -TERM <pid>                           # /readyz сразу 503, через drain процесс выходит
```

Подними Prometheus в Docker (`prom/prometheus`, конфиг с `scrape_configs` на твой сервис через `host.docker.internal:8080`) и построй в его UI `rate(http_requests_total[1m])` и `histogram_quantile(0.95, sum by (le) (rate(http_request_duration_seconds_bucket[1m])))`.

</Hw>
