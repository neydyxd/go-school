# Тестируемый дизайн: интерфейсы, fake и mock, httptest

Табличные тесты хороши, пока функция чистая: получила аргументы, вернула результат. Реальный бэкенд-код почти никогда так не выглядит. Обработчик регистрации ходит в базу, сервис уведомлений шлёт письма, клиент платёжки дёргает внешний API, скидка зависит от `time.Now()`. Тест такого кода либо требует поднять базу, SMTP и платёжный шлюз, либо невозможен вовсе: как проверить «скидка по пятницам», если сегодня вторник?

Хорошая новость: проблема почти всегда не в тестах, а в дизайне. Код, который сам создаёт свои зависимости (`sql.Open` внутри функции, `http.Get` с зашитым адресом, вызов `time.Now()` в глубине логики), невозможно изолировать. Код, который **получает** зависимости снаружи через маленькие интерфейсы, тестируется легко, а заодно становится проще для чтения и замены частей.

В этом уроке разберём, как проектировать такие швы в Go, чем фейк отличается от мока и стаба (любимый вопрос собесов), когда моки вредят, и как тестировать HTTP без сети с помощью `httptest`.

## Проблема: зашитые зависимости

```go
// плохо: зависимость от реального времени зашита внутрь
func Greeting() string {
	h := time.Now().Hour()
	switch {
	case h >= 5 && h < 12:
		return "Доброе утро"
	case h >= 12 && h < 18:
		return "Добрый день"
	case h >= 18 && h < 23:
		return "Добрый вечер"
	default:
		return "Доброй ночи"
	}
}
```

Тест этой функции проходит или падает в зависимости от того, когда его запустили. Проверить границу 4:59/5:00 нельзя вообще. Такой тест называют **flaky** (нестабильным): самый ненавидимый вид тестов, потому что он учит команду игнорировать красный CI.

## Шов: зависимость через интерфейс

Решение: сделать зависимость параметром. В Go для этого есть два инструмента, оба уже знакомы по модулю про интерфейсы.

**Интерфейс**, если у зависимости есть состояние или несколько методов:

```go
type Clock interface {
	Now() time.Time
}

func Greeting(c Clock) string {
	h := c.Now().Hour()
	// ...та же логика
}
```

**Функциональный тип**, если зависимость это одна функция без состояния:

```go
type Service struct {
	now func() time.Time // в проде time.Now, в тесте замыкание с фиксированным временем
}
```

В проде передаётся настоящая реализация, в тесте подставная:

```go
type fakeClock struct{ t time.Time }

func (f fakeClock) Now() time.Time { return f.t }

func TestGreeting(t *testing.T) {
	c := fakeClock{time.Date(2026, 11, 4, 4, 59, 0, 0, time.UTC)}
	if got := Greeting(c); got != "Доброй ночи" {
		t.Errorf("Greeting(04:59) = %q, want %q", got, "Доброй ночи")
	}
}
```

Место, где поведение программы можно подменить, не меняя код, называют **швом** (seam). Интерфейс и есть шов.

### Правила хорошего шва в Go

1. **Интерфейс объявляет потребитель, а не производитель.** Пакет `postgres` возвращает конкретный `*postgres.UserRepo`. Пакет `signup`, которому нужно найти пользователя и сохранить нового, объявляет у себя `type Store interface { ByEmail(...); Save(...) }` ровно с этими двумя методами. Неявная реализация интерфейсов в Go делает это бесплатным: `*postgres.UserRepo` подходит, хотя про `signup.Store` ничего не знает.
2. **Маленькие интерфейсы.** Чем меньше методов, тем проще фейк. Интерфейс на 25 методов «на всякий случай» тянет за собой фейк на 25 методов.
3. **Принимай интерфейсы, возвращай структуры.** Конструктор `NewService(store Store, mailer Mailer) *Service` получает зависимости, а возвращает конкретный тип.
4. **Зависимости собираются в одном месте**, обычно в `main`: там создаются настоящая база, настоящий SMTP-клиент, и всё передаётся в конструкторы. Этот приём называется **внедрением зависимостей** (dependency injection) и в Go делается руками, без фреймворков.

