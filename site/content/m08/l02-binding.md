# Binding, валидация, ответы

Любой обработчик начинается одинаково: разобрать вход, проверить его, ответить ошибкой, если что-то не так. На stdlib это `json.NewDecoder(r.Body).Decode(&req)` и пачка `if req.Name == ""`. В Gin разбор и проверка описываются **тегами структуры** и выполняются одной строкой. Это главная причина, по которой берут фреймворк.

Но в binding Gin спрятано несколько ловушек, на которых валятся новички и даже код в продакшне: `Bind` сам пишет 400 и ломает твой ответ, `required` не пропускает ноль, ошибки валидации по умолчанию выглядят как `Key: 'In.Email' Error:Field validation for 'Email' failed on the 'email' tag`, а пустой список уходит клиенту как `null`.

В этом уроке: семейства `Bind*` и `ShouldBind*`, теги validator/v10, кастомные правила, перевод ошибок в понятный JSON на русском, единый формат ответов и пагинация.

## ShouldBind против Bind

У Gin два семейства методов привязки:

| Семейство | Примеры | При ошибке |
|---|---|---|
| `ShouldBind*` | `ShouldBind`, `ShouldBindJSON`, `ShouldBindQuery`, `ShouldBindUri`, `ShouldBindHeader` | только возвращает `error`, ответ решаешь ты |
| `Bind*` (`MustBindWith`) | `Bind`, `BindJSON`, `BindQuery`, `BindUri` | возвращает `error` **и** сам вызывает `c.AbortWithError(400, err)` (или 413, если тело упёрлось в `http.MaxBytesReader`), который тут же записывает заголовок |

Почти всегда нужен `ShouldBind*`. Вот что происходит с `Bind`:

```go
r.POST("/users", func(c *gin.Context) {
	var in CreateUser
	if err := c.Bind(&in); err != nil {
		c.JSON(422, gin.H{"error": err.Error()})
		return
	}
	c.JSON(201, in)
})
// POST /users с телом {"name":"a"} (не проходит валидацию)
```

<Predict :options="['422 и наш JSON', '400 и наш JSON', '400 и пустое тело', '422 и пустое тело']" :correct="1">

`Bind` внутри зовёт `AbortWithStatus(400)`, а тот сразу пишет заголовки (`WriteHeaderNow`). Наш `c.JSON(422, ...)` уже не может сменить статус: Gin в debug-режиме печатает `[WARNING] Headers were already written. Wanted to override status code 400 with 422`, а тело дописывает. Клиент видит 400 с телом, в котором ты собирался отдать 422. Плюс ошибка попадает в `c.Errors` с типом `ErrorTypeBind`. С `ShouldBindJSON` ответ полностью твой.

</Predict>

### Какой ShouldBind выбрать

```go
c.ShouldBindJSON(&body)    // тело как JSON, Content-Type не важен
c.ShouldBindQuery(&q)      // только query-строка, теги form:"..."
c.ShouldBindUri(&p)        // параметры пути, теги uri:"..."
c.ShouldBindHeader(&h)     // заголовки, теги header:"..."
c.ShouldBind(&x)           // движок выбирается по методу и Content-Type
```

`ShouldBind` без суффикса удобен для форм (`application/x-www-form-urlencoded`, `multipart/form-data`), но для JSON API бери явный `ShouldBindJSON`: клиент, забывший заголовок `Content-Type`, не уйдёт молча в другой движок.

Тело запроса это поток, его можно прочитать один раз. Второй `ShouldBindJSON` в том же запросе вернёт `EOF`. Если тело нужно разобрать дважды (например, сначала узнать тип события, потом привязать к нужной структуре), есть `ShouldBindBodyWith(&x, binding.JSON)`: он читает тело целиком и кладёт байты в контекст.

## Теги и validator/v10

Валидацию в Gin делает библиотека `github.com/go-playground/validator/v10`. Gin создаёт свой экземпляр валидатора и настраивает его читать тег `binding` (у «голого» validator по умолчанию тег `validate`):

```go
type CreateProduct struct {
	Name     string   `json:"name" binding:"required,min=2,max=100"`
	Email    string   `json:"email" binding:"omitempty,email"`
	Price    int64    `json:"price" binding:"required,gt=0"`
	Category string   `json:"category" binding:"required,oneof=books games music"`
	Tags     []string `json:"tags" binding:"max=5,dive,min=2,max=20"`
}
```

