# JSON API и валидация

Почти любой бэкенд, который ты будешь писать, это JSON API: фронт или другой сервис шлёт JSON, ты проверяешь его, что-то делаешь и отвечаешь JSON. Звучит просто, и именно поэтому здесь много тихих багов. Опечатка клиента в имени поля превращается в пустое значение без ошибки. Гигабайтное тело кладёт сервер по памяти. ID больше 2^53 теряет точность на фронте. Ошибка базы с паролем в тексте улетает клиенту в 500.

В этом уроке разбираем `encoding/json` с точки зрения сервера: теги и их опции, строгий разбор тела запроса, валидацию руками и единый формат ошибок со статусами 400, 404, 409, 422 и 500. В конце смотрим на `encoding/json/v2`, который в Go 1.27 стал частью стандартной библиотеки, и на то, что изменилось внутри старого пакета.

Валидацию через библиотеку `validator/v10` разберём в модуле про Gin. Здесь пишем руками: так понятно, что именно проверяется и почему.

## encoding/json: основы

```go
type User struct {
	ID    int64  `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

b, err := json.Marshal(User{ID: 1, Email: "a@b.ru", Name: "Ира"})
// {"id":1,"email":"a@b.ru","name":"Ира"}

var u User
err = json.Unmarshal([]byte(`{"id":2,"email":"c@d.ru"}`), &u)
// u = {ID:2 Email:c@d.ru Name:""}
```

Правила, которые надо знать:

- В JSON попадают только **экспортируемые** поля (с заглавной буквы). Неэкспортируемые пакет `json` не видит через рефлексию.
- Без тега имя в JSON совпадает с именем поля (`Email`). С тегом `json:"email"` используется имя из тега.
- При `Unmarshal` в v1-семантике имена сравниваются **без учёта регистра**: `"EMAIL"` тоже заполнит `Email`.
- Поля, которых нет в JSON, остаются как были (у свежей переменной это нулевые значения).
- Поля из JSON, которых нет в структуре, **молча игнорируются**.
- `Unmarshal` принимает указатель. Передал значение: ошибка `json: Unmarshal(non-pointer ...)`.

Соответствие типов: объект JSON в структуру или `map[string]T`, массив в срез, строка в `string`, число в любой числовой тип (но дробное в `int` это ошибка), `true`/`false` в `bool`, `null` оставляет значение как есть (а указатель, срез, мапу делает `nil`).

### Теги и их опции

```go
type Product struct {
	ID        int64      `json:"id,string"`              // число как строка
	Name      string     `json:"name"`
	Tags      []string   `json:"tags,omitempty"`         // нет тегов: поля нет
	Discount  *int       `json:"discount,omitempty"`     // nil: поля нет; &0: "discount":0
	CreatedAt time.Time  `json:"created_at,omitzero"`    // нулевое время: поля нет
	Password  string     `json:"-"`                      // никогда
}
```

- **`omitempty`**: пропустить поле, если значение «пустое»: `false`, `0`, `""`, `nil`-указатель или интерфейс, срез, мапа или массив длины 0. Структуры пустыми **не считаются никогда**, поэтому `time.Time{}` с `omitempty` всё равно выводится как `"0001-01-01T00:00:00Z"`.
- **`omitzero`** (Go 1.24): пропустить поле, если оно равно нулевому значению своего типа, а если у типа есть метод `IsZero() bool`, то когда он возвращает `true`. Для `time.Time` и своих структур это то, что нужно. Можно указать обе опции: поле опустится, если сработала любая.
- **`-`**: поле не участвует ни в выводе, ни в разборе. Для паролей и внутренних полей. (Если нужно поле с именем `"-"`, тег пишется `json:"-,"`.)
- **`string`**: число или `bool` кодируется строкой: `"id":"9007199254740993"`.

Зачем `string` для ID: в JavaScript все числа это `float64`, у которого точных целых только до 2^53. Snowflake-ID и другие большие `int64` фронт молча округлит, и пользователь откроет чужой заказ. Поэтому большие ID отдают строкой.

::: trap omitempty и «ноль это тоже значение»
```go
// плохо: цена 0 (бесплатно) пропадает из ответа
type Item struct {
	Price int `json:"price,omitempty"`
}
// {"price":0} не будет, будет {}; фронт решит, что цены нет
```

`omitempty` не отличает «не задано» от «задано нулём». Если ноль осмыслен, либо не ставь `omitempty`, либо сделай поле указателем: `nil` значит «не задано», `&0` значит ноль.

```go
// хорошо
type Item struct {
	Price    int  `json:"price"`              // всегда есть
	Discount *int `json:"discount,omitempty"` // nil: скидки нет
}
```

То же при разборе запроса: `{"age":0}` и `{}` дают одинаковое `Age: 0`. Если для PATCH надо отличать «не прислали» от «прислали ноль», поле делают указателем.
:::

### Числа, any и null

Если разбирать в `any` (`map[string]any`), JSON-числа становятся `float64`:

```go
var m map[string]any
json.Unmarshal([]byte(`{"id": 9007199254740993}`), &m)
fmt.Println(m["id"])
```

<Predict :options="['9007199254740993', '9.007199254740992e+15', '9.007199254740993e+15', 'ошибка: число слишком большое']" :correct="1">

В `any` число попадает как `float64`, а `9007199254740993` это 2^53 + 1: ближайшее представимое `float64` это 2^53 = `9007199254740992`, и `fmt` печатает его в экспоненциальной форме. Точность потеряна молча. Лечится разбором в типизированную структуру с `int64` или `decoder.UseNumber()`, который кладёт число как `json.Number` (строку).

</Predict>

Ещё две детали вывода. `nil`-срез кодируется как `null`, пустой срез как `[]`. Фронт, который делает `items.map(...)`, на `null` упадёт, поэтому в ответах списков инициализируй срез: `items := []Item{}`. И `json.Marshal`, и `json.Encoder` по умолчанию экранируют `<`, `>`, `&` в последовательности `\u003c`, `\u003e`, `\u0026`, чтобы JSON можно было безопасно вставить в HTML: `{"x":"<a&b>"}` превращается в `{"x":"\u003ca\u0026b\u003e"}`. Для API это безвредно, а у `Encoder` отключается через `SetEscapeHTML(false)`.

### Свои правила кодирования

Тип может управлять своим JSON, реализовав `json.Marshaler` и `json.Unmarshaler`:

```go
type Money int64 // копейки

