# context: отмена, дедлайны и значения

Представь обычный бэкенд-запрос: клиент дёрнул `GET /orders`, хендлер позвал сервис, сервис пошёл в PostgreSQL и параллельно в соседний микросервис за ценами. Через 200 мс клиент закрыл вкладку. Если ничего не сделать, сервер честно доделает работу: дождётся базы, дождётся соседа, соберёт JSON и отправит его в закрытое соединение. Под нагрузкой такие «зомби-запросы» съедают пул соединений к базе и горутины, и сервис падает не от полезной работы, а от мусора.

Пакет `context` решает ровно эту задачу: протаскивает через все слои программы **сигнал «хватит»** (отмена или истёкший дедлайн) и немного данных, привязанных к запросу. В Go это стандарт де-факто: `net/http`, `database/sql`, pgx, gRPC-клиенты, Redis-клиенты принимают `ctx` первым аргументом. На собесе по Go вопрос «зачем нужен context и как работает отмена» звучит почти всегда, а «что можно класть в context.Value» любят как вопрос на вкус и зрелость.

В прошлых уроках ты гасил горутины через `done`-канал и `select`. Context это тот же `done`-канал, только стандартизированный, вложенный деревом и с дедлайнами.

## Интерфейс Context

Весь контракт в четырёх методах:

```go
type Context interface {
	Deadline() (deadline time.Time, ok bool) // когда контекст истечёт, если дедлайн есть
	Done() <-chan struct{}                   // закроется при отмене или дедлайне
	Err() error                              // nil, пока не отменён; потом Canceled или DeadlineExceeded
	Value(key any) any                       // значение по ключу или nil
}
```

Главный метод `Done()`. Это канал, который **закрывается** в момент отмены. Закрытый канал читается мгновенно всеми, кто его ждёт, поэтому одна отмена будит сколько угодно горутин сразу. Ровно тот приём с `close(done)`, который ты уже видел, только спрятанный за интерфейсом.

`Err()` объясняет, почему закрылся `Done()`:

- `nil`: контекст ещё жив;
- `context.Canceled`: кто-то вызвал `cancel()`;
- `context.DeadlineExceeded`: истёк дедлайн или таймаут.

Сравнивай через `errors.Is(err, context.Canceled)`, а не `==`: библиотеки часто оборачивают эти ошибки (например, `*url.Error` из `http.Client`).

### Корни дерева: Background и TODO

Контексты создаются только от родителя. Корней два:

- `context.Background()`: пустой контекст, никогда не отменяется, без значений и дедлайна. Им начинают `main`, тесты, фоновые задачи.
- `context.TODO()`: технически то же самое, но читается как «я ещё не решил, откуда тут брать контекст». Удобно при рефакторинге: линтеры и ревьюер видят, что место недоделано.

В тестах с Go 1.24 есть `t.Context()`: контекст, который отменяется после завершения теста, прямо перед запуском `Cleanup`-функций.

## Дерево контекстов

Каждая функция `With*` берёт родителя и возвращает **производный** контекст:

```go
ctx, cancel := context.WithCancel(parent)        // отменяется вручную
ctx, cancel := context.WithTimeout(parent, 2*time.Second)
ctx, cancel := context.WithDeadline(parent, t)   // то же, но абсолютным временем
ctx := context.WithValue(parent, key, val)       // добавляет значение
```

Получается дерево. Правило одно, и его спрашивают на собесе: **отмена идёт только вниз**. Отменили родителя, умерли все потомки. Отменили потомка, родитель и братья живут дальше.

```text
Background
 └── WithTimeout(5s)          ← запрос целиком
      ├── WithCancel          ← поход в базу
      └── WithTimeout(1s)     ← поход в сервис цен
```

Если запрос отменили, умирают и база, и сервис цен. Если сервис цен не уложился в свою секунду, умирает только его ветка, а запрос может, например, отдать ответ без цен.

```go
parent, cancelP := context.WithCancel(context.Background())
child, cancelC := context.WithCancel(parent)
cancelC()
fmt.Println(parent.Err(), child.Err())
cancelP()
fmt.Println(parent.Err(), child.Err())
```