```go
// main.go: сборка приложения
func main() {
	db := mustOpenDB(os.Getenv("DATABASE_URL"))
	repo := postgres.NewUserRepo(db)
	mailer := smtp.NewMailer(os.Getenv("SMTP_ADDR"))
	svc := signup.NewService(repo, mailer, time.Now)
	// ...
}
```

::: trap Интерфейс ради мока на стороне производителя
Плохо: пакет объявляет огромный интерфейс рядом с реализацией и возвращает его.

```go
package postgres

type UserRepository interface { // 15 методов, нужны всем по чуть-чуть
	ByID(ctx context.Context, id int64) (User, error)
	ByEmail(ctx context.Context, email string) (User, error)
	// ...ещё 13
}

func NewUserRepo(db *sql.DB) UserRepository { ... }
```

Каждый потребитель зависит от всех 15 методов, каждый фейк реализует все 15, а добавление шестнадцатого метода ломает все фейки в проекте.

Хорошо: производитель отдаёт конкретный тип, потребитель описывает только то, что использует.

```go
package postgres

func NewUserRepo(db *sql.DB) *UserRepo { ... }
```

```go
package signup

type Store interface {
	ByEmail(email string) (User, error)
	Save(u User) (User, error)
}
```
:::

## Тестовые двойники: dummy, stub, spy, mock, fake

Подставные объекты в тестах называют **тестовыми двойниками** (test doubles), по аналогии с дублёрами в кино. Классификацию ввёл Джерард Месарош, популяризировал Мартин Фаулер, и на собесах её спрашивают дословно.

| Вид | Что делает | Пример |
|---|---|---|
| **Dummy** | просто заполняет параметр, не вызывается | `nil` логгер, пустая структура |
| **Stub** | возвращает заранее заданные ответы | `ByEmail` всегда возвращает `ErrNotFound` или «db down» |
| **Spy** | stub, который ещё и запоминает вызовы, проверка после | мейлер, складывающий письма в слайс |
| **Mock** | заранее запрограммирован ожиданиями и сам проверяет, что его вызвали правильно | `m.EXPECT().Send("a@b.c", ...)` |
| **Fake** | рабочая упрощённая реализация | хранилище на `map` вместо PostgreSQL |

Главное различие на собесе: **stub и fake проверяют состояние** (что вернула функция, что лежит в хранилище после), **mock проверяет поведение** (какие методы вызваны, с какими аргументами, сколько раз). Spy посередине: записывает поведение, а проверяешь его ты обычными `if`.

### Stub и spy руками

В Go двойники обычно пишут руками: интерфейсы маленькие, и это несколько строк.

```go
type Mailer interface {
	Send(to, subject string) error
}

// spyMailer запоминает письма и может вернуть заданную ошибку.
type spyMailer struct {
	sent []string
	err  error
}

func (s *spyMailer) Send(to, subject string) error {
	s.sent = append(s.sent, to+": "+subject)
	return s.err
}

func TestWelcomeSpy(t *testing.T) {
	spy := &spyMailer{}
	if err := NewService(spy).Welcome("a@b.c"); err != nil {
		t.Fatal(err)
	}
	if len(spy.sent) != 1 || spy.sent[0] != "a@b.c: Добро пожаловать" {
		t.Errorf("отправлено %q, want одно приветствие на a@b.c", spy.sent)
	}
}

func TestWelcomeError(t *testing.T) {
	errSMTP := errors.New("smtp down")
	err := NewService(&spyMailer{err: errSMTP}).Welcome("a@b.c")
	if !errors.Is(err, errSMTP) {
		t.Errorf("Welcome() err = %v, want обёрнутую %v", err, errSMTP)
	}
}
```

Второй тест показывает, зачем двойники нужны помимо скорости: **воспроизвести сбой**. Заставить настоящий SMTP-сервер упасть ровно в нужный момент сложно, а stub возвращает ошибку по заказу.

### Fake: настоящая логика без инфраструктуры

Fake это работающая реализация, только простая: хранилище на `map`, очередь на слайсе, файловая система в памяти (`testing/fstest.MapFS` в стандартной библиотеке как раз такой). Fake ведёт себя как настоящая зависимость: сохранил, потом нашёл; сохранил дубликат, получил ошибку.

