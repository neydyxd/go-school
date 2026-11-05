# Тестирование сервиса

Тесты отдельных функций ты писал с [модуля про тестирование](/m03/l01-testing): табличные тесты, фейки, бенчмарки. Сервис добавляет новые вопросы. Как проверить бизнес-логику, не поднимая базу? Как протестировать хендлер Gin, не запуская сервер? Что делать с SQL: мокать или гонять в настоящем PostgreSQL? Почему тест, который проходил месяц, вдруг упал в CI и сам починился на перезапуске?

На собесе по бэкенду вопрос «как ты тестируешь свой сервис» звучит почти всегда, и ответ «юнит-тесты на моках» уже не впечатляет. Хороший ответ звучит так: сервисный слой с fake, хендлеры через `httptest` с настоящим роутером, репозиторий в настоящей базе с изоляцией тестов, немного сквозных проверок, и понимание, откуда берутся flaky-тесты.

В конце урока ТЗ проекта модуля: REST API «Трекер задач», где всё из уроков про Gin, безопасность, конфиг и тесты собирается вместе.

## Пирамида тестов для бэкенда

```text
            /\        e2e: поднятый сервис + база, curl или HTTP-клиент
           /  \       единицы штук, медленные, ловят ошибки сборки и конфига
          /----\
         /      \     интеграционные: репозиторий в настоящей PostgreSQL,
        /        \    роутер + хендлеры + middleware через httptest
       /----------\   десятки, секунды
      /            \
     /              \ юнит: сервис с fake, валидация, переходы статусов, чистые функции
    /________________\ сотни, миллисекунды
```

Идея пирамиды (Майк Кон, развитие у Мартина Фаулера): чем выше уровень, тем тест дороже, медленнее и хрупче, поэтому внизу их больше. Для типичного CRUD-сервиса на Go пирамиду честно корректируют: бизнес-логики часто немного, а основные ошибки живут в SQL и в HTTP-обвязке (не тот код ответа, забытая проверка владельца). Поэтому интеграционных тестов в Go-сервисах обычно больше, чем в учебной картинке. Они достаточно быстрые: тест репозитория в локальной PostgreSQL идёт миллисекунды.

| Слой | Чем тестируем | Что ловим |
|---|---|---|
| service | юнит-тесты с fake-репозиторием | бизнес-правила, валидацию, проброс ошибок |
| http | `httptest` + настоящий роутер + fake-сервисы | коды ответов, формат JSON, middleware, binding |
| repository | настоящая PostgreSQL, своя схема на тест | SQL, ограничения, NULL, пагинацию, IDOR в запросах |
| весь сервис | запущенный бинарник, curl | конфиг, миграции, wiring в `main` |

## Юнит-тесты сервиса с fake

### Fake, stub, mock, spy

Термины путают даже на собесах, поэтому коротко:

- **Stub** возвращает заранее заданный ответ: `Get` всегда отдаёт задачу 7.
- **Fake** это рабочая упрощённая реализация: репозиторий на map вместо PostgreSQL. Ведёт себя как настоящий.
- **Spy** это fake или stub, который ещё и запоминает вызовы: сколько раз вызвали `UpdateStatus`.
- **Mock** заранее программируется ожиданиями («должен быть вызван `Get(1, 7)` ровно один раз, затем `UpdateStatus`») и сам проверяет их. В Go это `gomock` или `mockery` + `testify/mock`.

В Go-сообществе предпочитают **fake**. Интерфейсы маленькие и объявлены у потребителя (это ты видел в [уроке про архитектуру](/m08/l04-architecture)), поэтому fake пишется за пять минут. Тест на fake проверяет **поведение** («после смены статуса задача в статусе done»), а тест на моках часто проверяет **реализацию** («вызвал Get, затем Update»): любой рефакторинг ломает такой тест, хотя поведение не изменилось.

```go
type fakeRepo struct {
	mu    sync.Mutex
	tasks map[int64]Task
}

func (f *fakeRepo) Get(_ context.Context, ownerID, id int64) (Task, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	t, ok := f.tasks[id]
	if !ok || t.OwnerID != ownerID {
		return Task{}, ErrNotFound
	}
	return t, nil
}
// ... остальные методы интерфейса TaskRepo

func TestChangeStatus(t *testing.T) {
	tests := []struct {
		name     string
		from, to string
		owner    int64
		wantErr  error
	}{
		{"todo -> in_progress", "todo", "in_progress", 1, nil},
		{"todo -> done нельзя", "todo", "done", 1, ErrBadTransition},
		{"чужая задача", "todo", "in_progress", 2, ErrNotFound},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeRepo{tasks: map[int64]Task{10: {ID: 10, OwnerID: 1, Status: tt.from}}}
			_, err := NewService(repo).ChangeStatus(context.Background(), tt.owner, 10, tt.to)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("ChangeStatus err = %v, want %v", err, tt.wantErr)
			}
		})
	}
}
```

Каждый подтест получает свой fake: тесты не влияют друг на друга, их можно запускать в любом порядке и параллельно.

### Контрактные тесты: чтобы fake не врал

Опасность fake в том, что он может вести себя иначе, чем база. Например, fake возвращает задачи в порядке вставки, а SQL без `ORDER BY` в произвольном. Тесты сервиса зелёные, прод сломан. Лекарство: **один набор тестов для обеих реализаций интерфейса**.