<Predict :options="['<nil> context canceled, затем context canceled context canceled', 'context canceled context canceled в обеих строках', '<nil> <nil>, затем context canceled context canceled', 'паника: повторная отмена']" :correct="0">

Отмена потомка не трогает родителя: в первой строке `<nil> context canceled`. Отмена родителя доходит до всех потомков, а повторная отмена уже мёртвого `child` ни на что не влияет и не паникует: `cancel` идемпотентна. Вторая строка: `context canceled context canceled`.

</Predict>

### cancel вызывать обязательно

`WithCancel`, `WithTimeout` и `WithDeadline` возвращают функцию `cancel`, и её **нужно вызвать всегда**, даже если работа закончилась успешно. Идиома:

```go
ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
defer cancel()
```

Почему это важно, видно под капотом ниже: дочерний контекст регистрирует себя в родителе, а у `WithTimeout` ещё и заводится таймер. Пока не вызван `cancel` (или не истёк таймер, или не отменён родитель), всё это висит в памяти. `go vet` ловит забытый `cancel` проверкой `lostcancel`.

## Как функция слушает отмену

Context сам ничего не прерывает. Он только сообщает. Прерваться должна твоя функция, и способа два.

**Ждёшь чего-то: добавь `ctx.Done()` в `select`.**

```go
func waitResult(ctx context.Context, results <-chan Result) (Result, error) {
	select {
	case r := <-results:
		return r, nil
	case <-ctx.Done():
		return Result{}, ctx.Err()
	}
}
```

**Крутишь цикл: проверяй `ctx.Err()` между итерациями.**

```go
func processAll(ctx context.Context, items []Item) error {
	for _, it := range items {
		if err := ctx.Err(); err != nil {
			return err // отменили: дальше не идём
		}
		process(it)
	}
	return nil
}
```

Проверка дешёвая (атомарное чтение), но не делай её на каждой из миллиарда итераций тесного цикла: раз в N итераций достаточно.

Библиотечные функции, принимающие `ctx`, делают это за тебя: `db.QueryContext(ctx, ...)` прервёт запрос в базе, `http.NewRequestWithContext(ctx, ...)` закроет соединение. Поэтому правило простое: **если функция принимает ctx, передай ей свой**, а не `context.Background()`.

## Дедлайны и таймауты

`WithTimeout(parent, d)` это просто `WithDeadline(parent, time.Now().Add(d))`. Таймаут удобен для «не дольше 2 секунд», дедлайн для «до 18:00».

Ключевое правило: **дедлайн потомка не может быть позже дедлайна родителя**. Если у родителя осталась секунда, а ты просишь у потомка час, потомок получит ту же секунду. Это защищает от ситуации, когда внутренний слой «щедро» даёт себе больше времени, чем есть у запроса.

`ctx.Deadline()` позволяет заранее понять, есть ли смысл начинать:

```go
if dl, ok := ctx.Deadline(); ok && time.Until(dl) < 50*time.Millisecond {
	return ErrNotEnoughTime // тяжёлый запрос всё равно не успеет
}
```

Типичная схема в сервисе: таймаут на весь запрос ставится один раз (в middleware или на сервере), а внутренние вызовы режут себе кусок поменьше:

```go
func (s *Service) Order(ctx context.Context, id int64) (Order, error) {
	o, err := s.repo.Get(ctx, id) // общий дедлайн запроса
	if err != nil {
		return Order{}, err
	}
	pctx, cancel := context.WithTimeout(ctx, 300*time.Millisecond)
	defer cancel()
	price, err := s.prices.Get(pctx, id) // цены не дольше 300 мс
	if err != nil {
		o.PriceUnavailable = true // деградируем, а не падаем
		return o, nil
	}
	o.Price = price
	return o, nil
}
```

## Причина отмены: WithCancelCause и Cause