```go
type MemStore struct {
	byEmail map[string]User
	nextID  int
}

func (s *MemStore) Save(u User) (User, error) {
	if _, ok := s.byEmail[u.Email]; ok {
		return User{}, ErrExists
	}
	u.ID = s.nextID
	s.nextID++
	s.byEmail[u.Email] = u
	return u, nil
}
```

Плюс фейка: тесты проверяют **результат**, а не то, как сервис дёргал хранилище. Можно переписать сервис (сначала `Save`, а дубликат ловить по ошибке, вместо `ByEmail` + `Save`), и тесты на фейке останутся зелёными. Минус: фейк это код, который надо поддерживать и который может разойтись с настоящей реализацией. Поэтому у серьёзных фейков бывают **контрактные тесты**: один и тот же набор тестов гоняют против фейка и против настоящей базы.

### Mock и генераторы

Для больших интерфейсов используют генераторы. Два популярных варианта:

**gomock** (`go.uber.org/mock`, форк заброшенного `golang/mock`): утилита `mockgen` генерирует мок по интерфейсу.

```bash
go run go.uber.org/mock/mockgen@latest -source=notify.go -destination=mock_notify_test.go -package=notify
```

```go
func TestWelcomeGomock(t *testing.T) {
	ctrl := gomock.NewController(t)
	m := NewMockMailer(ctrl)
	m.EXPECT().Send("a@b.c", "Добро пожаловать").Return(nil)

	if err := NewService(m).Welcome("a@b.c"); err != nil {
		t.Fatal(err)
	}
	// контроллер сам проверит в Cleanup, что все EXPECT выполнились
}
```

**testify/mock** (часто с генератором `mockery`):

```go
type mockMailer struct{ mock.Mock }

func (m *mockMailer) Send(to, subject string) error {
	args := m.Called(to, subject)
	return args.Error(0)
}

func TestWelcomeMock(t *testing.T) {
	m := &mockMailer{}
	m.On("Send", "a@b.c", "Добро пожаловать").Return(nil).Once()
	if err := NewService(m).Welcome("a@b.c"); err != nil {
		t.Fatal(err)
	}
	m.AssertExpectations(t)
}
```

Моки уместны, когда **взаимодействие и есть результат**: отправили ровно одно письмо, списали деньги ровно один раз, не вызвали платёжку при невалидной корзине. Для «сохранил и потом нашёл» фейк лучше.

::: trap Тест на моках, который проверяет реализацию
Плохо: тест жёстко фиксирует, какие методы и в каком порядке вызываются.

```go
store.EXPECT().ByEmail("a@b.c").Return(User{}, ErrNotFound)
store.EXPECT().Save(User{Email: "a@b.c"}).Return(User{ID: 1, Email: "a@b.c"}, nil)
```

Ты оптимизировал `Register`: убрал лишний `ByEmail` и ловишь дубликат по ошибке `Save`. Поведение для пользователя не изменилось, а тест упал: он проверял не что делает код, а как. Такие тесты называют **хрупкими**, они тормозят рефакторинг вместо того, чтобы его страховать.

Хорошо: проверить результат на фейке.

```go
func TestRegisterSaves(t *testing.T) {
	s := NewMemStore()
	u, err := Register(s, "a@b.c")
	if err != nil || u.Email != "a@b.c" {
		t.Fatalf("Register(a@b.c) = %+v, %v, want пользователя a@b.c", u, err)
	}
	// пользователь действительно лежит в хранилище, как бы Register его туда ни положил
	got, err := s.ByEmail("a@b.c")
	if err != nil || got != u {
		t.Errorf("ByEmail(a@b.c) = %+v, %v, want %+v", got, err, u)
	}
	if _, err := Register(s, "a@b.c"); !errors.Is(err, ErrExists) {
		t.Errorf("повторный Register(a@b.c): ошибка %v, want ErrExists", err)
	}
}
```

Правило большого пальца: мокай то, что **выходит наружу** системы (письма, платежи, события в очередь), а внутренние зависимости с состоянием заменяй фейками.
:::