```go
// contract_test.go
func testRepoContract(t *testing.T, newRepo func(t *testing.T) TaskRepo) {
	t.Run("Get видит только свои задачи", func(t *testing.T) { ... })
	t.Run("List сортирует по id и режет страницу", func(t *testing.T) { ... })
}

func TestFakeRepo(t *testing.T) { testRepoContract(t, func(*testing.T) TaskRepo { return NewFakeRepo() }) }
func TestPgRepo(t *testing.T)   { testRepoContract(t, func(t *testing.T) TaskRepo { return NewPgRepo(testdb.New(t, "testdata/schema.sql")) }) }
```

Если fake и PostgreSQL разошлись, упадёт один из прогонов. Так устроена домашка этого урока.

::: trap Мокать всё подряд
```go
// плохо: тест повторяет реализацию построчно
mockRepo.EXPECT().Get(gomock.Any(), int64(1), int64(10)).Return(task, nil)
mockRepo.EXPECT().UpdateStatus(gomock.Any(), int64(1), int64(10), "done").Return(task, nil)
mockLogger.EXPECT().Info(gomock.Any()).Times(1)
```

Поменял `Get` + `UpdateStatus` на один `UPDATE ... RETURNING`, и тест упал, хотя поведение то же. А ошибку «поставили статус не той задаче» такой тест мог и пропустить, если мок настроен с `gomock.Any()`.

Хорошо: fake хранилища, проверка результата и состояния (`repo.Get` после операции), логгер настоящий с `slog.DiscardHandler` (Go 1.24). Мок оправдан для внешнего мира, который нельзя запустить локально: платёжный шлюз, почтовый сервис. И даже там часто проще поднять `httptest.NewServer` с фейковым ответом.
:::

## Хендлеры через httptest

Пакет `net/http/httptest` даёт две вещи: `NewRequest` собирает `*http.Request` без сети, `NewRecorder` это `http.ResponseWriter`, который запоминает код, заголовки и тело. Gin-движок реализует `http.Handler`, поэтому тест прогоняет запрос через **весь роутер**: middleware, binding, хендлер, формат ошибок.