`ctx.Err()` говорит только «отменили» или «время вышло». В логах это бесполезно: **кто** отменил и **почему**? С Go 1.20 есть `WithCancelCause`, с 1.21 `WithTimeoutCause` и `WithDeadlineCause`:

```go
ctx, cancel := context.WithCancelCause(parent)
cancel(ErrShutdown) // причина

ctx.Err()            // context.Canceled, как и раньше (совместимость)
context.Cause(ctx)   // ErrShutdown
```

```go
ctx, cancel := context.WithTimeoutCause(parent, time.Second, ErrPaymentTimeout)
defer cancel()
// когда истечёт: ctx.Err() == DeadlineExceeded, context.Cause(ctx) == ErrPaymentTimeout
```

Правила `Cause`:

- до отмены возвращает `nil`;
- причину ставит **первая** отмена контекста или любого его предка, повторные вызовы игнорируются;
- если отменили обычным `cancel()` без причины, `Cause` вернёт то же, что `Err()`.

Отсюда удобный приём: функция возвращает наружу `context.Cause(ctx)` вместо `ctx.Err()`, и вызывающий видит настоящую причину. Если отменили родителя с причиной `ErrShutdown`, наружу уйдёт `ErrShutdown`; если сработал собственный таймаут, уйдёт его причина. Это задача 4 домашки. С Go 1.26 так делает и `signal.NotifyContext`: по причине отмены видно, какой сигнал пришёл.

## AfterFunc: «когда отменят, сделай вот это»

`context.AfterFunc(ctx, f)` (Go 1.21) регистрирует функцию, которая запустится **в отдельной горутине** после отмены `ctx`. Если `ctx` уже отменён, `f` стартует сразу. Возвращается `stop`: вызов `stop()` снимает регистрацию и возвращает `true`, если успел до запуска `f`.

Зачем это, если есть `select`? Не всё в мире умеет `select`. Блокирующее чтение из `net.Conn`, ожидание на `sync.Cond`, сторонний API без контекста. AfterFunc позволяет «прикрутить» отмену снаружи:

```go
func readWithContext(ctx context.Context, conn net.Conn, buf []byte) (int, error) {
	stop := context.AfterFunc(ctx, func() {
		// Дедлайн в прошлом немедленно прерывает висящий Read.
		conn.SetReadDeadline(time.Now())
	})
	defer stop()
	n, err := conn.Read(buf)
	if ctx.Err() != nil {
		return n, ctx.Err()
	}
	return n, err
}
```

Важная деталь: для стандартных контекстов `AfterFunc` не держит горутину на ожидание. Функция записывается в список потомков контекста и запускается при отмене. Поэтому тысяча `AfterFunc` на живом контексте почти ничего не стоят, а `stop()` полностью убирает следы.

## WithoutCancel: работа, которая переживает запрос

Классика: хендлер ответил клиенту и хочет в фоне записать аудит или отправить письмо. Наивно:

```go
go s.audit.Write(r.Context(), event) // плохо
```

Контекст запроса отменяется, как только хендлер вернулся. Фоновая запись получит `context canceled` в первом же обращении к базе. Брать `context.Background()` тоже плохо: потеряются значения запроса (request id, трассировка), по которым потом ищут логи.

`context.WithoutCancel(parent)` (Go 1.21) возвращает контекст, который **сохраняет значения** родителя, но **не наследует отмену и дедлайн**: `Done()` возвращает `nil`, `Err()` всегда `nil`, `Deadline()` пуст. Фоновой работе почти всегда нужен свой таймаут, иначе она может повиснуть навсегда:

```go
bg, cancel := context.WithTimeout(context.WithoutCancel(r.Context()), 5*time.Second)
go func() {
	defer cancel()
	s.audit.Write(bg, event)
}()
```

## Значения: WithValue

`context.WithValue(parent, key, val)` кладёт одну пару ключ-значение. Поиск `ctx.Value(key)` идёт вверх по цепочке родителей, пока не найдёт совпадающий ключ.

