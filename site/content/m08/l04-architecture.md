# Архитектура сервиса

Первая версия сервиса почти всегда выглядит одинаково: обработчик Gin, внутри разбор JSON, пара проверок, SQL-запрос, разбор ошибки базы по тексту и ответ. Пока ручек пять, это даже удобно. На двадцатой начинается боль: бизнес-правило «нельзя выдать книгу, если экземпляров нет» размазано по трём обработчикам, тест на него требует поднять HTTP и базу, а смена `pgx` на что-то другое означает переписать каждую ручку.

Слои handler, service и repository решают ровно эту проблему: каждому коду своё место, зависимости идут в одну сторону, бизнес-логика тестируется без HTTP и без базы. Но у слоёв есть обратная сторона: интерфейс на каждую структуру, пять уровней пустых прокси-функций, папки `domain/entity/valueobject/usecase/port/adapter` для CRUD из трёх таблиц. На собесе ценят, когда ты умеешь объяснить и пользу слоёв, и момент, когда их хватит.

В этом уроке: слои и направление зависимостей, «чистая» и гексагональная архитектура без фанатизма, раскладка проекта (`cmd/`, `internal/`, почему не `pkg/`), интерфейсы на стороне потребителя, DI руками, доменные ошибки и контекст через слои.

## От спагетти к слоям

Вот кусок обработчика из домашки (пакет `legacy`):

```go
r.POST("/books/:id/take", func(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var copies int
	var title string
	err := pool.QueryRow(c, `SELECT title, copies FROM books WHERE id = $1`, id).Scan(&title, &copies)
	if err != nil {
		c.JSON(404, gin.H{"error": "not found"})
		return
	}
	if copies == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no copies"})
		return
	}
	_, err = pool.Exec(c, `UPDATE books SET copies = $1 WHERE id = $2`, copies-1, id)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"id": id, "title": title, "copies": copies - 1})
})
```

Найди проблемы, прежде чем читать дальше. Вот они:

1. **Гонка**: «прочитать copies, вычесть в Go, записать» под параллельными запросами выдаст последний экземпляр дважды.
2. **Любая ошибка базы превращается в 404**, включая «база недоступна».
3. **Текст ошибки базы уходит клиенту** (`err.Error()` в 500): там бывают имена таблиц, куски SQL, адрес сервера.
4. **`c` передан как контекст**: у `*gin.Context` по умолчанию `Done()` возвращает `nil`, так что отмена запроса клиентом не дойдёт до базы.
5. **Мусорный id молча становится нулём** (`id, _ :=`).
6. Правило «нет экземпляров» и HTTP-код 400 перемешаны с SQL. Протестировать правило можно, только подняв HTTP и базу.

Разложим по слоям:

```text
            HTTP-запрос
                │
  ┌─────────────▼─────────────┐
  │ handler (httpapi)         │  разбор JSON и параметров, вызов сервиса,
  │                           │  перевод ошибок в HTTP-коды
  └─────────────┬─────────────┘
                │ вызывает через интерфейс BookService
  ┌─────────────▼─────────────┐
  │ service                   │  бизнес-правила: проверки, нормализация,
  │                           │  сценарии из нескольких шагов, транзакции
  └─────────────┬─────────────┘
                │ вызывает через интерфейс BookRepository
  ┌─────────────▼─────────────┐
  │ repository (postgres)     │  SQL, перевод ошибок драйвера в доменные
  └─────────────┬─────────────┘
                ▼
            PostgreSQL

  library (домен): Book, ErrNotFound, ValidationError. Его импортируют все.
```

Правило одно: **зависимости идут сверху вниз и к домену**. Сервис не знает про Gin и HTTP-коды, репозиторий не знает про JSON, домен не знает ни про что.

| Слой | Знает про | Не знает про | Тестируется |
|---|---|---|---|
| handler | HTTP, Gin, JSON, сервис (интерфейс) | SQL, pgx | httptest + fake-сервис |
| service | домен, репозиторий (интерфейс) | HTTP, SQL | fake-репозиторий в памяти |
| repository | SQL, pgx, домен | HTTP, бизнес-правила | настоящая база (testdb, testcontainers) |