Правила через запятую, проверяются слева направо:

| Тег | Смысл |
|---|---|
| `required` | значение не равно нулевому для типа |
| `omitempty` | если значение нулевое, остальные правила пропускаются |
| `min`, `max` | для строк длина в **символах** (рунах), для слайсов и map число элементов, для чисел значение |
| `gt`, `gte`, `lt`, `lte` | больше, не меньше, меньше, не больше |
| `oneof=a b c` | одно из значений через пробел |
| `email`, `url`, `uuid` | формат строки |
| `dive` | правила **после** него применяются к каждому элементу слайса или map |

Про `dive`: `binding:"max=5,dive,min=2"` значит «в слайсе не больше 5 элементов, каждый длиной от 2 символов». Ошибка элемента придёт с индексом в имени поля: `tags[1]`.

Про длину строк: `min=2` для `"Ян"` выполнится, хотя в UTF-8 это 4 байта. validator считает руны через `utf8.RuneCountInString`, так что для кириллицы `max=100` значит 100 букв.

### required и нулевые значения

```go
type Order struct {
	Quantity int  `json:"quantity" binding:"required"`
	Gift     *int `json:"gift" binding:"required"`
}
// POST /orders с телом {"quantity":0,"gift":0}
```

<Predict :options="['201: оба поля переданы', '422: не прошёл quantity', '422: не прошёл gift', '422: не прошли оба']" :correct="1">

`required` проверяет не «поле есть в JSON», а «значение не нулевое». После `json.Unmarshal` у `Quantity` ноль, и для валидатора это то же самое, что поле не передали. С указателем иначе: `Gift` указывает на ноль, а `required` у указателя проверяет только `!= nil`. Отсюда правило: если ноль (или `false`) допустимое значение, а поле обязательно, делай его указателем.

</Predict>

::: trap required на bool и числе
Плохо: `Active bool binding:"required"`. Клиент присылает `"active": false`, и запрос отвергается с «обязательное поле». Это классическая жалоба «не могу выключить флаг».

Хорошо: `Active *bool binding:"required"`, и в коде `*in.Active`. Или вообще без `required`, если отсутствие поля и `false` для тебя одно и то же.
:::

### Неизвестные поля

По умолчанию лишние поля JSON молча игнорируются: тело `{"quantity":1,"gift":0,"admin":true}` пройдёт, поле `admin` просто никуда не попадёт. Это безопасно, пока ты не привязываешь тело **прямо к доменной или БД-структуре**: тогда клиент может прислать `"role":"admin"` или `"balance":1000000` и записать то, чего не должен (mass assignment). Защита: отдельная структура запроса только с разрешёнными полями. Включить строгий режим глобально можно через `binding.EnableDecoderDisallowUnknownFields = true`, но это глобальная переменная пакета, и она ломает обратную совместимость для старых клиентов. Отдельный DTO надёжнее.

## Кастомные правила и имена полей

Валидатор Gin доступен через `binding.Validator.Engine()`. На нём регистрируют свои теги и функцию, которая называет поля в ошибках:

```go
var slugRe = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

func SetupValidator(v *validator.Validate) error {
	// В ошибках поле будет называться как в JSON: "email", а не "Email".
	v.RegisterTagNameFunc(func(f reflect.StructField) string {
		name, _, _ := strings.Cut(f.Tag.Get("json"), ",")
		if name == "-" {
			return ""
		}
		return name
	})
	return v.RegisterValidation("slug", func(fl validator.FieldLevel) bool {
		return slugRe.MatchString(fl.Field().String())
	})
}

func main() {
	v, ok := binding.Validator.Engine().(*validator.Validate)
	if !ok {
		log.Fatal("неожиданный валидатор")
	}
	if err := SetupValidator(v); err != nil {
		log.Fatal(err)
	}
	// ... роутер
}
```

Теперь `Slug string binding:"required,slug"` работает как встроенное правило. Регистрируй до того, как сервер начнёт принимать запросы: валидатор Gin общий на весь процесс.