Ключи сравниваются через `==`, включая **тип**. Поэтому ключ делают своим неэкспортируемым типом: никакой другой пакет не сможет случайно (или намеренно) создать такой же ключ и перетереть твоё значение.

```go
type ctxKey struct{} // или: type ctxKey int; const requestIDKey ctxKey = 0

func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, ctxKey{}, id)
}

func RequestID(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(ctxKey{}).(string)
	return id, ok
}
```

Ключ-строка `"request_id"` это плохо: любой пакет может положить своё значение с тем же ключом. `go vet` такое не ловит, а `staticcheck` ругается (SA1029).

```go
type key string

ctx := context.WithValue(context.Background(), key("id"), 42)
fmt.Println(ctx.Value("id"), ctx.Value(key("id")))
```

<Predict :options="['42 42', '<nil> 42', '42 <nil>', 'ошибка компиляции']" :correct="1">

`"id"` это нетипизированная строковая константа, при передаче в `any` она становится `string`. Ключ в контексте имеет тип `key`. Значения одинаковые, типы разные, `==` для `any` сравнивает и тип, и значение. Поэтому первый поиск даёт `<nil>`, второй `42`. Ровно поэтому свой тип ключа и защищает от коллизий.

</Predict>

### Что класть в context, а что нет

Это вопрос на зрелость. Правило из документации: в context кладут **данные, привязанные к запросу** (request-scoped), которые проходят через границы API и процессов, а не **опциональные параметры функций**.

Можно:

- идентификатор запроса и трассировки (trace id, span);
- аутентифицированного пользователя или его права, извлечённые middleware;
- локаль, если она пришла с запросом.

Нельзя:

- подключение к базе, логгер, конфиг: это зависимости, их передают в конструктор структуры;
- параметры бизнес-логики (`limit`, `userID` для выборки): это аргументы функции;
- всё, без чего функция не может работать. Значение в контексте **невидимо** в сигнатуре: компилятор не проверит, что его положили, а читатель кода не узнает, что оно нужно.

Тест на здравый смысл: если функция сломается, когда значения в контексте нет, значит, это должен быть явный параметр.

## Передача через слои

Соглашения, которые проверяют на ревью:

1. `ctx` всегда **первый параметр** и называется `ctx`: `func (r *Repo) Get(ctx context.Context, id int64) (User, error)`.
2. **Не храни context в структуре.** Контекст живёт столько же, сколько операция, а структура обычно дольше. Исключение: тип, который сам представляет одну операцию (как `http.Request`).
3. Не передавай `nil`. Не знаешь, что передать: `context.TODO()`.
4. В HTTP-сервере контекст запроса берут из `r.Context()`. Он отменяется, когда клиент закрыл соединение или хендлер вернулся.

Сквозной пример трёх слоёв:

```go
// handler: берём контекст запроса, ставим общий таймаут
func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	u, err := h.svc.User(ctx, r.PathValue("id"))
	if errors.Is(err, context.DeadlineExceeded) {
		http.Error(w, "timeout", http.StatusGatewayTimeout)
		return
	}
	// ...
}

// service: просто передаёт ctx дальше
func (s *Service) User(ctx context.Context, id string) (User, error) {
	return s.repo.ByID(ctx, id)
}

// repository: отдаёт ctx драйверу, тот прервёт SQL-запрос
func (r *Repo) ByID(ctx context.Context, id string) (User, error) {
	var u User
	err := r.db.QueryRowContext(ctx, `SELECT id, name FROM users WHERE id = $1`, id).Scan(&u.ID, &u.Name)
	return u, err
}
```

Ни один слой не создаёт `context.Background()` посреди цепочки. Если где-то он появился, отмена в этом месте рвётся, и ниже по стеку запрос станет «вечным».

::: hood Как устроен context внутри
Исходник `src/context/context.go` небольшой и читается за вечер. Основные типы:

```text
emptyCtx    Background и TODO: Done() возвращает nil, никогда не отменяется
cancelCtx   WithCancel:  mu, done (канал, создаётся лениво), children, err, cause
timerCtx    WithTimeout: cancelCtx + time.Timer + deadline
valueCtx    WithValue:   родитель + одна пара key, val
```