## Чистая архитектура без фанатизма

«Чистая архитектура» Роберта Мартина и «гексагональная» (ports and adapters) Алистера Кокберна говорят одно и то же разными словами: **бизнес-логика в центре и не зависит от инфраструктуры**. База, HTTP, очередь, внешний API это «адаптеры» снаружи, а центр общается с ними через «порты», то есть интерфейсы, которые он сам объявляет.

В Go это выражается почти без церемоний: сервис объявляет нужный ему интерфейс репозитория, а пакет `postgres` его неявно реализует. Никаких `implements`, никаких регистраций.

Где начинается фанатизм:

- **Интерфейс на каждую структуру** «на будущее»: `UserService` + `UserServiceImpl`, `IUserRepository`. Интерфейс заводят, когда есть второй потребитель или подмена в тестах, а не по шаблону.
- **Слой-прокси**: сервис, который в каждом методе просто зовёт репозиторий с теми же аргументами. Если правил нет, это нормально для пары методов, но если так устроен весь сервис, слой лишний: handler может звать репозиторий напрямую (через интерфейс), правила появятся, тогда и появится сервис.
- **Тройной маппинг**: `CreateUserRequest` → `CreateUserCommand` → `User` → `UserModel` → `UserRow` → `UserResponse`. Отдельные DTO нужны на границах, которые меняются независимо: тело HTTP-ответа (это контракт с клиентом) и строка базы. Между ними хватит доменной структуры.
- **Папки по учебнику DDD** для сервиса из трёх таблиц.

Признак здоровой архитектуры простой: новую ручку добавить легко, правило протестировать дёшево, а при чтении кода не нужно прыгать через пять файлов, чтобы найти SQL.

## Раскладка проекта

Официальная рекомендация живёт на go.dev в статье «Organizing a Go module». Для сервиса она сводится к этому:

```text
library/
├── go.mod
├── cmd/
│   └── library/
│       └── main.go          # сборка: конфиг, пул, слои, http.Server
├── internal/
│   ├── library/             # домен: Book, ошибки
│   ├── service/             # бизнес-логика
│   ├── postgres/            # репозитории
│   └── httpapi/             # обработчики Gin
└── migrations/
```

- **`cmd/<имя>/main.go`**: точки входа. Бинарников может быть несколько (API, воркер очереди, утилита миграций), у каждого своя папка. `main` тонкий: прочитать конфиг, собрать зависимости, запустить.
- **`internal/`**: код, который нельзя импортировать **снаружи модуля**. Это правило компилятора, а не соглашение: `go build` откажется собирать чужой импорт с ошибкой `use of internal package ... not allowed`. Для сервиса весь код в `internal/`: у него нет внешних пользователей, и так ты свободно меняешь API пакетов.
- **`pkg/`** придуман сообществом для «кода, который можно импортировать снаружи». В сервисе таких пакетов почти никогда нет, а если есть, им честнее жить в корне модуля или в отдельном модуле. Стандартная библиотека и сам Go `pkg/` не используют.

::: trap golang-standards/project-layout это не стандарт
Репозиторий `golang-standards/project-layout` с десятками папок (`/api`, `/web`, `/build`, `/deployments`, `/pkg`, `/vendor`...) часто считают официальным стандартом из-за названия организации. Это не так: Расс Кокс из команды Go открыл в нём issue с заголовком «this is not a standard Go project layout», и теперь в README крупно написано, что это «NOT an official standard defined by the core Go dev team». Плохо: копировать всё дерево в сервис на три ручки. Хорошо: `cmd/` и `internal/`, остальные папки по мере реальной необходимости.
:::

### По слоям или по фичам

Раскладка выше идёт **по слоям**. Когда предметных областей становится много (книги, читатели, штрафы), её режут **по фичам**: `internal/books/{service.go,postgres.go,http.go}`, `internal/readers/...`. Всё про книги рядом, а пакет `books` можно читать сверху вниз. Оба варианта нормальны, для сервиса из пары сущностей слои проще, для большого монолита удобнее фичи.