::: trap Nil-указатель в интерфейсе
Фейк объявлен как указатель, а функция-конструктор вернула его через интерфейс:

```go
func newStore(test bool) Store {
	var s *MemStore
	if test {
		s = &MemStore{data: map[int]string{1: "a"}}
	}
	return s // плохо: интерфейс с типом *MemStore и значением nil
}

func main() {
	s := newStore(false)
	fmt.Println(s == nil) // false
}
```

Проверка `s == nil` не сработает, а вызов метода упадёт с nil pointer dereference внутри. Это та же ловушка из урока про устройство интерфейсов: интерфейс равен `nil`, только если в нём нет ни типа, ни значения.

```go
// хорошо: возвращать nil интерфейса явно
func newStore(test bool) Store {
	if !test {
		return nil
	}
	return &MemStore{data: map[int]string{1: "a"}}
}
```
:::

## httptest: HTTP без сети

Пакет `net/http/httptest` закрывает две задачи. Подробно про `net/http` будет в модуле про HTTP, сейчас достаточно двух фактов: сервер обрабатывает запросы через интерфейс `http.Handler` с методом `ServeHTTP(w http.ResponseWriter, r *http.Request)`, а клиент шлёт запросы через `*http.Client`.

### Тестируем обработчик: NewRecorder

`httptest.NewRecorder()` это `ResponseWriter`, который пишет ответ в память. Обработчик вызывается напрямую, без сокетов:

```go
func TestStatusHandler(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/status", nil)

	StatusHandler(fakeClock{...}).ServeHTTP(rec, req)

	res := rec.Result()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("код %d, want 200", res.StatusCode)
	}
	if ct := res.Header.Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type %q, want application/json", ct)
	}
	// rec.Body.String() это тело ответа
}
```

`httptest.NewRequest` удобнее `http.NewRequest`: не возвращает ошибку (паникует на кривом вводе, что в тесте нормально) и заполняет поля, которые сервер обычно ставит сам.

::: hood Почему rec.Result(), а не rec.Header()
В настоящем HTTP-сервере заголовки уходят клиенту в момент первой записи тела или вызова `WriteHeader`. Всё, что ставится в `w.Header()` после этого, клиент не увидит. Рекордер повторяет это поведение через снимок: в момент первой записи он копирует заголовки, и `rec.Result()` возвращает ответ именно со снимком. А `rec.Header()` это живая мапа, в которой видны и поздние изменения.

```go
h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("упс"))
	w.Header().Set("X-Late", "1")       // поздно: ответ уже начался
	w.WriteHeader(http.StatusInternalServerError) // тоже поздно
})
rec := httptest.NewRecorder()
h.ServeHTTP(rec, httptest.NewRequest("GET", "/", nil))
fmt.Println(rec.Code, rec.Header().Get("X-Late") == "", rec.Result().Header.Get("X-Late") == "")
```

Тест на `rec.Header()` пропустит баг «заголовок ставится после тела», а тест на `rec.Result().Header` поймает его, как поймал бы реальный клиент.
:::

<Predict :options="['500 false false', '200 false true', '500 true true', '200 true false']" :correct="1">

`w.Write` без явного `WriteHeader` отправляет статус 200, поэтому поздний `WriteHeader(500)` игнорируется (настоящий сервер напишет в лог `superfluous response.WriteHeader call`), и `rec.Code == 200`. Живая мапа `rec.Header()` содержит `X-Late`, сравнение с пустой строкой даёт `false`. В снимке `Result().Header` заголовка нет: `true`.

</Predict>

### Тестируем клиента: NewServer

Обратная задача: код, который **ходит** в чужой API. `httptest.NewServer(handler)` поднимает настоящий HTTP-сервер на `127.0.0.1` со случайным свободным портом. Адрес лежит в `srv.URL` (что-то вроде `http://127.0.0.1:62741`).