**Отмена.** Когда создаётся `cancelCtx`, функция `propagateCancel` смотрит на родителя:

1. Родитель никогда не отменяется (`Done() == nil`, как у Background): регистрироваться не нужно.
2. Родитель уже отменён: потомок отменяется сразу.
3. Родитель сам `cancelCtx` (или стандартный контекст поверх него): потомок записывается в `parent.children`. При отмене родитель под мьютексом обходит `children` и отменяет каждого, рекурсивно вниз.
4. Родитель чужого типа, но с методом `AfterFunc`: используется он.
5. Иначе запускается горутина, которая ждёт `parent.Done()` или `child.Done()`. Это единственный случай, когда context сам тратит горутину.

Отсюда понятно, зачем вызывать `cancel`: он удаляет потомка из `parent.children` и останавливает таймер. Забытый `cancel` у долгоживущего родителя это настоящая утечка памяти.

**Канал Done** создаётся лениво при первом вызове `Done()`. Если отменили раньше, чем кто-то спросил канал, подставляется общий заранее закрытый канал `closedchan`. Мелочь, но экономит аллокации.

**Значения** хранятся связным списком: каждый `WithValue` это новый узел с указателем на родителя. `Value` идёт вверх по цепочке, то есть поиск O(глубина). Это нормально для 3-5 значений, но превращать контекст в мапу на сотню ключей не стоит.
:::

::: trap Забытый cancel
```go
// Плохо: таймер и регистрация в родителе живут до истечения 10 секунд,
// даже если запрос в базу завершился за 5 мс. На 1000 RPS это 10 000 висящих таймеров.
func load(ctx context.Context) error {
	ctx, _ = context.WithTimeout(ctx, 10*time.Second)
	return db.PingContext(ctx)
}
```

```go
// Хорошо
func load(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	return db.PingContext(ctx)
}
```
`go vet` найдёт такую ошибку. Не отключай проверку `lostcancel`.
:::

::: trap Фоновая горутина на контексте запроса
```go
// Плохо: хендлер вернулся, r.Context() отменён, письмо не уйдёт
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	// ...
	go h.mailer.SendWelcome(r.Context(), user)
	w.WriteHeader(http.StatusCreated)
}
```

```go
// Хорошо: значения сохранены, отмена отвязана, свой таймаут
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	// ...
	ctx, cancel := context.WithTimeout(context.WithoutCancel(r.Context()), 10*time.Second)
	go func() {
		defer cancel()
		h.mailer.SendWelcome(ctx, user)
	}()
	w.WriteHeader(http.StatusCreated)
}
```
В настоящем сервисе важные фоновые задачи (письма, вебхуки) вообще лучше ставить в очередь задач: горутина умрёт вместе с процессом при деплое. Это тема модуля про фоновые задачи.
:::

::: trap Функция «принимает» ctx, но не слушает его
```go
// Плохо: ctx в сигнатуре для красоты, отмена ни на что не влияет
func Poll(ctx context.Context, check func() bool) {
	for !check() {
		time.Sleep(time.Second)
	}
}
```

```go
// Хорошо: и ожидание, и выход завязаны на ctx
func Poll(ctx context.Context, check func() bool) error {
	t := time.NewTicker(time.Second)
	defer t.Stop()
	for !check() {
		select {
		case <-t.C:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return nil
}
```
:::

::: trap context.Value вместо параметров
```go
// Плохо: зависимость спрятана, сигнатура врёт
func ListOrders(ctx context.Context) ([]Order, error) {
	db := ctx.Value("db").(*sql.DB)        // паника, если забыли положить
	limit := ctx.Value("limit").(int)
	// ...
}
```

```go
// Хорошо: зависимость в структуре, параметры в аргументах
func (r *OrderRepo) List(ctx context.Context, limit int) ([]Order, error) {
	// r.db задан в конструкторе
}
```
:::

## Что выведет программа?