## Интерфейсы на стороне потребителя

В Java интерфейс объявляют рядом с реализацией. В Go наоборот: **интерфейс объявляет тот, кто им пользуется**, и ровно с теми методами, которые ему нужны.

```go
// package service
type BookRepository interface {
	Create(ctx context.Context, b library.Book) (library.Book, error)
	Get(ctx context.Context, id int64) (library.Book, error)
	TakeCopy(ctx context.Context, id int64) (library.Book, error)
}

type Service struct{ repo BookRepository }

func New(repo BookRepository) *Service { return &Service{repo: repo} }
```

```go
// package postgres: про service ничего не знает
type BookRepo struct{ pool *pgxpool.Pool }

func NewBookRepo(pool *pgxpool.Pool) *BookRepo { return &BookRepo{pool: pool} }
func (r *BookRepo) Create(ctx context.Context, b library.Book) (library.Book, error) { ... }
```

::: hood Почему это работает и не даёт циклов импорта
Интерфейсы в Go реализуются неявно: тип удовлетворяет интерфейсу, если у него есть нужные методы, и это проверяется при присваивании (`service.New(repo)`), а не при объявлении типа. Поэтому:

- `postgres` не импортирует `service`, `service` не импортирует `postgres`, оба импортируют только домен. Связывает их `main`, которому позволено знать всех. Цикл импорта, который в Go запрещён компилятором, просто не может возникнуть.
- Интерфейс получается маленьким: у потребителя три метода, даже если у `BookRepo` их двадцать. Fake в тестах реализует три метода, а не двадцать.
- Правило из Go Code Review Comments: интерфейсы обычно принадлежат пакету, который их использует; реализующий пакет возвращает конкретный тип. Отсюда поговорка «принимай интерфейсы, возвращай структуры».

Проверить на этапе компиляции, что `BookRepo` подходит, можно одной строкой в `main` или тесте: `var _ service.BookRepository = (*postgres.BookRepo)(nil)`.
:::

## DI руками в main

Dependency injection в Go это не фреймворк, а привычка передавать зависимости аргументами конструктора. Вся «магия» умещается в `main`:

```go
func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := pgxpool.New(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		return err
	}
	defer pool.Close()

	repo := postgres.NewBookRepo(pool)
	svc := service.New(repo)
	h := httpapi.New(svc)

	r := gin.New()
	h.Register(r.Group("/api"))

	srv := &http.Server{Addr: ":8080", Handler: r, ReadHeaderTimeout: 5 * time.Second}
	// запуск и graceful shutdown, как в модуле про net/http
}
```

Плюсы: граф зависимостей виден глазами, ошибка в нём это ошибка компиляции, IDE переходит к определению, нет глобальных переменных (`var db *sql.DB` на уровне пакета ломает тесты и параллельный запуск).

Когда сервисов десятки, `main` разрастается, и появляются инструменты:

- **google/wire**: генерация кода. Описываешь провайдеры, `wire` генерирует ту самую функцию сборки, как если бы ты написал её руками. Ошибки ловятся при генерации. Сейчас репозиторий переведён в архив (read-only), новых версий не будет.
- **uber-go/fx**: DI во время выполнения на рефлексии плюс жизненный цикл (`OnStart`/`OnStop`). Мощно для больших монолитов, но ошибки графа видны только при запуске, а стек вызовов обрастает слоями fx.

Для джуна и для большинства сервисов ответ «собираю руками в `main`, это 30 строк, и их легко читать» абсолютно нормальный.

## Доменные ошибки и их маппинг

Каждый слой говорит на своём языке ошибок: pgx на языке `pgx.ErrNoRows` и `*pgconn.PgError` с кодом `23505`, HTTP на языке кодов 404 и 409. Мост между ними это **доменные ошибки**:

```go
// package library
var (
	ErrNotFound  = errors.New("книга не найдена")
	ErrISBNTaken = errors.New("книга с таким ISBN уже есть")
	ErrNoCopies  = errors.New("свободных экземпляров нет")
)

type ValidationError struct{ Field, Message string }

func (e *ValidationError) Error() string { return e.Field + ": " + e.Message }
```

Репозиторий переводит ошибки драйвера в доменные:

```go
var pgErr *pgconn.PgError
if errors.As(err, &pgErr) && pgErr.Code == "23505" { // unique_violation
	return library.Book{}, library.ErrISBNTaken
}
if errors.Is(err, pgx.ErrNoRows) {
	return library.Book{}, library.ErrNotFound
}
```

Сервис возвращает их как есть или обёрнутыми (`fmt.Errorf("take book %d: %w", id, err)`), а HTTP-слой переводит в коды **в одном месте**:

```go
func writeError(c *gin.Context, err error) {
	var ve *library.ValidationError
	switch {
	case errors.As(err, &ve):
		respond(c, 422, errorBody{Code: "validation_failed", Message: "проверь поля запроса",
			Fields: map[string]string{ve.Field: ve.Message}})
	case errors.Is(err, library.ErrNotFound):
		respond(c, 404, errorBody{Code: "not_found", Message: library.ErrNotFound.Error()})
	case errors.Is(err, library.ErrISBNTaken):
		respond(c, 409, errorBody{Code: "isbn_taken", Message: library.ErrISBNTaken.Error()})
	case errors.Is(err, library.ErrNoCopies):
		respond(c, 409, errorBody{Code: "no_copies", Message: library.ErrNoCopies.Error()})
	default:
		log.ErrorContext(c.Request.Context(), "unexpected error", "err", err) // подробности в лог
		respond(c, 500, errorBody{Code: "internal", Message: "внутренняя ошибка"})
	}
}
```

Формат тела тот же, что в уроке про binding: `{"error": {"code", "message", "fields"}}`, где `fields` есть только у ошибок валидации. В `message` идёт текст доменной ошибки, а не `err.Error()`: обёрнутая ошибка несёт контекст вроде `take book 7: ...`, который клиенту ни к чему.

Почему одно место: когда маппинг размазан по обработчикам, одна ручка отвечает на «не найдено» 404, другая 400, третья 500, а новая доменная ошибка в половине мест проваливается в 500.

```go
var ErrNotFound = errors.New("не найдено")

wrapped := fmt.Errorf("get book 7: %w", ErrNotFound)
fmt.Println(wrapped == ErrNotFound, errors.Is(wrapped, ErrNotFound))
```

<Predict :options="['true true', 'false true', 'false false', 'true false']" :correct="1">

`%w` создаёт новую ошибку, которая **оборачивает** исходную: это другое значение, и `==` даёт `false`. `errors.Is` разворачивает цепочку через `Unwrap` и находит `ErrNotFound`. Поэтому в маппинге всегда `errors.Is`/`errors.As`: стоит сервису добавить контекст к ошибке, и сравнение через `==` молча превратит 404 в 500.

</Predict>

::: trap Не сравнивай текст ошибки базы
Плохо: `strings.Contains(err.Error(), "duplicate key")`. Текст ошибки PostgreSQL зависит от `lc_messages` сервера (на русской локали он будет по-русски) и от версии. Хорошо: `errors.As(err, &pgErr)` и код SQLSTATE `23505`: коды стабильны и описаны в приложении A документации PostgreSQL.
:::

Ещё одна ловушка с ошибками-типами, типизированный nil, разобрана в задаче в конце урока: обязательно реши её, это любимый вопрос собеса.

## Контекст через слои

`context.Context` проходит через все слои первым аргументом: от обработчика до `pool.Query(ctx, ...)`. Он несёт отмену (клиент закрыл соединение, сработал таймаут из middleware), дедлайн и значения запроса (request ID для логов).

В обработчике Gin контекст запроса это `c.Request.Context()`, а не сам `c`:

```go
type key struct{}

r.Use(func(c *gin.Context) {
	c.Request = c.Request.WithContext(context.WithValue(c.Request.Context(), key{}, "req-42"))
})
r.GET("/", func(c *gin.Context) {
	var ctx context.Context = c // *gin.Context реализует context.Context
	fmt.Println(ctx.Value(key{}), c.Request.Context().Value(key{}))
})
```