```go
func TestFetchProfile(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/profiles/42" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"id": 42, "name": "Гофер"}`)
	}))
	defer srv.Close()

	got, err := FetchProfile(srv.Client(), srv.URL, 42)
	// проверки
}
```

Чтобы это работало, клиентский код должен принимать **базовый адрес** и, желательно, `*http.Client` снаружи. Функция с зашитым `http.Get("https://api.example.com/...")` не тестируется без сети, и это опять вопрос дизайна, а не тестов. Сервер в тесте может отдать что угодно: 500, обрезанный JSON, медленный ответ. Так проверяются ветки обработки ошибок, которые на реальном API не воспроизвести.

`httptest.NewTLSServer` делает то же по HTTPS с самоподписанным сертификатом; `srv.Client()` уже настроен ему доверять.

::: hood Чем NewServer отличается от NewRecorder
`NewRecorder` это структура в памяти: нет сокетов, нет горутин сервера, нет сериализации HTTP. Самый быстрый вариант для юнит-тестов обработчиков.

`NewServer` это настоящий `http.Server` на настоящем TCP-листенере `127.0.0.1:0` (порт 0 значит «дай любой свободный»). Запрос проходит весь путь: клиент сериализует его в байты, ядро передаёт через loopback, сервер парсит, ответ идёт обратно. Медленнее, зато проверяет и то, что рекордер не умеет: заголовки, которые ставит транспорт, редиректы, таймауты клиента, повторное использование соединений. `srv.Close()` закрывает листенер и ждёт завершения активных запросов, поэтому его вызывают в `defer` или `t.Cleanup`.
:::

## Тесты уровней: юнит и интеграционные

Двойники дают быстрые изолированные **юнит-тесты**. Но фейк хранилища не поймает ошибку в SQL-запросе, поэтому настоящий репозиторий проверяют **интеграционными тестами** против реальной PostgreSQL (обычно в Docker, например через библиотеку testcontainers). Их отделяют от быстрых тестов:

```go
func TestUserRepo(t *testing.T) {
	if testing.Short() {
		t.Skip("интеграционный тест: нужен PostgreSQL")
	}
	// ...
}
```

```bash
go test -short ./...   # быстрые тесты на каждое сохранение
go test ./...          # всё, включая интеграционные, в CI
```

Распространённая пропорция («пирамида тестов»): много быстрых юнит-тестов, меньше интеграционных, совсем немного сквозных. Интеграционные тесты с базой разберём в модуле про базы данных.

::: trap Тест, который ходит в интернет
Плохо: юнит-тест обращается к реальному внешнему API.

```go
func TestFetchRate(t *testing.T) {
	rate, err := FetchRate("https://api.exchangerate.example/v1/usd") // сеть, чужой сервер
	if err != nil || rate <= 0 {
		t.Fatal(err)
	}
}
```

Он медленный, падает без интернета и когда у провайдера плановые работы, и не может проверить обработку ошибок. Хорошо: `httptest.NewServer` с заранее заданными ответами, а проверку настоящего API, если она нужна, вынести в отдельный тест за `testing.Short()` или в мониторинг.
:::

## Что выведет программа?

```go
type Store interface{ Get(id int) string }

type MemStore struct{ data map[int]string }

func (m *MemStore) Get(id int) string { return m.data[id] }

func newStore(test bool) Store {
	var s *MemStore
	if test {
		s = &MemStore{data: map[int]string{1: "a"}}
	}
	return s
}