```go
parent, cancel1 := context.WithTimeout(context.Background(), time.Second)
defer cancel1()
child, cancel2 := context.WithTimeout(parent, time.Hour)
defer cancel2()

d1, _ := parent.Deadline()
d2, _ := child.Deadline()
fmt.Println(d1.Equal(d2))
```

<Predict :options="['true', 'false', 'паника: дедлайн потомка позже родителя']" :correct="0">

Дедлайн потомка не бывает позже дедлайна родителя. `WithDeadline` видит, что у родителя срок раньше, и вместо своего таймера просто создаёт отменяемый контекст: его `Deadline()` берётся у родителя. Поэтому `true`. Попросить у потомка больше времени, чем есть у родителя, нельзя.

</Predict>

```go
ctx, cancel := context.WithCancelCause(context.Background())
cancel(errors.New("диск переполнен"))
cancel(errors.New("второй вызов"))
fmt.Println(ctx.Err())
fmt.Println(context.Cause(ctx))
```

<Predict :options="['context canceled, затем диск переполнен', 'диск переполнен, затем диск переполнен', 'context canceled, затем второй вызов', 'паника: повторная отмена']" :correct="0">

`Err()` для совместимости всегда возвращает стандартную `context.Canceled`. Причину отдаёт `Cause`, и её ставит только **первая** отмена: второй вызов `cancel` ничего не меняет и не паникует. Вывод: `context canceled`, затем `диск переполнен`.

</Predict>

```go
parent, cancel := context.WithTimeout(context.Background(), time.Second)
bg := context.WithoutCancel(parent)
cancel()
_, hasDeadline := bg.Deadline()
fmt.Println(parent.Err(), bg.Err(), hasDeadline, bg.Done() == nil)
```

<Predict :options="['context canceled context canceled true false', 'context canceled <nil> false true', 'context canceled <nil> true false', '<nil> <nil> false true']" :correct="1">

`WithoutCancel` отвязывает и отмену, и дедлайн: `Err()` всегда `nil`, `Deadline()` пуст, а `Done()` возвращает `nil`-канал (чтение из него блокирует навсегда, то есть «никогда не отменюсь»). Значения родителя при этом доступны. Вывод: `context canceled <nil> false true`.

</Predict>

::: interview
<details><summary>Зачем нужен context.Context? Какие у него методы?</summary>

Чтобы передавать через цепочку вызовов и горутин сигнал отмены, дедлайн и данные, привязанные к запросу. Четыре метода: `Deadline()` (когда истечёт), `Done()` (канал, закрывается при отмене), `Err()` (`nil`, `Canceled` или `DeadlineExceeded`), `Value(key)` (значение по ключу). Сам контекст ничего не прерывает: код должен слушать `Done()` или проверять `Err()`.

</details>
<details><summary>Как распространяется отмена в дереве контекстов?</summary>

Только вниз. Отмена родителя отменяет всех потомков (через список `children` у `cancelCtx`). Отмена потомка не влияет ни на родителя, ни на соседние ветки. Дедлайн потомка не может быть позже дедлайна родителя: при попытке берётся родительский.

</details>
<details><summary>Почему cancel нужно вызывать, даже если операция успешно завершилась?</summary>

Дочерний контекст регистрируется в родителе (в мапе `children`), а `WithTimeout` ещё и держит таймер. `cancel` удаляет регистрацию и останавливает таймер. Без него ресурсы освободятся только при отмене родителя или истечении таймера, а у долгоживущего родителя это утечка. Идиома: `defer cancel()` сразу после создания. `go vet` проверяет это анализатором `lostcancel`.

</details>
<details><summary>Чем отличаются context.Canceled и context.DeadlineExceeded? Как узнать настоящую причину?</summary>

`Canceled` означает явный вызов `cancel()`, `DeadlineExceeded` означает истёкший дедлайн или таймаут. Сравнивать через `errors.Is`, потому что ошибки часто обёрнуты. Подробную причину дают `WithCancelCause` (1.20), `WithTimeoutCause` и `WithDeadlineCause` (1.21) и функция `context.Cause(ctx)`: она возвращает ошибку, переданную при первой отмене, а если причины нет, то же, что `Err()`.