<Predict :options="['req-42 и req-42', 'nil и req-42', 'req-42 и nil', 'не скомпилируется: gin.Context не context.Context']" :correct="1">

Вывод: `<nil> req-42`. `*gin.Context` формально реализует `context.Context`, но по умолчанию (`engine.ContextWithFallback == false`) его `Done()` возвращает `nil`, `Err()` всегда `nil`, а `Value` видит только ключи из `c.Set` (строковые) и не заглядывает в контекст запроса. Передав `c` в сервис, ты теряешь и отмену, и request ID. Можно включить `r.ContextWithFallback = true`, но явное `c.Request.Context()` надёжнее: сервис вообще не должен знать про Gin.

</Predict>

::: trap gin.Context в сигнатуре сервиса
Плохо: `func (s *Service) Take(c *gin.Context, id int64)`. Сервис привязан к Gin, его нельзя вызвать из воркера очереди или CLI, а тест требует собрать `gin.Context`. Хорошо: `func (s *Service) Take(ctx context.Context, id int64)`, а в обработчике `h.svc.Take(c.Request.Context(), id)`.
:::

## Тесты по слоям

Слои окупаются в тестах:

- **Сервис** тестируется с fake-репозиторием: структура в памяти, которая реализует три метода интерфейса и запоминает вызовы. Тесты быстрые и проверяют правила: нормализация ISBN, отказ без похода в базу, проброс доменных ошибок.
- **Репозиторий** тестируется на **настоящем PostgreSQL**, потому что его работа это SQL: уникальные индексы, `RETURNING`, атомарность `UPDATE ... WHERE copies > 0`. Мок базы здесь проверял бы, что ты вызвал мок.
- **Обработчики** тестируются через `httptest` с fake-сервисом: коды, формат ошибок, разбор параметров.
- **Сквозной тест** собирает всё как в `main` и гоняет пару сценариев на настоящей базе.

Fake пишется руками за пять минут, генераторы моков (mockery, gomock) нужны реже, чем кажется. Подробно про тестирование сервиса будет в уроке о тестах этого модуля.

## Что выведет программа?

```go
type Repo interface{ Get(id int64) (string, error) }

type pg struct{}

func (*pg) Get(id int64) (string, error) { return "книга", nil }
func (*pg) Delete(id int64) error        { return nil }

var r Repo = &pg{}
_, ok := r.(interface{ Delete(int64) error })
fmt.Println(ok)
```

<Predict :options="['true', 'false', 'не скомпилируется: в Repo нет Delete', 'паника']" :correct="0">

Интерфейс `Repo` ограничивает только то, что видит потребитель, но динамический тип внутри остаётся `*pg` со всеми методами. Утверждение типа на другой интерфейс проверяет методы конкретного типа и находит `Delete`. Так устроены необязательные возможности в stdlib (например, `http.Flusher` у `ResponseWriter`). А вот вызвать `r.Delete` напрямую не даст компилятор: у статического типа `Repo` такого метода нет.

</Predict>

```go
type NotFoundError struct{ ID int64 }

func (e *NotFoundError) Error() string { return fmt.Sprintf("книга %d не найдена", e.ID) }

func find(id int64) error {
	var err *NotFoundError
	if id > 10 {
		err = &NotFoundError{ID: id}
	}
	return err
}

fmt.Println(find(1) == nil)
```

<Predict :options="['true', 'false', 'не скомпилируется', 'паника: nil pointer dereference']" :correct="1">

Интерфейс `error` равен `nil`, только если у него нет ни типа, ни значения. Здесь в интерфейс положили типизированный `(*NotFoundError)(nil)`: тип есть, значение `nil`, и интерфейс не равен `nil`. Вызывающий решит, что произошла ошибка, и ответит 500. Правило: функция, возвращающая `error`, возвращает литерал `nil` для успеха, а не переменную конкретного типа ошибки. Это классический вопрос собеса про интерфейсы.