func main() {
	fmt.Println(newStore(false) == nil)
}
```

<Predict :options="['true', 'false', 'паника: nil pointer dereference', 'ошибка компиляции']" :correct="1">

Возвращается интерфейс, внутри которого тип `*MemStore` и значение `nil`. Интерфейс равен `nil` только когда пуст и тип, и значение, поэтому `false`. Паники нет, пока не вызван метод: `Get` на nil-указателе упадёт на обращении к `m.data`.

</Predict>

```go
rec := httptest.NewRecorder()
fmt.Println(rec.Code)
```

<Predict :options="['0', '200', '-1', 'паника: ответ ещё не записан']" :correct="1">

`NewRecorder` создаёт рекордер с `Code: 200`: это статус по умолчанию, если обработчик ничего не вызвал. Значит тест «код равен 200» пройдёт и на обработчике, который вообще ничего не делает. Проверяй ещё и тело.

</Predict>

::: interview
<details><summary>Что такое внедрение зависимостей и как его делают в Go?</summary>

Код получает свои зависимости (хранилище, клиент, часы, логгер) снаружи, а не создаёт сам. В Go это обычно параметры конструктора с типом маленького интерфейса или функции: `NewService(store Store, mailer Mailer, now func() time.Time)`. Вся сборка происходит в `main`. Фреймворки (wire, fx) существуют, но для большинства сервисов ручная сборка проще и прозрачнее.

</details>
<details><summary>Где объявлять интерфейс: рядом с реализацией или у потребителя?</summary>

У потребителя. Интерфейсы в Go реализуются неявно, поэтому потребитель описывает ровно те методы, которые ему нужны, а производитель возвращает конкретный тип. Это даёт маленькие интерфейсы, простые фейки и отсутствие зависимостей между пакетами из-за интерфейса. Исключение: интерфейс, который и есть продукт пакета (`io.Reader`, `http.Handler`).

</details>
<details><summary>Чем mock отличается от stub и fake?</summary>

Stub возвращает заготовленные ответы. Fake это рабочая упрощённая реализация (хранилище на map). Оба используются для проверки состояния: что вернула функция, что оказалось в хранилище. Mock заранее программируется ожиданиями вызовов и сам проверяет, что его вызвали с нужными аргументами нужное число раз, то есть проверяет поведение. Spy записывает вызовы, а проверяет их тест. Dummy только заполняет параметр.

</details>
<details><summary>Когда моки вредят?</summary>

Когда тест проверяет внутреннюю реализацию: порядок и состав вызовов зависимостей, которые не являются наблюдаемым результатом. Такие тесты падают при каждом рефакторинге без изменения поведения и проходят, если поведение сломано, но вызовы те же. Моки хороши для эффектов наружу (письма, платежи, события), для внутреннего состояния лучше фейки.

</details>
<details><summary>Как протестировать HTTP-обработчик?</summary>

Создать запрос `httptest.NewRequest`, рекордер `httptest.NewRecorder()`, вызвать `handler.ServeHTTP(rec, req)` и проверить `rec.Result()`: статус, заголовки, тело. Без сети и без сервера. Зависимости обработчика (хранилище, часы) подставить через конструктор.

</details>
<details><summary>Как протестировать код, который ходит во внешний API?</summary>

Принимать базовый URL и `*http.Client` снаружи, а в тесте поднять `httptest.NewServer` с обработчиком, который отдаёт нужные ответы: успех, 404, 500, битый JSON, медленный ответ. Передать `srv.URL` и `srv.Client()`. Сервер настоящий (TCP на loopback), поэтому проверяется и сериализация, и обработка ошибок. Закрыть `srv.Close()`.

</details>
<details><summary>Как тестировать код, зависящий от времени?</summary>

Не вызывать `time.Now()` в логике напрямую, а получать часы как зависимость: интерфейс `Clock` или поле `now func() time.Time`. В тесте подставить фиксированное время и проверить границы. Для конкурентного кода с таймерами и `time.Sleep` с Go 1.25 есть пакет `testing/synctest`: внутри «пузыря» время виртуальное и прыгает вперёд, когда все горутины ждут.

</details>
<details><summary>Чем юнит-тест отличается от интеграционного и как их разделить?</summary>

Юнит-тест проверяет кусок логики в изоляции, зависимости заменены двойниками; быстрый и детерминированный. Интеграционный проверяет взаимодействие с реальной инфраструктурой (база, очередь); медленнее и требует окружения. Разделяют через `testing.Short()` и `go test -short`, build-теги (`//go:build integration`) или отдельные пакеты. Фейк не заменяет интеграционный тест репозитория: он не знает, что твой SQL с ошибкой.

</details>
:::

## Источники