</details>
<details><summary>Что можно класть в context.Value, а что нельзя?</summary>

Можно данные, привязанные к запросу и проходящие сквозь слои: request id, trace id, аутентифицированного пользователя. Нельзя зависимости (БД, логгер, конфиг) и параметры функций: их не видно в сигнатуре, компилятор не проверяет их наличие, тип теряется до `any`. Ключ должен быть неэкспортируемым собственным типом, чтобы не было коллизий с другими пакетами. Поиск значения идёт вверх по цепочке, это O(глубины).

</details>
<details><summary>Почему context не хранят в полях структуры?</summary>

Контекст описывает одну операцию и живёт столько же, сколько она. Структура (сервис, репозиторий) живёт дольше и обслуживает много операций. Контекст в поле путает, к какой операции относится отмена, и мешает вызывающему задать свой дедлайн. Передают его первым параметром каждого метода. Исключение: типы, которые сами представляют одну операцию, как `http.Request`.

</details>
<details><summary>Как сделать фоновую задачу после ответа HTTP-хендлера, чтобы её не убила отмена запроса?</summary>

Взять `context.WithoutCancel(r.Context())`: значения (request id, трассировка) сохранятся, а отмена и дедлайн нет. И обязательно навесить свой таймаут через `WithTimeout`, иначе задача может висеть вечно. Для действительно важной работы лучше очередь задач: горутина не переживёт рестарт процесса.

</details>
<details><summary>Что делает context.AfterFunc и когда он нужен?</summary>

Регистрирует функцию, которая запустится в отдельной горутине после отмены контекста (или сразу, если он уже отменён). Возвращает `stop`, который снимает регистрацию. Нужен, чтобы подружить с отменой то, что не умеет `select`: блокирующий `net.Conn.Read` (через `SetReadDeadline`), `sync.Cond.Wait`, сторонние API. Для стандартных контекстов не тратит горутину на ожидание.

</details>
<details><summary>Как HTTP-сервер узнаёт, что клиент ушёл?</summary>

`net/http` отменяет `r.Context()`, когда соединение с клиентом закрыто или хендлер вернулся (для HTTP/2 также при сбросе потока). Если хендлер передаёт этот контекст в базу и исходящие запросы, вся цепочка прерывается сама. Если где-то в середине написать `context.Background()`, отмена дальше не пройдёт.

</details>
:::

## Источники