</Predict>

::: interview
<details><summary>Зачем делить сервис на handler, service и repository?</summary>

Чтобы у каждого кода было одно место и одна причина меняться: HTTP-детали в обработчике, бизнес-правила в сервисе, SQL в репозитории. Правила тестируются быстро без HTTP и базы, транспорт можно добавить (gRPC, очередь) без переписывания логики, хранилище можно поменять за интерфейсом. Зависимости идут в одну сторону, к домену.

</details>
<details><summary>Что такое чистая или гексагональная архитектура и как она выглядит в Go?</summary>

Бизнес-логика в центре и не зависит от инфраструктуры; база, HTTP, очереди это адаптеры, которые общаются с центром через порты (интерфейсы, объявленные центром). В Go это сервис с интерфейсом репозитория в своём пакете и пакет `postgres`, неявно его реализующий. Без папок по учебнику и интерфейсов на всё подряд.

</details>
<details><summary>Где в Go объявлять интерфейс и почему?</summary>

В пакете-потребителе и с минимальным набором методов. Реализация возвращает конкретный тип. Так нет циклов импорта, интерфейсы маленькие, fake в тестах простой. Go Code Review Comments формулирует это так: интерфейс обычно принадлежит пакету, который его использует, а реализующий пакет возвращает конкретный тип. Отсюда поговорка «принимай интерфейсы, возвращай структуры».

</details>
<details><summary>Какую структуру папок ты бы сделал для сервиса? Нужен ли pkg/?</summary>

`cmd/<имя>/main.go` для точек входа, весь остальной код в `internal/`: компилятор запрещает импортировать его снаружи модуля, и API пакетов можно свободно менять. `pkg/` не нужен: в сервисе нет кода для внешних потребителей. `golang-standards/project-layout` не официальный стандарт, это прямо написано в его README; официальная рекомендация в статье «Organizing a Go module» на go.dev.

</details>
<details><summary>Как ты делаешь dependency injection в Go?</summary>

Руками: конструкторы принимают зависимости аргументами, граф собирается в `main` (пул, репозиторий, сервис, обработчики, сервер). Никаких глобальных переменных. Для больших графов есть `wire` (генерация кода, ошибки при генерации, но репозиторий в архиве) и `fx` (рефлексия в рантайме, жизненный цикл, ошибки видны при старте).

</details>
<details><summary>Как ошибки базы превращаются в HTTP-коды?</summary>

Репозиторий переводит ошибки драйвера в доменные (`pgx.ErrNoRows` в `ErrNotFound`, SQLSTATE 23505 в `ErrISBNTaken`), сервис пробрасывает их, при необходимости оборачивая через `%w`, а HTTP-слой в одной функции через `errors.Is`/`errors.As` выбирает код и тело. Неизвестные ошибки дают 500 с общим текстом, подробности только в лог.

</details>
<details><summary>Почему в сервис передают c.Request.Context(), а не *gin.Context?</summary>

Сервис не должен зависеть от Gin. Кроме того, у `*gin.Context` по умолчанию (`ContextWithFallback = false`) `Done()` возвращает `nil`, а `Value` не видит значений контекста запроса: отмена и request ID потеряются. `c.Request.Context()` несёт и отмену, и дедлайн из middleware таймаута, и значения.

</details>
<details><summary>Почему функция вернула ошибку, хотя внутри был nil?</summary>

Потому что вернули типизированный nil: переменную типа `*MyError` со значением `nil`, упакованную в интерфейс `error`. У интерфейса есть тип, поэтому он не равен `nil`. Успех нужно возвращать литералом `nil`.

</details>
:::

## Источники