```go
func init() { gin.SetMode(gin.TestMode) } // без отладочного вывода маршрутов

func TestCreateTask(t *testing.T) {
	tests := []struct {
		name     string
		body     string
		svcErr   error
		wantCode int
		wantErr  string // код ошибки в едином формате
	}{
		{"успех", `{"title":"купить молоко"}`, nil, 201, ""},
		{"битый JSON", `{"title":`, nil, 400, "bad_request"},
		{"пустой title", `{"title":""}`, nil, 422, "validation_failed"},
		{"сервис: проект чужой", `{"title":"x"}`, service.ErrNotFound, 404, "not_found"},
		{"сервис: база упала", `{"title":"x"}`, errors.New("conn reset"), 500, "internal"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewRouter(Deps{Tasks: fakeTasks{err: tt.svcErr}, Tokens: fakeTokens{}, ...})
			req := httptest.NewRequest(http.MethodPost, "/v1/projects/3/tasks", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer user-1")
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			if w.Code != tt.wantCode {
				t.Fatalf("код %d, want %d; тело %s", w.Code, tt.wantCode, w.Body)
			}
			// дальше проверка тела, см. следующий раздел
		})
	}
}
```

Роутер собирается функцией `NewRouter(deps)`, и в тестах в неё передаются fake-сервисы. Если роутер собирается внутри `main`, протестировать его нельзя: это ещё один аргумент за DI руками.

Для теста одного хендлера без роутера есть `gin.CreateTestContext(w)`, он возвращает `*gin.Context` и движок. Но тестировать через роутер полезнее: ловятся ошибки в путях (`/tasks/:id` против `/task/:id`) и в порядке middleware.

```go
h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	// забыли и WriteHeader, и Write
})
w := httptest.NewRecorder()
h.ServeHTTP(w, httptest.NewRequest("GET", "/", nil))
fmt.Println(w.Code, w.Body.Len())
```

<Predict :options="['0 0', '200 0', '500 0', '204 0']" :correct="1">

`NewRecorder` создаёт рекордер с `Code: 200`, это поведение настоящего сервера: хендлер, который ничего не написал, отдаёт 200 с пустым телом. Поэтому тест «код 200» не доказывает, что хендлер что-то сделал. Проверяй тело и состояние, а для пустых ответов ожидай явный 204, который хендлер должен поставить сам.

</Predict>

## Проверка JSON без хрупкости

Сравнивать тело ответа со строкой соблазнительно и плохо:

```go
want := `{"id":1,"title":"a","created_at":"2026-11-05T10:00:00Z"}`
if w.Body.String() != want { ... } // упадёт из-за created_at, порядка полей, пробелов
```

Порядок полей зависит от того, map это или структура, `created_at` меняется при каждом запуске, а `id` зависит от того, сколько строк вставили до этого.

```go
got, _ := json.Marshal(map[string]any{"title": "a", "id": 1})
want := `{"id":1,"title":"a"}`
fmt.Println(string(got) == want)

type Task struct {
	Title string `json:"title"`
	ID    int    `json:"id"`
}
got2, _ := json.Marshal(Task{ID: 1, Title: "a"})
fmt.Println(string(got2) == want)
```

<Predict :options="['true true', 'true false', 'false true', 'false false']" :correct="1">

Ключи map `encoding/json` сортирует по алфавиту, поэтому первая строка совпала. Поля структуры пишутся в порядке объявления: `{"title":"a","id":1}`, и строковое сравнение ломается, хотя JSON по смыслу тот же. Переставил поля в структуре, и упали все тесты со строками. Сравнивай JSON по смыслу.

</Predict>

Варианты от простого к гибкому:

1. **Распаковать в структуру и сравнить поля**, которые важны. Лишние поля игнорируются автоматически. Подходит для большинства тестов.
2. **Распаковать оба JSON в `any` и сравнить по смыслу**, предварительно выбросив нестабильные ключи (`id`, `created_at`). Это делает функция `JSONDiff` из домашки: она ещё и показывает путь к первому отличию, например `$.items[1].title`.
3. **Golden-файлы**: ожидаемый ответ лежит в `testdata/create_task.golden`, тест сравнивает с ним, а флаг `-update` перезаписывает файл. Удобно для больших ответов: в ревью виден дифф golden-файла, то есть изменение контракта API.

```go
var update = flag.Bool("update", false, "перезаписать golden-файлы")

func golden(t *testing.T, name string, got []byte) {
	t.Helper()
	path := filepath.Join("testdata", name+".golden")
	got = normalize(t, got) // выбросить id и created_at, отформатировать с отступами
	if *update {
		if err := os.WriteFile(path, got, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("нет golden-файла, запусти с -update: %v", err)
	}
	if !bytes.Equal(got, want) {
		t.Errorf("ответ не совпал с %s:\ngot:\n%s\nwant:\n%s", path, got, want)
	}
}
```

```bash
go test ./internal/http/ -run TestListTasks -update   # обновить эталоны
git diff testdata/                                    # посмотреть, что изменилось в контракте
```

Golden-файл страхует от тихих несовместимых изменений: переименовал поле, и дифф golden-файла в пулл-реквесте покажет это ревьюеру. Главное правило: `-update` запускают осознанно и читают дифф, а не «обновил, чтобы позеленело».

## Тестирование middleware аутентификации

Middleware аутентификации это код, где ошибка стоит дороже всего, а тестируют его часто одним кейсом «с токеном работает». Минимальная таблица:

| Запрос | Ожидание |
|---|---|
| без `Authorization` | 401, обработчик не вызван, `WWW-Authenticate: Bearer` |
| `Basic ...` вместо `Bearer` | 401 |
| `Bearer` без токена | 401 |
| поддельный, просроченный, с чужим `aud` токен | 401 |
| валидный токен | 200, обработчик видит правильного пользователя |
| пользователь без роли на админском маршруте | 403, не 401 |
| аноним на админском маршруте | 401, не 403 |

Две техники делают такие тесты простыми. Первая: middleware получает функцию проверки токена (`Verifier`), и в тесте подставляется fake, который знает пару строковых токенов. JWT отдельно тестируется в своём пакете. Вторая: время подменяется (`jwt.WithTimeFunc`, поле `now func() time.Time`), поэтому кейс «просроченный токен» не требует ждать 15 минут.

Хорошая проверка набора тестов: **мутанты**. Сломай middleware намеренно (пропускает запрос без заголовка, отвечает 403 вместо 401, не проверяет роль) и убедись, что тесты падают. Если какой-то мутант выжил, в таблице не хватает кейса. Задание ex5 устроено именно так: ты пишешь набор тестов, а проверяющий тест гоняет его на нескольких сломанных версиях.

## Интеграционные тесты репозитория

Мокать базу в тестах репозитория бессмысленно: логика репозитория **и есть** SQL. Опечатка в имени колонки, забытый `ORDER BY`, `NULL` вместо пустой строки, нарушение `CHECK`, ошибка в `JOIN` с проверкой владельца: всё это ловит только настоящая PostgreSQL. Вопрос лишь в том, как изолировать тесты друг от друга.

| Способ | Как | Плюсы | Минусы |
|---|---|---|---|
| Схема на тест | `CREATE SCHEMA t_xxx`, `search_path` в строке подключения, миграции или сид | полная изоляция, параллельно, код может открывать свои транзакции | миграции на каждый тест (для маленькой схемы это миллисекунды) |
| Транзакция на тест | `BEGIN` перед тестом, `ROLLBACK` после, репозиторий работает внутри tx | очень быстро | код под тестом не может сам делать `COMMIT`; в pgx вложенный `Begin` на `Tx` даёт savepoint, но это уже другое поведение |
| Общая база + `TRUNCATE` | очистка таблиц между тестами | просто | тесты нельзя запускать параллельно |
| База из шаблона | `CREATE DATABASE t_xxx TEMPLATE app_tpl` | изоляция на уровне базы, миграции один раз | создание базы медленнее схемы |

::: hood Как устроен internal/testdb
Хелпер, которым ты пользовался в [модуле про PostgreSQL](/m07/l01-postgres-intro), делает изоляцию «схема на тест»:

```text
testdb.New(t, "testdata/schema.sql")
  1. админское соединение: CREATE SCHEMA t_3f9a1c2b7d4e   (6 случайных байт в hex)
  2. t.Cleanup: DROP SCHEMA t_3f9a1c2b7d4e CASCADE
  3. URL базы + ?search_path=t_3f9a1c2b7d4e
     pgx передаёт search_path как параметр запуска сессии, поэтому каждое
     соединение пула видит только эту схему: CREATE TABLE tasks попадёт в неё
  4. pgxpool на этот URL (4 соединения), t.Cleanup(pool.Close)
  5. сид-файлы через простой протокол (можно несколько операторов в одном Exec)
```

Порядок `t.Cleanup` важен: функции выполняются в обратном порядке регистрации (LIFO). Пул регистрируется вторым, поэтому закрывается первым, и только потом удаляется схема. Наоборот было бы `DROP SCHEMA` при живых соединениях пула. Параллельные пакеты тестов (`go test ./...` запускает пакеты одновременно) создают разные схемы и друг другу не мешают.
:::

### testcontainers-go

Учебная база у тебя поднята через `docker compose`, и тесты ходят в неё. В CI или в команде, где у каждого своя локальная база, удобнее, чтобы тест **сам** поднимал контейнер. Это делает библиотека testcontainers-go: она через Docker API запускает контейнер, ждёт готовности и отдаёт строку подключения. Пример из документации модуля `postgres` (в homework мы его не ставим, база уже есть):

```go
import (
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

func TestMain(m *testing.M) {
	ctx := context.Background()
	ctr, err := postgres.Run(ctx, "postgres:17-alpine",
		postgres.WithDatabase("app"),
		postgres.WithUsername("app"),
		postgres.WithPassword("secret"),
		postgres.BasicWaitStrategies(), // ждать, пока база реально примет соединения
	)
	if err != nil {
		log.Fatal(err)
	}
	dsn, _ := ctr.ConnectionString(ctx, "sslmode=disable")
	testDSN = dsn // дальше тесты создают свои схемы, как testdb
	code := m.Run()
	_ = testcontainers.TerminateContainer(ctr)
	os.Exit(code)
}
```

Плюсы: тестам не нужна заранее поднятая инфраструктура, версия PostgreSQL зафиксирована в коде, в CI достаточно Docker. Минусы: нужен Docker (на macOS это Docker Desktop, Colima или OrbStack), первый запуск скачивает образ, старт контейнера занимает секунды, поэтому контейнер поднимают один раз на пакет в `TestMain`, а не на каждый тест. У модуля есть `Snapshot` и `Restore` для быстрого отката базы между тестами.

Медленные интеграционные тесты иногда отделяют от быстрых: `if testing.Short() { t.Skip("интеграционный") }` и запуск `go test -short ./...` для быстрой проверки. В нашем курсе база всегда поднята, поэтому домашка так не делает.

## Flaky-тесты

Flaky-тест то проходит, то падает без изменений в коде. Он хуже отсутствующего: команда привыкает перезапускать CI, и настоящая поломка проходит незамеченной. Частые причины в Go-сервисах:

1. **Время.** `time.Now()` в коде, проверка «создано не позже секунды назад», тест на полночь или смену часового пояса. Лечение: часы как зависимость (`now func() time.Time`), фиксированное время в тестах.
2. **`time.Sleep` как синхронизация.** «Подождём 100 мс, горутина успеет». На нагруженном CI не успевает. Лечение: каналы, `sync.WaitGroup`, `Eventually` с опросом условия, `testing/synctest` (Go 1.25) с виртуальным временем.
3. **Порядок.** Итерация по map, SQL без `ORDER BY`, ожидание порядка от параллельных горутин. Лечение: явная сортировка или сравнение без учёта порядка.
4. **Общее состояние.** Глобальные переменные, общая таблица без изоляции, тесты зависят от порядка запуска. Лечение: своя схема или fake на тест, запуск с `-shuffle=on`.
5. **Гонки.** Тест проходит без `-race` и иногда падает с ним или на другой машине. Лечение: `go test -race -count=10`.
6. **Внешний мир.** Реальные порты (`:8080` занят), сеть, DNS, сторонние API. Лечение: `httptest.NewServer` (сам выбирает свободный порт), fake для внешних API.

Поиск: `go test -race -count=20 -shuffle=on -run TestFlaky ./...`. Если тест падает один раз из двадцати, это повод чинить, а не перезапускать.

::: trap time.Sleep вместо синхронизации
```go
// плохо
go worker.Process(task)
time.Sleep(100 * time.Millisecond)
if repo.Get(task.ID).Status != "done" { t.Error(...) }
```

```go
// хорошо: опрашиваем условие с таймаутом, тест быстрый, когда всё хорошо,
// и падает с понятным сообщением, когда плохо
go worker.Process(task)
Eventually(t, func() bool { return repo.Get(task.ID).Status == "done" }, 2*time.Second, 10*time.Millisecond)
```

Ещё лучше, если код позволяет дождаться результата напрямую (канал `done`, `WaitGroup`). `Eventually` нужен, когда наблюдать можно только внешнее состояние.
:::

::: trap Ожидание порядка без ORDER BY
```go
// плохо: в fake порядок вставки, в PostgreSQL без ORDER BY порядок не гарантирован
rows := testdb.Rows(t, pool, `SELECT title FROM tasks WHERE owner_id = 1`)
if rows[0][0] != "первая" { ... }
```

PostgreSQL обычно отдаёт строки маленькой таблицы в порядке вставки, поэтому тест зелёный месяцами. После `UPDATE` (строка физически переезжает) или на большой таблице с параллельным сканированием порядок меняется. Хорошо: `ORDER BY id` в запросе, если порядок часть контракта, или сравнение без учёта порядка, если нет.
:::

## Что выведет программа?

```go
func TestOrder(t *testing.T) {
	t.Cleanup(func() { fmt.Println("cleanup 1") })
	t.Cleanup(func() { fmt.Println("cleanup 2") })
	defer fmt.Println("defer")
	fmt.Println("тело")
}
```

<Predict :options="['тело, cleanup 1, cleanup 2, defer', 'тело, defer, cleanup 1, cleanup 2', 'тело, defer, cleanup 2, cleanup 1', 'тело, cleanup 2, cleanup 1, defer']" :correct="2">

`defer` выполняется при выходе из функции теста, а `t.Cleanup` после того, как тест и все его подтесты завершились, и в обратном порядке регистрации, как стек `defer`. Поэтому хелперы вроде `testdb.New` регистрируют закрытие пула позже удаления схемы: закрытие выполнится раньше.

</Predict>

```go
func TestParallel(t *testing.T) {
	for _, name := range []string{"a", "b", "c"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			fmt.Print(name, " ")
		})
	}
	fmt.Print("| ")
}
```

<Predict :options="['a b c |', '| и затем a, b, c в каком-то порядке', 'c c c |', '| c c c']" :correct="1">

`t.Parallel()` приостанавливает подтест до тех пор, пока родительская функция не вернётся. Поэтому сначала печатается `| `, а потом параллельные подтесты в произвольном порядке. С Go 1.22 у каждой итерации своя переменная `name`, поэтому `c c c` не будет (до 1.22 было бы). Следствие: `defer` в родительском тесте выполнится раньше параллельных подтестов, общие ресурсы освобождай через `t.Cleanup`, он ждёт подтесты.

</Predict>

::: interview
<details><summary>Как ты тестируешь сервис на Go? Расскажи про уровни.</summary>

Юнит-тесты сервисного слоя с fake-репозиториями: бизнес-правила, валидация, проброс ошибок. Тесты HTTP-слоя через `httptest` с настоящим роутером и fake-сервисами: коды ответов, формат JSON, middleware. Интеграционные тесты репозитория в настоящей PostgreSQL с изоляцией (схема на тест или testcontainers в CI). Несколько сквозных проверок запущенного сервиса. Всё с `-race` в CI.

</details>
<details><summary>Чем fake отличается от mock? Что предпочитают в Go?</summary>

Fake это рабочая упрощённая реализация (репозиторий на map), mock заранее программируется ожиданиями вызовов и проверяет их. Stub отдаёт заготовленный ответ, spy запоминает вызовы. В Go обычно предпочитают fake: интерфейсы маленькие, тесты проверяют поведение и состояние, а не порядок вызовов, поэтому не ломаются от рефакторинга. Чтобы fake не расходился с базой, на обе реализации запускают один контрактный тест.

</details>
<details><summary>Как протестировать хендлер Gin?</summary>

`gin.SetMode(gin.TestMode)`, собрать роутер функцией с зависимостями (fake-сервисы), `httptest.NewRequest` с телом и заголовками, `httptest.NewRecorder`, `router.ServeHTTP(w, req)`, проверить `w.Code`, заголовки и тело. Табличный тест на успех, ошибки валидации, ошибки сервиса (404, 409, 500), отсутствие авторизации. Для одного хендлера без роутера есть `gin.CreateTestContext`.

</details>
<details><summary>Почему не стоит мокать базу в тестах репозитория? Как изолировать такие тесты?</summary>

Логика репозитория это SQL: опечатки, `ORDER BY`, `NULL`, ограничения, `JOIN` с проверкой владельца ловит только настоящая база. Изоляция: своя схема на тест с `search_path` (параллельно и без ограничений для кода), транзакция с откатом (быстро, но код не может коммитить), `TRUNCATE` между тестами (без параллельности), база из шаблона. Базу поднимают docker compose или testcontainers-go.

</details>
<details><summary>Что такое testcontainers и когда он нужен?</summary>

Библиотека, которая из теста через Docker API поднимает контейнер (PostgreSQL, Redis, Kafka), ждёт готовности и отдаёт адрес. Удобно в CI и в команде: тестам не нужна заранее поднятая инфраструктура, версия зафиксирована в коде. Минусы: нужен Docker, старт контейнера занимает секунды, поэтому контейнер поднимают один раз в `TestMain`, а изолируют тесты схемами или snapshot/restore.

</details>
<details><summary>Как проверять JSON-ответы, чтобы тесты не были хрупкими?</summary>

Не сравнивать строки: порядок полей, пробелы и нестабильные значения (`id`, `created_at`) ломают тест без изменения смысла. Распаковывать в структуру и проверять важные поля, или сравнивать по смыслу через `any` с игнором нестабильных ключей. Для больших ответов golden-файлы с флагом `-update`: изменения контракта видны в диффе на ревью.

</details>
<details><summary>Откуда берутся flaky-тесты и как с ними бороться?</summary>

Время (`time.Now`, `time.Sleep` как синхронизация), порядок (map, SQL без `ORDER BY`), общее состояние между тестами, гонки, внешние зависимости (порты, сеть). Лечение: часы как зависимость, явная синхронизация или `Eventually`, изоляция данных на тест, `-race`, `-shuffle=on`, `-count=N` для поиска, `httptest.NewServer` вместо реальных портов. Flaky-тест чинят сразу, а не перезапускают.

</details>
<details><summary>Как протестировать middleware аутентификации?</summary>

Через роутер и `httptest`, с fake-проверкой токена и подменённым временем. Таблица: нет заголовка, не та схема, пустой токен, поддельный и просроченный токен (401), валидный (200 и правильный пользователь в контексте), не та роль (403), аноним на защищённом маршруте (401). Проверять, что на 401/403 обработчик не вызван. Качество набора проверяют мутантами: намеренно сломанный middleware должен ронять тесты.

</details>
:::

## Источники

- `RU` [Тестирование, документация Gin](https://gin-gonic.com/ru/docs/testing/): официальный пример теста роутера через `httptest`.
- `RU` [Тестирование веб-сервиса на Go, Хабр](https://habr.com/ru/articles/271239/): интеграционные тесты HTTP-сервиса с настоящей базой и миграциями goose; статья старая, но подход актуален.
- `RU` [Как не укусить себя за хвост во время написания функциональных тестов на Go, Хабр](https://habr.com/ru/companies/sberbank/articles/882060/): testcontainers с PostgreSQL и как не получить цикл импортов при общем тестовом коде.
- `EN` [Документация net/http/httptest](https://pkg.go.dev/net/http/httptest): `NewRequest`, `NewRecorder`, `NewServer`.
- `EN` [The Practical Test Pyramid, Martin Fowler](https://martinfowler.com/articles/practical-test-pyramid.html): пирамида, контрактные тесты, где проходят границы уровней.
- `EN` [Testcontainers for Go: модуль Postgres](https://golang.testcontainers.org/modules/postgres/): `postgres.Run`, стратегии ожидания, `Snapshot` и `Restore`.
- `EN` [Using Subtests and Sub-benchmarks, Go Blog](https://go.dev/blog/subtests): табличные тесты с `t.Run`, `t.Parallel` и порядок выполнения.

## Домашка

Пакет `homework/m08/l07-testing-service/`. Модель задачи и интерфейс `TaskRepo` в готовом файле `task.go`, контрактный тест в `contract_test.go`. Для ex4 нужна поднятая база: `cd homework && docker compose up -d`.

<Hw id="ex1" title="Fake-репозиторий" level="база">

Файл `ex1_fake.go`: `FakeRepo` в памяти. Он должен пройти контрактный тест, быть потокобезопасным и отдавать копии, а не ссылки на внутреннее состояние.

```bash
cd homework && go test -race -v ./m08/l07-testing-service/ -run TestFakeRepo
```

</Hw>

<Hw id="ex2" title="Сервис со статусами" level="база">

Файл `ex2_service.go`: `CanTransition`, `Service.Create` и `Service.ChangeStatus`. Тесты используют spy-репозиторий и проверяют не только результат, но и то, что при ошибке валидации в хранилище не ходили.

```bash
cd homework && go test -v ./m08/l07-testing-service/ -run 'TestCanTransition|TestService'
```

</Hw>

<Hw id="ex3" title="Сравнение JSON по смыслу" level="посложнее">

Файл `ex3_jsondiff.go`: `JSONDiff(got, want, ignore...)` сравнивает два JSON без учёта порядка ключей и пробелов, выбрасывает нестабильные поля и описывает первое отличие с путём вида `$.items[1].title`.

```bash
cd homework && go test -v ./m08/l07-testing-service/ -run TestJSONDiff
```

</Hw>

<Hw id="ex4" title="Репозиторий в настоящей PostgreSQL" level="посложнее">

Файл `ex4_pgrepo.go`: `PgRepo` на pgx. Он проходит тот же контрактный тест, что и fake, плюс проверки, которые видны только на настоящей базе: `DEFAULT now()`, `CHECK`-ограничение, кавычки и попытка инъекции. Схема в `testdata/schema.sql`.

```bash
cd homework && go test -v ./m08/l07-testing-service/ -run TestPgRepo
```

</Hw>

<Hw id="ex5" title="Набор тестов для auth middleware" level="посложнее">

Файл `ex5_suite.go`: здесь пишешь не код, а тесты. Функция `AuthSuite(t, h)` должна проверить готовый роутер со средствами аутентификации и авторизации. Проверяющий тест сначала гоняет её на правильном роутере (должна пройти), затем на восьми сломанных версиях (каждая должна быть поймана).

Формат тела ошибки в готовом роутере упрощён, в проекте используй единый формат из [урока про binding](/m08/l02-binding).

```bash
cd homework && go test -v ./m08/l07-testing-service/ -run TestAuthSuite
```

</Hw>

<Hw id="ex6" title="Eventually вместо Sleep" level="база">

Файл `ex6_eventually.go`: помощник для тестов, который опрашивает условие с интервалом до таймаута и сообщает об ошибке ровно один раз.

```bash
cd homework && go test -race -v ./m08/l07-testing-service/ -run TestEventually
```

</Hw>

<Hw id="ex7" title="testcontainers в песочнице" level="со звёздочкой">

Без тестов в курсе. Вне репозитория курса создай модуль, добавь `github.com/testcontainers/testcontainers-go/modules/postgres` и напиши `TestMain`, который поднимает `postgres:17-alpine`, и один тест, который создаёт таблицу и вставляет строку. Замерь время первого и второго запуска `go test`. Подумай, как совместить это с изоляцией «схема на тест» из `internal/testdb`.

</Hw>

## Проект модуля: Трекер задач

REST API для личного трекера задач: пользователи, проекты, задачи со статусами. Этот проект продолжится в модуле про фоновые задачи (напоминания и отчёты на River), поэтому делай его аккуратно.

Каркас лежит в `projects/tasktracker/`, это отдельный Go-модуль (`goschool/tasktracker`). В нём готовы: `main` с конфигом, логгером, миграциями и graceful shutdown; конфиг с валидацией; миграции goose; единый формат ошибок; логирование запросов с `request_id`; `/healthz`; регистрация целиком (хендлер, сервис, репозиторий) как образец; хеширование паролей; генерация refresh-токенов. Всё остальное отвечает `501 not_implemented` и ждёт тебя. Готовые куски помечены «ГОТОВЫЙ ОБРАЗЕЦ», заготовки помечены TODO с подсказками.

### Запуск

```bash
cd projects/tasktracker
cp .env.example .env      # и поменяй JWT_SECRET: openssl rand -base64 48
make db-create            # база tasktracker в контейнере из homework (порт 55432)
make run                  # миграции применятся на старте
curl -s localhost:8080/healthz
make test                 # go test -race ./...
```

### Требования

**Аутентификация**

- `POST /v1/auth/register` `{"email","password"}`: 201 `{"id","email","created_at"}`. Email нормализуется (trim, нижний регистр), пароль 8-72 байта, хранится bcrypt. Email занят: 409, некорректный email или длина пароля: 422 с `fields`. Готово.
- `POST /v1/auth/login` `{"email","password"}`: 200 `{"access_token","refresh_token","expires_in"}`. Неверный email или пароль: 401 с одинаковым сообщением и одинаковым временем ответа (для несуществующего email всё равно вызови сравнение bcrypt с фиктивным хешем, см. [урок про аутентификацию](/m08/l05-auth)).
- `POST /v1/auth/refresh` `{"refresh_token"}`: новая пара, старый refresh больше не действует (ротация). Повторное использование старого: отзыв всего семейства и 401.
- `POST /v1/auth/logout` (с access-токеном): отзыв всех refresh-токенов пользователя, 204.
- Access-токен: JWT HS256, 15 минут, `sub`, `role`, `iss`, `aud`, `exp`, `iat`, `nbf`. Проверка алгоритма, `exp`, `iss`, `aud` обязательна.

**Проекты** (все маршруты требуют `Authorization: Bearer`)

- `POST /v1/projects` `{"name"}`: 201, имя 1-100 символов.
- `GET /v1/projects`: только свои проекты, по id.
- `GET /v1/projects/:id`, `DELETE /v1/projects/:id`: чужой или несуществующий проект: 404. Удаление проекта удаляет его задачи (`ON DELETE CASCADE` уже в миграции), ответ 204.

**Задачи**

- `POST /v1/projects/:id/tasks` `{"title"}`: 201, задача в статусе `todo`. Проект чужой: 404.
- `GET /v1/projects/:id/tasks?status=todo&limit=20&after_id=0`: фильтр по статусу, keyset-пагинация по id. `limit` по умолчанию 20; не число: 400 `bad_request`; больше 100 или меньше 1: 422 `validation_failed`. Ответ `{"items":[...],"next_after_id":N}`, где `next_after_id` есть, только если страница полная.
- `GET /v1/tasks/:id`, `PATCH /v1/tasks/:id` `{"title"?, "status"?}`, `DELETE /v1/tasks/:id`.
- Переходы статуса: `todo -> in_progress`, `in_progress -> todo | done`, `done -> in_progress`. Недопустимый переход: 409. Неизвестный статус: 422 `validation_failed` с `fields.status`.
- **Доступ только владельцу проекта.** Задачу чужого проекта нельзя прочитать, изменить, удалить или создать: всегда 404. Владелец проверяется в SQL (`JOIN projects ... AND p.owner_id = $1`), `user_id` берётся только из токена.

**Общее**

- Единый формат ошибок из [урока про binding](/m08/l02-binding): `{"error":{"code":"...","message":"...","fields":{...}}}`, поле `fields` только у ошибок валидации. Коды: `bad_request` (400, тело или параметр не разобрать), `validation_failed` (422, с картой `fields`), `unauthorized` (401), `forbidden` (403), `not_found` (404), `conflict` (409), `internal` (500). Текст ошибок базы клиенту не уходит. Готово в `internal/http/errors.go`, сервис возвращает `*service.ValidationError`.
- Транзакции: репозитории принимают `repository.DBTX` (общее у пула и `pgx.Tx`), а `repository.TxManager.InTx` выполняет несколько операций сервиса в одной транзакции (готово, проверено тестом). В модуле про фоновые задачи сервис в той же транзакции, где создаёт задачу, поставит напоминание в очередь River через `InsertTx`.
- Логи slog в JSON с `request_id`, без токенов, паролей и query-строк.
- Конфиг из окружения с валидацией при старте, graceful shutdown по SIGTERM (готово, не сломай).
- Миграции goose в `migrations/`, новые изменения схемы только новыми файлами.

**Тесты**

- Сервисы: юнит-тесты с fake-репозиториями (образец: `internal/service/auth_test.go`). Обязательно: переходы статусов, валидация, логин с неверным паролем, логин с несуществующим email (spy-хешер: `Check` вызван и в этом случае), ротация и повторное использование refresh.
- Хендлеры: `httptest` с настоящим роутером и fake-сервисами (образец: `internal/http/router_test.go`). Обязательно: 401 без токена и с плохим токеном, 404 на чужие ресурсы, 400 на битый JSON, 422 с `fields` на плохой `limit` и пустой `title`, формат ошибок.
- Репозитории: настоящая PostgreSQL, своя схема на тест (образец и хелпер: `internal/repository/repository_test.go`). Обязательно: IDOR на уровне SQL (пользователь B не видит задачи пользователя A), пагинация, фильтр по статусу, каскадное удаление, два параллельных `MarkUsed` одного refresh-токена (успешен ровно один).
- `go test -race ./...` и `go vet ./...` зелёные.

**Со звёздочкой**

- Rate limit на `/v1/auth/login` и `/v1/auth/register`: 5 запросов в минуту на IP, 429 с `Retry-After`. Не забудь `SetTrustedProxies` (уже стоит в роутере).
- Роль `admin`: `GET /v1/admin/users` со списком пользователей, обычный пользователь получает 403.
- OpenAPI-спецификация всего API в `api/openapi.yaml` (можно взять за основу домашку m08/l06) и тест, который проверяет, что каждый маршрут роутера (`router.Routes()`) описан в спецификации.

### Чеклист приёмки

- [ ] `go build ./...`, `go vet ./...`, `go test -race ./...` зелёные.
- [ ] Регистрация, логин, refresh, logout работают; повторный refresh старым токеном даёт 401 и отзывает новый; два одновременных refresh одним токеном не дают две пары.
- [ ] Время ответа логина не зависит от того, существует ли email (сравни `curl -w '%{time_total}'` для чужого и несуществующего email).
- [ ] Access-токен с `alg=none`, чужим секретом, истёкший или без `exp` не принимается.
- [ ] Пароли и refresh-токены в базе только в виде хешей (проверь `make psql`, `SELECT * FROM users`).
- [ ] Пользователь B получает 404 на любые операции с проектами и задачами пользователя A, включая создание задачи в чужом проекте.
- [ ] Недопустимый переход статуса даёт 409, неизвестный статус 422.
- [ ] Пагинация: три страницы по 2 задачи из 5 задач без повторов и пропусков, фильтр по статусу работает вместе с пагинацией.
- [ ] Все ошибки в едином формате, 500 не раскрывает текст ошибки базы.
- [ ] В логах нет токенов, паролей и query-строк; у каждой строки запроса есть `request_id`.
- [ ] Ctrl+C во время долгого запроса: сервер дожидается его и выходит с логом «остановлен».
- [ ] Тесты есть на всех трёх уровнях: сервис с fake, хендлеры через `httptest`, репозиторий в базе.

### Проверка руками

Нужен `jq` (`brew install jq`), чтобы доставать токены из ответов.

```bash
API=localhost:8080/v1

# регистрация двух пользователей
curl -s -X POST $API/auth/register -H 'Content-Type: application/json' \
  -d '{"email":"alice@example.com","password":"password-alice"}'
curl -s -X POST $API/auth/register -H 'Content-Type: application/json' \
  -d '{"email":"bob@example.com","password":"password-bob"}'

# логин
ALICE=$(curl -s -X POST $API/auth/login -H 'Content-Type: application/json' \
  -d '{"email":"alice@example.com","password":"password-alice"}')
A=$(echo "$ALICE" | jq -r .access_token); AR=$(echo "$ALICE" | jq -r .refresh_token)
B=$(curl -s -X POST $API/auth/login -H 'Content-Type: application/json' \
  -d '{"email":"bob@example.com","password":"password-bob"}' | jq -r .access_token)

# неверный пароль: 401
curl -s -o /dev/null -w '%{http_code}\n' -X POST $API/auth/login -H 'Content-Type: application/json' \
  -d '{"email":"alice@example.com","password":"wrong-password"}'

# проект и задачи alice
P=$(curl -s -X POST $API/projects -H "Authorization: Bearer $A" -H 'Content-Type: application/json' \
  -d '{"name":"Дом"}' | jq -r .id)
for t in молоко хлеб кран лампочка; do
  curl -s -X POST $API/projects/$P/tasks -H "Authorization: Bearer $A" \
    -H 'Content-Type: application/json' -d "{\"title\":\"$t\"}" > /dev/null
done
curl -s "$API/projects/$P/tasks?limit=2" -H "Authorization: Bearer $A" | jq
T=$(curl -s "$API/projects/$P/tasks?limit=1" -H "Authorization: Bearer $A" | jq -r '.items[0].id')

# переходы статуса: todo -> done нельзя (409), todo -> in_progress можно (200)
curl -s -o /dev/null -w '%{http_code}\n' -X PATCH $API/tasks/$T -H "Authorization: Bearer $A" \
  -H 'Content-Type: application/json' -d '{"status":"done"}'
curl -s -o /dev/null -w '%{http_code}\n' -X PATCH $API/tasks/$T -H "Authorization: Bearer $A" \
  -H 'Content-Type: application/json' -d '{"status":"in_progress"}'

# IDOR: bob пытается прочитать, изменить и пополнить чужое: везде 404
curl -s -o /dev/null -w '%{http_code}\n' $API/tasks/$T -H "Authorization: Bearer $B"
curl -s -o /dev/null -w '%{http_code}\n' -X PATCH $API/tasks/$T -H "Authorization: Bearer $B" \
  -H 'Content-Type: application/json' -d '{"title":"взлом"}'
curl -s -o /dev/null -w '%{http_code}\n' -X POST $API/projects/$P/tasks -H "Authorization: Bearer $B" \
  -H 'Content-Type: application/json' -d '{"title":"подкинул"}'

# без токена и с мусором: 401
curl -s -o /dev/null -w '%{http_code}\n' $API/projects
curl -s -o /dev/null -w '%{http_code}\n' $API/projects -H 'Authorization: Bearer not-a-jwt'

# ротация refresh: первый обмен 200, повторный тем же токеном 401
curl -s -o /dev/null -w '%{http_code}\n' -X POST $API/auth/refresh -H 'Content-Type: application/json' \
  -d "{\"refresh_token\":\"$AR\"}"
curl -s -o /dev/null -w '%{http_code}\n' -X POST $API/auth/refresh -H 'Content-Type: application/json' \
  -d "{\"refresh_token\":\"$AR\"}"

# плохой limit: 422 validation_failed, limit не число: 400 bad_request
curl -s -o /dev/null -w '%{http_code}\n' "$API/projects/$P/tasks?limit=1000" -H "Authorization: Bearer $A"
curl -s -o /dev/null -w '%{http_code}\n' "$API/projects/$P/tasks?limit=abc" -H "Authorization: Bearer $A"
```