- `RU` [Пакет context в Go: взгляд профессионала](https://habr.com/ru/companies/pt/articles/764850/): разбор от Positive Technologies, история пакета, правила передачи и `WithValue`.
- `RU` [Golang: context изнутри](https://habr.com/ru/articles/827544/): как устроены `cancelCtx`, `timerCtx`, `valueCtx`; читать вместе с блоком «Под капотом».
- `RU` [Каскадная отмена контекстов в Go](https://habr.com/ru/articles/921902/): от `propagateCancel` в исходниках до паттернов в продакшне, включая `Cause`.
- `EN` [Документация пакета context](https://pkg.go.dev/context): официальное описание каждой функции, особенно `AfterFunc`, `WithoutCancel`, `Cause`, и примеры.
- `EN` [Go Concurrency Patterns: Context](https://go.dev/blog/context): классическая статья из блога Go, откуда пошли идиомы.
- `EN` [Contexts and structs](https://go.dev/blog/context-and-structs): почему контекст передают параметром, а не хранят в структуре.

## Домашка

Все задачи в пакете `homework/m04/l05-context/`. Тесты используют `testing/synctest`: внутри «пузыря» время фейковое, поэтому секунды ожидания проходят мгновенно, а горутина, зависшая навсегда, роняет тест с сообщением `deadlock`. Подробнее про synctest в уроке [про гонки, дедлоки и утечки](./l07-problems).

Как читать падение synctest:

- `panic: deadlock: main bubble goroutine has exited but blocked goroutines remain`: тест закончился, а твоя горутина осталась висеть навсегда. Это утечка.
- `panic: deadlock: all goroutines in bubble are blocked`: тест ждёт того, что никогда не случится, например канал никто не закроет или никто в него не отправит. Стек может указывать на хелпер в тесте, а виноват твой код.
- В стеках ищи горутины с пометкой `(durable), synctest bubble N` и строки из своего файла `exN_*.go`: там и застряло.
- Время в сообщениях фейковое: «длился 5s» у теста, который шёл миллисекунды, это нормально.
- Паника останавливает остальные тесты пакета, поэтому запускай задания по одному через `-run`, как в командах ниже.
- Если тест висит без вывода, скорее всего это дедлок на `sync.Mutex`: такую блокировку synctest не считает «надёжной» (её может снять кто-то вне пузыря) и не распознаёт. Добавь `-timeout 30s`, и при зависании Go напечатает стеки всех горутин.

<Hw id="ex1" title="Sleep с отменой" level="база">

Файл `ex1_sleep.go`, функция `Sleep(ctx, d)`: подождать `d` или отмену, что раньше. Без новых горутин; таймер после выхода лучше остановить.

```bash
cd homework && go test -v -race ./m04/l05-context/ -run TestSleep
```

</Hw>

<Hw id="ex2" title="Первый успешный ответ" level="посложнее">

Файл `ex2_first.go`, функция `First`: запустить функции конкурентно, вернуть первый успех, остальным отменить контекст, при полном провале собрать ошибки через `errors.Join`. Возвращаться только после завершения всех горутин.

```bash
cd homework && go test -v -race ./m04/l05-context/ -run TestFirst
```

</Hw>

<Hw id="ex3" title="Request ID в контексте" level="база">

Файл `ex3_values.go`: `WithRequestID` и `RequestID` с приватным типом ключа. Тест проверяет, что строковый ключ `"request_id"` из чужого кода не читает и не перетирает твоё значение.

```bash
cd homework && go test -v ./m04/l05-context/ -run TestRequestID
```

</Hw>

<Hw id="ex4" title="Таймаут с причиной" level="посложнее">

Файл `ex4_call.go`, функция `Call`: выполнить `fn` с `WithTimeoutCause` и вернуть наружу настоящую причину отмены: `ErrSlow` для своего таймаута, причину родителя, если отменили его, `DeadlineExceeded`, если у родителя дедлайн короче.

```bash
cd homework && go test -v -race ./m04/l05-context/ -run TestCall
```

</Hw>

<Hw id="ex5" title="Фоновая работа после запроса" level="база">

Файл `ex5_detach.go`, функция `Detach`: контекст со значениями родителя, без его отмены и со своим таймаутом.

```bash
cd homework && go test -v -race ./m04/l05-context/ -run TestDetach
```

</Hw>

<Hw id="ex6" title="Слияние двух контекстов" level="со звёздочкой">

Файл `ex6_merge.go`, функция `Merge(a, b)`: контекст отменяется, когда отменён любой из двух, причина берётся у того, кто первым. После `cancel` не должно оставаться горутин, даже если `b` живёт вечно. Тест считает горутины: решение «горутина ждёт `<-b.Done()`» не пройдёт. Подсказка: `AfterFunc`.

```bash
cd homework && go test -v -race ./m04/l05-context/ -run TestMerge
```

</Hw>

<Hw id="ex7" title="Отмена в живом HTTP-сервере" level="посложнее">

Без тестов, руками. Напиши сервер с хендлером, который 10 секунд ждёт в `select` между `time.After(10*time.Second)` и `r.Context().Done()` и в обоих случаях пишет в лог, чем всё кончилось, включая `context.Cause(r.Context())`. Запусти `curl localhost:8080/slow` и прерви его через секунду по Ctrl+C. Убедись, что в логе отмена, а не «доработал до конца». Потом добавь внутри хендлера `context.WithTimeout(r.Context(), 3*time.Second)` и проверь, что срабатывает уже он.

</Hw>