- `RU` [От wtf-кода к чистой архитектуре](https://habr.com/ru/articles/691562/): перевод главы книги «Go With The Domain» от Three Dots Labs, рефакторинг реального Go-сервиса в слои.
- `RU` [Архитектура продуктового Go-сервиса](https://habr.com/ru/articles/881918/): как устроены service и usecase в живом проекте и куда их ведёт рефакторинг.
- `RU` [Паттерны внедрения зависимостей в Gin](https://gin-gonic.com/ru/docs/middleware/dependency-injection/): официальная документация Gin, DI через замыкания и структуры обработчиков.
- `EN` [Organizing a Go module](https://go.dev/doc/modules/layout): официальная рекомендация по раскладке, раздел Server project.
- `EN` [this is not a standard Go project layout](https://github.com/golang-standards/project-layout/issues/117): issue Расса Кокса в golang-standards/project-layout.
- `EN` [Go Code Review Comments](https://go.dev/wiki/CodeReviewComments): разделы Interfaces и Contexts.
- `EN` [DDD Lite in Go](https://threedots.tech/post/ddd-lite-in-go-introduction/): Three Dots Labs о том, сколько DDD реально нужно сервису на Go.
- `EN` [Standard Package Layout](https://www.gobeyond.dev/standard-package-layout/): Бен Джонсон о доменном пакете в центре и адаптерах вокруг.

## Домашка

Домашка урока это рефакторинг. В `homework/m08/l04-architecture/legacy/legacy.go` лежит спагетти-обработчик библиотеки. Раскладываешь его по слоям под готовые интерфейсы: домен в `library/` уже есть, схема и данные в `testdata/schema.sql`, точка сборки в `cmd/library/main.go`. Тесты репозитория и сквозной тест ходят в учебный PostgreSQL: подними его, если ещё не поднят (`cd homework && docker compose up -d`). Посмотреть данные руками: `docker compose exec postgres psql -U goschool`.

<Hw id="ex1" title="Сервисный слой" level="база">

Файл `service/ex1_service.go`: `Create` (нормализация и проверки по порядку, `*library.ValidationError`), `Get`, `Take`. Тесты идут с fake-репозиторием в памяти: проверяют, что при ошибке валидации репозиторий не вызывается, ISBN нормализуется, длина считается в символах, а арабско-индийские цифры не проходят за «13 цифр».

```bash
cd homework && go test -v ./m08/l04-architecture/service/
```

</Hw>

<Hw id="ex2" title="Репозиторий на PostgreSQL" level="посложнее">

Файл `postgres/ex2_repository.go`: `Create` с переводом SQLSTATE 23505, `Get`, атомарный `TakeCopy`. Тест гонки детерминированный: держит строку книги под `FOR UPDATE`, ждёт через `pg_stat_activity`, пока оба `TakeCopy` упрутся в блокировку, и отпускает. «Прочитать, вычесть, записать» выдаст последний экземпляр дважды и тест упадёт.

```bash
cd homework && go test -v -race ./m08/l04-architecture/postgres/
```

</Hw>

<Hw id="ex3" title="HTTP-слой и маппинг ошибок" level="база">

Файл `httpapi/ex3_handler.go`: `Register` и `writeError`. Тесты с fake-сервисом проверяют коды и формат для каждой доменной ошибки (в том числе обёрнутых), что текст неизвестной ошибки не уходит клиенту и что в сервис попадает `c.Request.Context()`, а не `*gin.Context`.

```bash
cd homework && go test -v ./m08/l04-architecture/httpapi/
```

</Hw>

<Hw id="ex4" title="Сборка руками" level="база">

Файл `ex4_app.go`: `NewRouter` связывает репозиторий, сервис и обработчики на группе `/api`. Сквозной тест проходит сценарий «создать, найти, взять до нуля» на настоящей базе. После зелёного запусти сервис командой из комментария в `cmd/library/main.go` (схему сначала загрузи в базу) и пройди тот же сценарий `curl`-ом.

```bash
cd homework && go test -v ./m08/l04-architecture/
```

</Hw>

<Hw id="ex5" title="Разбор legacy" level="посложнее">

Без кода: выпиши все проблемы `legacy.go`, которые найдёшь (в уроке названы шесть, их там больше: посмотри на `BindJSON`, на проверку `title`, на статусы). Для каждой укажи, какой слой твоего решения её закрывает и каким тестом это проверяется.

</Hw>