::: hood Что такое ValidationErrors
Ошибка из `ShouldBind*` бывает двух видов:
- **ошибка разбора**: `unexpected EOF`, `EOF` (пустое тело), `json: cannot unmarshal string into Go struct field ... of type int64`, `strconv.ParseInt: parsing "abc"` для query и uri. Валидатор до неё не дошёл;
- **ошибка валидации**: тип `validator.ValidationErrors`, это `[]validator.FieldError`. У каждой `FieldError` есть `Tag()` (`"min"`), `Param()` (`"2"`), `Kind()` (`reflect.String`), `Field()` (имя по `RegisterTagNameFunc`), `Namespace()` (полный путь: `CreateProduct.address.city`, `CreateProduct.tags[1]`).

Различаешь их через `errors.As(err, &ve)`. Первое это 400 Bad Request (запрос сломан), второе 422 Unprocessable Content (запрос понятен, но данные нарушают правила). Многие API отвечают 400 на оба, это тоже допустимо, главное, чтобы было одинаково во всём сервисе и в ответе было видно, какое поле не так.
:::

### Перевод ошибок в понятный JSON

Сырой `err.Error()` клиенту отдавать нельзя: он на английском, с именами Go-структур и тегами. Нужен ответ вида:

```json
{
  "error": {
    "code": "validation_failed",
    "message": "проверь поля запроса",
    "fields": {
      "name": "минимальная длина 2",
      "tags[1]": "минимальная длина 2",
      "category": "одно из: books, games, music"
    }
  }
}
```

Переводчик проходит по `ValidationErrors` и строит сообщение по тегу:

```go
func FieldErrors(err error) map[string]string {
	var ve validator.ValidationErrors
	if !errors.As(err, &ve) {
		return nil
	}
	out := make(map[string]string, len(ve))
	for _, fe := range ve {
		key := fe.Namespace() // "CreateProduct.tags[1]"
		if _, rest, ok := strings.Cut(key, "."); ok {
			key = rest // без имени структуры: "tags[1]"
		}
		out[key] = message(fe)
	}
	return out
}

func message(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "обязательное поле"
	case "email":
		return "некорректный email"
	case "oneof":
		return "одно из: " + strings.Join(strings.Fields(fe.Param()), ", ")
	// min и max зависят от fe.Kind(): строка, слайс или число
	}
	return "не прошло проверку " + fe.Tag()
}
```

У анонимной структуры (`var in struct{...}` прямо в обработчике) имени типа в `Namespace()` нет, поэтому ключ без точки оставляем как есть, иначе все ошибки схлопнутся в пустой ключ. Ключ берём из `Namespace()`, а не из `Field()`: у вложенной структуры `Field()` даст просто `city`, и клиент не поймёт, это адрес доставки или адрес оплаты. Полную версию переводчика напишешь в домашке.

У validator есть и готовые переводы (`github.com/go-playground/validator/v10/translations/ru` вместе с `universal-translator`). Они покрывают все теги, но тянут ещё два пакета и требуют регистрации переводчика; свой `switch` на десяток тегов проще читать и проще поменять формулировку под продукт. Оба варианта нормальные.

## Единый формат ответов

Клиенту (фронтенду, мобильному приложению, другому сервису) важнее всего предсказуемость: любая ошибка любой ручки выглядит одинаково. Популярные варианты:

1. **Обёртка** (envelope): успех `{"data": ..., "meta": ...}`, ошибка `{"error": {"code", "message", "fields"}}`. Просто и наглядно.
2. **RFC 9457 Problem Details**: стандарт IETF для ошибок HTTP API, тип `application/problem+json`, поля `type`, `title`, `status`, `detail`, `instance` плюс свои расширения. Хорош для публичных API.

Обязательные свойства любого формата:

- **машиночитаемый код** (`validation_failed`, `not_found`, `conflict`), на который клиент пишет `switch`. Текст сообщения меняется, код нет;
- **поля с ошибками** для форм;
- **никаких внутренностей**: SQL, стектрейсы, пути к файлам и `err.Error()` от базы клиенту не уходят. Их место в логе;
- **статус HTTP соответствует смыслу**: 400 не разобрали, 401 не знаем, кто ты, 403 знаем, но нельзя, 404 нет ресурса, 409 конфликт (занят email), 422 не прошла валидация, 500 наша ошибка.

Удобно собрать ответы в пару функций и использовать их везде:

```go
type ErrorBody struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Fields  map[string]string `json:"fields,omitempty"`
}

func fail(c *gin.Context, status int, code, msg string, fields map[string]string) {
	c.AbortWithStatusJSON(status, gin.H{"error": ErrorBody{Code: code, Message: msg, Fields: fields}})
}

func bindError(c *gin.Context, err error) {
	if fields := FieldErrors(err); fields != nil {
		fail(c, http.StatusUnprocessableEntity, "validation_failed", "проверь поля запроса", fields)
		return
	}
	fail(c, http.StatusBadRequest, "bad_request", "не удалось разобрать запрос", nil)
}
```

::: trap null вместо пустого списка
Плохо:

```go
var page []Product // nil
for ... { page = append(page, p) }
c.JSON(200, gin.H{"data": page}) // {"data": null}, если ничего не нашлось
```

`encoding/json` кодирует nil-слайс как `null`, и фронтенд падает на `data.map is not a function`. То же с `Tags []string` в ответе, если клиент поля не прислал.

Хорошо: `page := []Product{}` или `make([]Product, 0, n)`. Пустой список в JSON всегда `[]`.
:::

## Query, URI и пагинация

Query-параметры привязываются к структуре тегами `form`, значение по умолчанию прямо в теге:

```go
type ListQuery struct {
	Page     int    `form:"page,default=1" binding:"min=1"`
	PerPage  int    `form:"per_page,default=20" binding:"min=1,max=100"`
	Category string `form:"category" binding:"omitempty,oneof=books games music"`
	Sort     string `form:"sort" binding:"omitempty,oneof=price -price name -name"`
}

type ProductURI struct {
	ID int64 `uri:"id" binding:"min=1"`
}
```

`page=abc` даст ошибку разбора (`strconv.ParseInt`), то есть 400; `page=0` и `per_page=500` дадут `ValidationErrors`, то есть 422. Для `ProductURI` не ставь `required`: у числа он означает «не ноль», и на `/products/0` клиент получит странное «обязательное поле» вместо «не меньше 1».

### Пагинация

Самая простая и самая распространённая: **offset/limit** (или page/per_page). Ответ несёт метаданные, чтобы клиент нарисовал навигацию:

```json
{
  "data": [ ... ],
  "meta": {"page": 2, "per_page": 20, "total": 57, "total_pages": 3}
}
```

Правила, которые проверяют на код-ревью:

- **верхняя граница `per_page`** обязательна. Без неё `?per_page=1000000` выгрузит всю таблицу и положит сервис;
- `total_pages = ceil(total / per_page)`, в целых числах `(total + per_page - 1) / per_page`;
- страница за концом списка это `200` и `"data": []`, а не 404 и не `null`;
- **порядок обязан быть детерминированным**. Если сортировать по цене, а у пяти товаров цена одинаковая, без второго ключа (обычно `id`) один и тот же товар может оказаться и на первой, и на второй странице. В SQL это `ORDER BY price, id`, в Go устойчивая сортировка `slices.SortStableFunc` по уже упорядоченному по `id` списку;
- поле сортировки только из белого списка (`oneof`). Подставлять `sort` из query в `ORDER BY` строкой это SQL-инъекция: плейсхолдеры `$1` работают для значений, но не для имён колонок.

У offset есть цена: `OFFSET 100000` заставляет базу прочитать и выбросить сто тысяч строк, а при вставке новых записей страницы «съезжают». Для лент и больших таблиц используют **keyset (cursor) пагинацию**: `WHERE (created_at, id) < ($1, $2) ORDER BY created_at DESC, id DESC LIMIT 20`, а клиенту отдают курсор на следующую страницу. Её подробно разбирают в модуле про базы данных; для админки и каталога на тысячи записей offset нормален.

Что `net/url` делает с точкой с запятой в query, посмотри в задаче в конце урока: это тоже защита, о которой мало кто знает.

## Что выведет программа?

```go
type Name struct {
	Name string `json:"name" binding:"required,min=2"`
}
var n Name
err := c.ShouldBindJSON(&n) // тело {"name":"Ян"}
fmt.Println(err == nil)
```

<Predict :options="['true', 'false: в Ян 4 байта, но min считает символы', 'false: min=2 значит больше двух', 'false: required не пропускает кириллицу']" :correct="0">

`min` и `max` для строк считают руны, а не байты, поэтому двухбуквенное имя проходит `min=2`. Если бы считались байты (как у `len`), то `"Ян"` прошло бы и `min=4`, а `max=3` его бы отвергло.

</Predict>