func (m Money) MarshalJSON() ([]byte, error) {
	sign := ""
	if m < 0 {
		sign, m = "-", -m // иначе -150 станет "-1.-50"
	}
	return []byte(fmt.Sprintf(`"%s%d.%02d"`, sign, m/100, m%100)), nil
}
```

Так делают для денег, своих форматов дат, enum-ов строкой. Деньги в `float64` не хранят и не передают: `0.1 + 0.2` не равно `0.3`.

## Строгий разбор тела запроса

Наивный хендлер:

```go
// плохо
func create(w http.ResponseWriter, r *http.Request) {
	var in CreateUser
	json.NewDecoder(r.Body).Decode(&in)
	// ...
}
```

Проблем сразу пять: ошибка декодирования проигнорирована; тело любого размера (клиент пришлёт 5 ГБ, и сервер будет их читать); неизвестные поля молча игнорируются (клиент пишет `"emial"`, получает пустой email и 422 «email required», не понимая почему); `Content-Type` не проверен; мусор после JSON (`{"a":1}{"b":2}`) пройдёт.

Строгая версия:

```go
func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	ct, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || ct != "application/json" {
		return errUnsupportedMediaType // 415
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1 МБ
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(dst); err != nil {
		var tooBig *http.MaxBytesError
		switch {
		case errors.As(err, &tooBig):
			return errTooLarge // 413
		case errors.Is(err, io.EOF):
			return errEmptyBody // 400
		default:
			return badRequest(err) // 400: синтаксис, типы, неизвестные поля
		}
	}
	if err := dec.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return badRequest(errors.New("в теле больше одного JSON-значения")) // 400
	}
	return nil
}
```

Разберём куски.

- **`mime.ParseMediaType`** вместо сравнения строк: `application/json; charset=utf-8` и `Application/JSON` валидны.
- **`http.MaxBytesReader`** отдаёт не больше N байт, а на попытку прочитать больше возвращает ошибку `*http.MaxBytesError`. Заодно он говорит серверу закрыть соединение после ответа, чтобы не дочитывать гигабайты.
- **`DisallowUnknownFields`** превращает неизвестное поле в ошибку `json: unknown field "emial"`. Отдельного типа ошибки для неё в v1 нет, только текст.
- **Второй `Decode`** должен вернуть `io.EOF`: значит, после первого значения только пробелы.
- **Пустое тело** даёт `io.EOF` уже на первом `Decode`. **Обрезанный JSON** (`{"email":`) даёт `io.ErrUnexpectedEOF`, это не то же самое.

Типы ошибок, которые полезно различать: `*json.SyntaxError` (битый JSON, есть `Offset`), `*json.UnmarshalTypeError` (значение не того типа, есть `Field`). Из них получаются понятные клиенту сообщения: «поле age должно быть числом».

::: hood Почему Decoder, а не Unmarshal
`json.Unmarshal` принимает `[]byte`: тело надо сначала прочитать целиком (`io.ReadAll`), и лимит размера придётся делать отдельно. `json.Decoder` читает из `io.Reader` буферами, умеет `DisallowUnknownFields`, `UseNumber` и последовательность значений (поток JSON-объектов), поэтому для HTTP-тел его используют чаще. Обратная сторона: `Decode` читает **одно** значение и не проверяет, что дальше ничего нет. Отсюда второй `Decode`.

`Unmarshal`, наоборот, проверяет весь буфер и на `{"a":1}xyz` вернёт ошибку. Так что оба варианта корректны, если понимать разницу.
:::

## Валидация руками

После разбора JSON синтаксически всё хорошо, но данные могут быть бессмысленными. Валидация это отдельный шаг с отдельным статусом.

```go
type CreateUser struct {
	Email string `json:"email"`
	Name  string `json:"name"`
	Age   *int   `json:"age"` // необязательное
}

type FieldErrors map[string]string

func (fe FieldErrors) Error() string { return "validation failed" }

func (c CreateUser) Validate() error {
	fe := FieldErrors{}
	if strings.TrimSpace(c.Email) == "" {
		fe["email"] = "required"
	} else if !strings.Contains(c.Email, "@") {
		fe["email"] = "invalid"
	}
	if n := utf8.RuneCountInString(strings.TrimSpace(c.Name)); n == 0 {
		fe["name"] = "required"
	} else if n > 50 {
		fe["name"] = "too_long"
	}
	if c.Age != nil && (*c.Age < 0 || *c.Age > 150) {
		fe["age"] = "out_of_range"
	}
	if len(fe) == 0 {
		return nil
	}
	return fe
}
```

Принципы:

- **Собирай все ошибки сразу**, а не первую: клиенту (и форме на фронте) удобнее подсветить все поля за один заход.
- **Машинные коды** (`required`, `too_long`) вместо готовых фраз: фронт переведёт и подставит в UI.
- **Длину строк меряй в рунах**, если ограничение для человека («до 50 символов»): `len("Ёжик")` это 8 байт. Ограничение по байтам уместно для хранения (`VARCHAR` в байтах, лимит заголовка).
- **Нормализуй** перед проверкой: `TrimSpace`, приведение email к нижнему регистру. И сохраняй нормализованное значение, а не исходное.
- **Валидация не заменяет ограничения в базе.** Уникальность email проверяется уникальным индексом, а не `SELECT` перед `INSERT` (между ними успеет вклиниться другой запрос). Ошибку нарушения индекса превращаешь в 409.

::: trap nil-ошибка, которая не nil
```go
// плохо
func (c CreateUser) Validate() error {
	var fe FieldErrors // nil-мапа
	// ... ничего не нашли
	return fe // интерфейс error с типом FieldErrors и nil внутри
}

if err := in.Validate(); err != nil { // true!
	writeError(w, err) // 422 на корректный запрос
}
```

Интерфейс равен `nil`, только если в нём нет ни типа, ни значения. Возвращая типизированный `nil`, ты кладёшь в интерфейс тип `FieldErrors`, и `err != nil` истинно. Это та же ловушка «nil-интерфейс против интерфейса с nil-указателем», что во втором модуле.

```go
// хорошо: явный nil, когда ошибок нет
if len(fe) == 0 {
	return nil
}
return fe
```
:::

## Единый формат ошибок и коды

Клиенту нужен **один формат** ошибки во всех ручках, тогда фронт пишет один обработчик:

```json
{"error": {"code": "validation_failed", "message": "invalid input",
           "fields": {"email": "invalid", "name": "required"}}}
```

Есть и стандарт: RFC 9457 «Problem Details» с типом `application/problem+json` и полями `type`, `title`, `status`, `detail`. Многие API используют свой простой формат, это нормально; важна единообразность.

Какой код когда:

| Ситуация | Код | `error.code` |
|---|---|---|
| не тот `Content-Type` | 415 | `unsupported_media_type` |
| тело больше лимита | 413 | `body_too_large` |
| битый JSON, не тот тип поля, неизвестное поле | 400 | `invalid_json` |
| JSON разобран, данные не проходят проверку | 422 | `validation_failed` |
| ресурса нет | 404 | `not_found` |
| конфликт: email занят, устаревшая версия | 409 | `conflict` |
| всё остальное (баг, база упала) | 500 | `internal` |

Где живёт перевод ошибок в коды: **на границе**, в хендлере. Сервисный слой ничего не знает про HTTP и возвращает доменные ошибки:

```go
var (
	ErrNotFound = errors.New("not found")
	ErrConflict = errors.New("conflict")
)

// сервис
func (s *Service) Get(ctx context.Context, id int64) (User, error) {
	u, err := s.repo.Get(ctx, id)
	if err != nil {
		return User{}, fmt.Errorf("get user %d: %w", id, err) // обёрнуто
	}
	return u, nil
}

// граница: одна функция на все хендлеры
func writeError(w http.ResponseWriter, err error) {
	var fe FieldErrors
	switch {
	case errors.As(err, &fe):
		writeJSON(w, 422, errBody("validation_failed", "invalid input", fe))
	case errors.Is(err, ErrNotFound):
		writeJSON(w, 404, errBody("not_found", "resource not found", nil))
	case errors.Is(err, ErrConflict):
		writeJSON(w, 409, errBody("conflict", "resource already exists", nil))
	default:
		slog.Error("internal error", "err", err) // подробности в лог
		writeJSON(w, 500, errBody("internal", "internal server error", nil))
	}
}
```

`errors.Is` и `errors.As` разворачивают цепочку `%w`, поэтому обёрнутые ошибки распознаются. А в `default` клиент получает общий текст, а детали уходят в лог.

::: trap Утечка внутренностей в 500
```go
// плохо
http.Error(w, err.Error(), 500)
// pq: password authentication failed for user "admin"
// open /etc/app/secrets.yaml: permission denied
```

Текст внутренней ошибки рассказывает атакующему про устройство системы: СУБД, пути, имена пользователей, иногда сам SQL. Клиенту отдавай общий `internal server error` (можно с request ID, чтобы поддержка нашла запись в логах), а полную ошибку пиши в лог.
:::

### Запись JSON-ответа

```go
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
```

Порядок как в прошлом уроке: заголовок, статус, тело. Если `Encode` упадёт на середине (например, в структуре канал или `NaN`), статус уже ушёл, и исправить ответ нельзя. Для надёжности можно сначала `json.Marshal` в буфер, проверить ошибку и только потом писать. `Encoder` добавляет перевод строки в конце, это нормально.

## encoding/json/v2 в Go 1.27

В Go 1.25 появился экспериментальный `encoding/json/v2` (за флагом `GOEXPERIMENT=jsonv2`), а в Go 1.27 он и низкоуровневый `encoding/json/jsontext` стали обычными пакетами стандартной библиотеки. Важнее для тебя другое: **старый `encoding/json` теперь работает на движке v2**. API и поведение v1 сохранены, Unmarshal стал заметно быстрее, но **тексты ошибок могут отличаться** от старых. Код, который сравнивает текст ошибки целиком, может сломаться при обновлении. Вернуть старую реализацию можно флагом сборки `GOEXPERIMENT=nojsonv2`, но это временная мера.

v2 выбирает более строгие умолчания:

```go
import "encoding/json/v2"

var u struct {
	Email string   `json:"email"`
	Tags  []string `json:"tags"`
}
json.Unmarshal([]byte(`{"EMAIL":"x"}`), &u)            // регистр важен: поле не заполнится
json.Unmarshal([]byte(`{"email":"a","email":"b"}`), &u) // ошибка: duplicate object member name
json.Unmarshal([]byte(`{"a":1}xyz`), &u)                // ошибка: мусор после значения
b, _ := json.Marshal(u)                                 // nil-срез: "tags":[], а не null
json.Unmarshal(data, &u, json.RejectUnknownMembers(true)) // аналог DisallowUnknownFields
```

- имена полей сравниваются **с учётом регистра**;
- **дубликаты ключей** в объекте это ошибка (в v1 побеждал последний, это использовали для обхода проверок);
- невалидный UTF-8 в строках это ошибка;
- `nil`-срез и `nil`-мапа кодируются как `[]` и `{}`;
- `omitempty` в v2 означает «пропустить, если значение кодируется в пустой JSON» (`""`, `null`, `[]`, `{}`), поэтому `0` с `omitempty` в v2 **выводится**; для «пропустить ноль» в v2 есть `omitzero`;
- настройки передаются опциями в вызов, а не методами декодера.

Неизвестные поля v2 по умолчанию тоже игнорирует, строгость включается опцией. Мигрировать существующий код не обязательно: v1 API поддерживается. В новых проектах v2 стоит рассмотреть, но учти, что большинство библиотек и примеров пока на v1, и в этом курсе мы пишем на v1.

## Что выведет программа?

```go
type Event struct {
	Name    string    `json:"name,omitempty"`
	Created time.Time `json:"created,omitempty"`
	Tags    []string  `json:"tags"`
	secret  string
}

b, _ := json.Marshal(Event{secret: "s"})
fmt.Println(string(b))
```

<Predict :options="['{}', '{\'tags\':null}', '{\'created\':\'0001-01-01T00:00:00Z\',\'tags\':null}', '{\'created\':\'0001-01-01T00:00:00Z\',\'tags\':[],\'secret\':\'s\'}']" :correct="2">

Варианты записаны с одинарными кавычками для читаемости, в реальном выводе двойные. `Name` пустой и с `omitempty`: пропущен. `Created` это структура, а структуры для `omitempty` никогда не пустые, поэтому нулевое время выведено. Нужен `omitzero`. `Tags` без `omitempty`, `nil`-срез стал `null`. `secret` неэкспортируемый, `json` его не видит.

</Predict>

```go
type Signup struct {
	Email string `json:"email"`
}

var s Signup
err := json.Unmarshal([]byte(`{"emial":"me@go.dev"}`), &s)
fmt.Printf("%v %q\n", err, s.Email)
```

<Predict :options="['<nil> \'me@go.dev\'', 'json: unknown field \'emial\' \'\'', '<nil> \'\'', 'ошибка: missing field email']" :correct="2">

Неизвестное поле `emial` молча проигнорировано, `Email` остался пустым, ошибки нет. Клиент получит от твоей валидации «email required» и будет долго искать, почему, ведь email он прислал. `Decoder.DisallowUnknownFields()` превращает такую опечатку в понятную 400.

</Predict>

::: interview
<details><summary>Чем omitempty отличается от omitzero?</summary>

`omitempty` пропускает `false`, `0`, `""`, `nil` и пустые срезы, мапы, массивы; структуры не пропускает никогда, поэтому нулевой `time.Time` выводится. `omitzero` (Go 1.24) пропускает значение, равное нулевому для своего типа, а если у типа есть `IsZero() bool`, спрашивает его. Для `time.Time` и своих структур нужен `omitzero`. Обе опции вместе: поле пропускается, если сработала любая. В v2 `omitempty` переопределён как «кодируется в пустой JSON», и ноль им уже не пропускается.

</details>
<details><summary>Как отличить в запросе «поле не прислали» от «прислали нулевое значение»?</summary>

Сделать поле указателем: `nil` значит не прислали, указатель на ноль значит прислали ноль. Для PATCH-запросов это стандартный приём. Альтернативы: разбирать в `map[string]json.RawMessage` и смотреть наличие ключа, или свой тип `Optional[T]` с `UnmarshalJSON`, который запоминает факт присутствия.

</details>
<details><summary>Почему большие ID отдают строкой?</summary>

В JavaScript числа это `float64`, точно представимы целые только до 2^53. `int64` больше этого (snowflake-ID, ID из некоторых БД) фронт округлит без ошибки. Тег `json:",string"` кодирует число строкой. Та же проблема в Go при разборе в `any`: числа станут `float64`; помогает типизированная структура или `Decoder.UseNumber()`.

</details>
<details><summary>Как правильно читать JSON-тело запроса?</summary>

Проверить `Content-Type` через `mime.ParseMediaType` (иначе 415); ограничить размер `http.MaxBytesReader` (превышение даёт `*http.MaxBytesError`, отвечаем 413); `json.NewDecoder` с `DisallowUnknownFields`; проверить ошибку `Decode` (`io.EOF` значит пустое тело); вторым `Decode` убедиться, что дальше `io.EOF`, а не ещё данные. Потом валидация с 422.

</details>
<details><summary>400 или 422 на ошибку валидации?</summary>

Распространённое соглашение: 400, если запрос не удалось разобрать (синтаксис JSON, не тот тип, неизвестное поле), и 422, если разобрали, но значения не проходят правила. Некоторые API используют 400 для всего, это тоже допустимо. Главное, единообразие, единый формат тела с машинным кодом и перечислением полей.

</details>
<details><summary>Где в слоистом приложении превращать ошибки в HTTP-статусы?</summary>

На границе, в HTTP-слое: одна функция `writeError` на все хендлеры. Сервисы и репозитории возвращают доменные ошибки (`ErrNotFound`, `ErrConflict`, типы валидации), обёрнутые через `%w`. Хендлер распознаёт их `errors.Is`/`errors.As` и выбирает статус. Неизвестное становится 500 с общим текстом, подробности в лог. Так сервисный слой не зависит от HTTP и переиспользуется в gRPC, CLI, воркерах.

</details>
<details><summary>Почему validate возвращает ошибку, хотя ошибок нет?</summary>

Классическая ловушка: функция возвращает `error`, внутри типизированная переменная (`var fe FieldErrors` или `*MyErr`), равная `nil`, и она возвращается как `error`. Интерфейс содержит тип, поэтому не равен `nil`. Возвращать нужно литерал `nil`.

</details>
<details><summary>Что изменилось с JSON в Go 1.27?</summary>

`encoding/json/v2` и `encoding/json/jsontext` стали обычными пакетами (в 1.25 были экспериментом). Старый `encoding/json` теперь реализован на движке v2: поведение сохранено, разбор быстрее, но тексты ошибок могут отличаться. v2 строже по умолчанию: регистрозависимые имена, ошибка на дубликаты ключей и невалидный UTF-8, `nil`-срезы как `[]`, опции передаются в вызов. Откат на старую реализацию: `GOEXPERIMENT=nojsonv2`.

</details>
:::

## Источники

- `RU` [JSON с опциональными полями в Go, Хабр (OTUS)](https://habr.com/ru/companies/otus/articles/529318/): указатели для «не прислали», `DisallowUnknownFields`, поведение `Unmarshal`.
- `RU` [Эволюция JSON в Go: от v1 к v2, Хабр](https://habr.com/ru/articles/923404/): чем v2 отличается от v1 и зачем его сделали.
- `EN` [How to Parse a JSON Request Body in Go](https://www.alexedwards.net/blog/how-to-properly-parse-a-json-request-body): Алекс Эдвардс разбирает строгий `decodeJSON` по шагам и сообщения об ошибках.
- `EN` [JSON and Go](https://go.dev/blog/json): официальное введение в `encoding/json`, типы и теги.
- `EN` [A new experimental Go API for JSON](https://go.dev/blog/jsonv2-exp): мотивация v2 и список изменённых умолчаний.
- `EN` [Документация encoding/json](https://pkg.go.dev/encoding/json): раздел `Marshal` про опции тегов, включая `omitzero`.
- `EN` [RFC 9457: Problem Details for HTTP APIs](https://www.rfc-editor.org/rfc/rfc9457): стандартный формат ошибок, если хочешь не изобретать свой.

## Домашка

<Hw id="ex1" title="Теги" level="база">

Файл `homework/m06/l03-json-api/ex1_tags.go`: расставь теги у структуры `Product`, чтобы ID шёл строкой, ноль в цене оставался, пустые теги и `nil`-скидка пропадали, нулевое время не выводилось, а секрет не участвовал ни в выводе, ни в разборе.

```bash
cd homework && go test -v ./m06/l03-json-api/ -run TestProduct
```

</Hw>

<Hw id="ex2" title="Строгий DecodeJSON" level="посложнее">

Файл `ex2_decode.go`, функция `DecodeJSON`: `Content-Type`, лимит размера, пустое тело, неизвестные поля, мусор после объекта. Каждая ошибка это `*APIError` с нужным статусом и кодом.

```bash
cd homework && go test -v ./m06/l03-json-api/ -run TestDecodeJSON
```

</Hw>

<Hw id="ex3" title="Валидация" level="база">

Файл `ex3_validate.go`, метод `CreateUser.Validate`: все ошибки сразу, длина имени в рунах, необязательный возраст, честный `nil`, когда всё хорошо.

```bash
cd homework && go test -v ./m06/l03-json-api/ -run TestValidate
```

</Hw>

<Hw id="ex4" title="Единый формат ошибок" level="посложнее">

Файл `ex4_errors.go`: `WriteJSON` и `WriteError`, который превращает `*APIError`, `FieldErrors`, `ErrNotFound`, `ErrConflict` (в том числе обёрнутые) в ответ единого формата, а всё остальное в 500 без утечки текста.

```bash
cd homework && go test -v ./m06/l03-json-api/ -run TestWrite
```

</Hw>

<Hw id="ex5" title="Ручка целиком" level="со звёздочкой">

Собери в своём модуле хендлер `POST /users` из кусков домашки: `DecodeJSON` (лимит 1 МБ), `Validate`, хранилище в памяти с мьютексом, где повторный email даёт `ErrConflict`, ответ 201 с `Location: /users/{id}` и телом пользователя через `WriteJSON`, все ошибки через `WriteError`. Проверь curl-ом все коды из таблицы урока: 201, 400, 409, 413, 415, 422. Потом замени в хранилище обычный `encoding/json` на `encoding/json/v2` для сериализации ответа и найди, что изменилось в выводе.

</Hw>