- `RU` [Тестируем веб приложение на Go (Хабр, OTUS)](https://habr.com/ru/companies/otus/articles/930136/): слоистое приложение, зависимости через интерфейсы и моки на testify/mock.
- `EN` [Mocks Aren't Stubs, Мартин Фаулер](https://martinfowler.com/articles/mocksArentStubs.html): первоисточник классификации двойников и спора «проверка состояния против проверки поведения».
- `EN` [Learn Go with Tests: Dependency Injection](https://quii.gitbook.io/learn-go-with-tests/go-fundamentals/dependency-injection): внедрение зависимостей через `io.Writer` на простом примере.
- `EN` [Learn Go with Tests: Mocking](https://quii.gitbook.io/learn-go-with-tests/go-fundamentals/mocking): шпионы руками и предупреждение о переусердствовании с моками.
- `EN` [Документация net/http/httptest](https://pkg.go.dev/net/http/httptest): `NewRecorder`, `NewServer`, `NewRequest`; прочитай описание `ResponseRecorder.Result`.
- `EN` [uber-go/mock (gomock)](https://github.com/uber-go/mock): README с установкой `mockgen` и режимами генерации.

## Домашка

Пакет `homework/m03/l03-testable-design/`.

<Hw id="ex1" title="Приветствие с подменными часами" level="база">

Файл `ex1_greeting.go`, функция `Greeting(c Clock) string`. Тест подставляет фейковые часы и проверяет все границы суток.

```bash
cd homework && go test -v ./m03/l03-testable-design/ -run TestGreeting
```

</Hw>

<Hw id="ex2" title="Фейк хранилища и регистрация" level="посложнее">

Файл `ex2_store.go`: реализуй фейк `MemStore` (хранилище на `map`) и функцию `Register(s Store, email string)`. Тесты проверяют фейк сам по себе, регистрацию поверх фейка и обработку сбоев через стаб `brokenStore`, который возвращает ошибки по заказу. Посмотри в тестах, чем стаб отличается от фейка.

```bash
cd homework && go test -v ./m03/l03-testable-design/ -run 'TestMemStore|TestRegister'
```

</Hw>

<Hw id="ex3" title="Клиент внешнего API" level="посложнее">

Файл `ex3_client.go`, функция `FetchProfile(c *http.Client, baseURL string, id int) (Profile, error)`. Тест поднимает `httptest.NewServer`, который отдаёт успех, 404, 500 и обрезанный JSON, а ещё проверяет заголовок `Accept`. Клиент в тест передаётся со шпионским транспортом (`http.RoundTripper`): он проверяет, что запрос ушёл именно через переданный `c`, а не через `http.Get`, и что тело каждого ответа закрыто.

```bash
cd homework && go test -v ./m03/l03-testable-design/ -run TestFetchProfile
```

</Hw>

<Hw id="ex4" title="Обработчик статуса" level="посложнее">

Файл `ex4_handler.go`, функция `StatusHandler(c Clock) http.Handler`. Тест вызывает обработчик через `httptest.NewRecorder` и проверяет код, заголовки (через `rec.Result()`, то есть «как увидит клиент») и JSON-тело.

```bash
cd homework && go test -v ./m03/l03-testable-design/ -run TestStatusHandler
```

</Hw>

<Hw id="ex5" title="Проверка-функция со шпионом" level="со звёздочкой">

Файл `ex5_remind.go`: готовая функция `Remind` рассылает напоминания о конце подписки. Напиши `CheckRemind(remind RemindFunc) error`: заведи своего шпиона, реализующего `Mailer` (запоминает вызовы и падает на заданном адресе), подбери подписчиков со всеми интересными случаями и сверь результат с правилами. Тест прогонит твою проверку против двух правильных реализаций и против одиннадцати мутантов. Внутри `CheckRemind` не вызывай `Remind` и не сравнивай с ней: проверка должна опираться на правила, иначе это не тест. Это формат «проверка-функция» из урока про `testing`.

```bash
cd homework && go test -v ./m03/l03-testable-design/ -run TestCheckRemind
```

</Hw>

<Hw id="ex6" title="Рефакторинг к тестируемости" level="со звёздочкой">

Без автопроверки. Возьми программу из мини-проекта модуля 01 или 02 (или любую свою) и найди в ней зашитые зависимости: `time.Now()`, `os.Getenv`, чтение файлов по фиксированному пути, `os.Stdout`. Вынеси каждую в параметр (интерфейс, функцию или `io.Reader`/`io.Writer`), собери всё в `main` и напиши табличный тест на логику, который раньше был невозможен.

</Hw>