```go
// POST /orders, тело {"quantity":1,"gift":0,"admin":true}
// структура Order из раздела про required, без поля Admin
```

<Predict :options="['201, лишнее поле проигнорировано', '400: неизвестное поле admin', '422: admin не прошёл валидацию', '500: паника при разборе']" :correct="0">

По умолчанию `encoding/json` (и binding Gin) пропускает неизвестные поля. Строгий режим включается глобально через `binding.EnableDecoderDisallowUnknownFields = true`. Надёжная защита от подмены лишних полей: отдельная структура запроса, где этих полей просто нет.

</Predict>

```go
// GET /products?sort=price;DROP&page=2
var q struct {
	Sort string `form:"sort"`
	Page int    `form:"page"`
}
err := c.ShouldBindQuery(&q)
fmt.Printf("%q %d %v\n", q.Sort, q.Page, err)
```

<Predict :options="['sort = price;DROP, page = 2, без ошибки', 'sort пустой, page = 2, без ошибки', 'sort пустой, page = 0, ошибка разбора', 'sort = price, page = 2, без ошибки']" :correct="1">

Вывод: `"" 2 <nil>`. С Go 1.17 `net/url` больше не считает `;` разделителем параметров: пара `sort=price;DROP` признаётся некорректной и **молча выбрасывается**, а `url.ParseQuery` возвращает ошибку `invalid semicolon separator in query`, которую `URL.Query()` (и Gin) игнорирует. Остальные параметры разбираются. Если клиенту правда нужна точка с запятой в значении, он кодирует её как `%3B`, и тогда `sort` придёт как `price;DROP` и упрётся в `oneof`.

</Predict>

::: interview
<details><summary>Чем ShouldBindJSON отличается от BindJSON?</summary>

`ShouldBindJSON` только возвращает ошибку, ответ формирует обработчик. `BindJSON` через `MustBindWith` вызывает `c.AbortWithError(400, err)` (413, если сработал `MaxBytesReader`): прерывает цепочку и сразу пишет статус, после чего ответить 422 или в своём формате уже нельзя, статус не поменяется. В API почти всегда нужен `ShouldBind*`.

</details>
<details><summary>Как в Gin устроена валидация и почему тег называется binding?</summary>

Gin использует go-playground/validator/v10, но создаёт свой экземпляр и вызывает `SetTagName("binding")`. После привязки JSON, query или uri к структуре он вызывает `validate.Struct`. Доступ к валидатору для настройки: `binding.Validator.Engine().(*validator.Validate)`.

</details>
<details><summary>Почему required не пропускает 0 и false и как с этим жить?</summary>

`required` проверяет, что значение не нулевое для типа, а не что поле было в JSON. После разбора не переданное поле и переданный ноль неотличимы. Если ноль допустим, поле делают указателем (`*int`, `*bool`): тогда `required` проверяет `!= nil`, а ноль проходит.

</details>
<details><summary>Как отдать клиенту ошибки валидации по-русски с именами полей из JSON?</summary>

Зарегистрировать `RegisterTagNameFunc`, который берёт имя из тега `json`. Затем через `errors.As` достать `validator.ValidationErrors` и для каждой `FieldError` построить сообщение по `Tag()`, `Param()` и `Kind()`. Ключ брать из `Namespace()` без первого сегмента, чтобы вложенные поля и элементы слайсов были видны: `address.city`, `tags[1]`. Альтернатива: готовые переводы `validator/v10/translations/ru` с universal-translator.

</details>
<details><summary>Как написать своё правило валидации?</summary>

`v.RegisterValidation("slug", func(fl validator.FieldLevel) bool { ... })` на валидаторе из `binding.Validator.Engine()`, до запуска сервера. В функции `fl.Field()` это `reflect.Value` поля, `fl.Param()` параметр тега. После регистрации тег используется как встроенный: `binding:"required,slug"`. Для проверок, затрагивающих несколько полей, есть `RegisterStructValidation`.

</details>
<details><summary>400 или 422 для ошибок валидации?</summary>

400 Bad Request означает «запрос не разобрать»: битый JSON, строка вместо числа. 422 Unprocessable Content означает «запрос понятен, но данные нарушают правила». Разделение помогает клиенту, но многие API отвечают 400 на всё. Важнее единообразие во всём сервисе и машиночитаемый код ошибки в теле.

</details>
<details><summary>Какой формат ошибок API ты бы выбрал?</summary>

Единый для всех ручек: машиночитаемый `code`, человекочитаемый `message`, при валидации карта `fields`. Либо обёртка `{"error": {...}}`, либо стандарт RFC 9457 Problem Details (`application/problem+json`). Внутренние детали (SQL, стектрейсы) только в лог. Код ошибки стабилен, текст может меняться.

</details>
<details><summary>Что важно в пагинации offset/limit и когда переходить на keyset?</summary>

Ограничить `per_page` сверху, считать `total_pages` с округлением вверх, отдавать `[]` за концом списка, сортировать детерминированно с уникальным вторым ключом (`ORDER BY price, id`), поле сортировки брать из белого списка. Offset дорогой на глубоких страницах (база читает и выбрасывает строки) и «съезжает» при вставках, поэтому для лент и больших таблиц берут keyset: `WHERE (created_at, id) < (курсор) ORDER BY ... LIMIT n`.

</details>
:::

## Источники

- `RU` [Привязка модели и валидация](https://gin-gonic.com/ru/docs/binding/binding-and-validation/): официальная документация Gin, `Bind` против `ShouldBind`, теги.
- `RU` [Пользовательские валидаторы](https://gin-gonic.com/ru/docs/binding/custom-validators/) и [Привязка URI](https://gin-gonic.com/ru/docs/binding/bind-uri/): короткие официальные примеры.
- `RU` [Валидация данных в Go с go-playground/validator](https://habr.com/ru/articles/927460/): обзор правил, кастомных валидаторов и разбора ошибок на Хабре.
- `EN` [validator/v10 на pkg.go.dev](https://pkg.go.dev/github.com/go-playground/validator/v10): полный список тегов (раздел Baked In Validators), `dive`, `RegisterTagNameFunc`.
- `EN` [RFC 9457: Problem Details for HTTP APIs](https://www.rfc-editor.org/rfc/rfc9457): стандартный формат ошибок HTTP API.
- `EN` [Пакет gin на pkg.go.dev](https://pkg.go.dev/github.com/gin-gonic/gin): методы `Context.ShouldBind*` и `Bind*` с описанием поведения при ошибке.

## Домашка

<Hw id="ex1" title="Ошибки валидации по-русски" level="база">

Файл `homework/m08/l02-binding/ex1_errors.go`: `SetupValidator` (имена полей из `json`, тег `slug`) и `FieldErrors` (карта «поле: сообщение» по таблице в комментарии). Тест прогоняет структуру со всеми популярными тегами, вложенной структурой и `dive`.

```bash
cd homework && go test -v ./m08/l02-binding/ -run 'TestFieldErrors|TestSlug'
```

</Hw>

<Hw id="ex2" title="Создание товара" level="база">

Файл `ex2_create.go`: допиши теги в `CreateProductRequest` и обработчик `CreateProduct` с единым форматом ответов из `store.go`. Тест различает 400 (битый JSON), 422 (валидация, с полями), 409 (занятый slug) и 201, проверяет `[]` вместо `null`. `BindJSON` вместо `ShouldBindJSON` тест поймает по статусу.

```bash
cd homework && go test -v ./m08/l02-binding/ -run TestCreateProduct
```

</Hw>

<Hw id="ex3" title="Список с фильтром, сортировкой и страницами" level="посложнее">

Файл `ex3_list.go`: теги `ListQuery` и `ProductURI`, обработчики `ListProducts` и `GetProduct`. Тест проверяет границы страниц, `total_pages`, пустую страницу, устойчивость сортировки на 40 товарах с одинаковыми ценами и отказ на `sort` не из белого списка.

```bash
cd homework && go test -v ./m08/l02-binding/ -run 'TestListProducts|TestGetProduct'
```

</Hw>

<Hw id="ex4" title="Problem Details" level="со звёздочкой">

Вне домашнего пакета перепиши функцию `fail` из урока так, чтобы ошибки уходили в формате RFC 9457: `Content-Type: application/problem+json`, поля `type`, `title`, `status`, `detail` и расширение `errors` с полями валидации. Подумай и запиши: что станет с клиентом, который уже разбирает `{"error": {"code": ...}}`, и как бы ты провёл такую смену формата без поломки (версия API, заголовок `Accept`, период двойной поддержки).

</Hw>
