# http.Client без сюрпризов

Бэкенд почти никогда не живёт один: он ходит в платёжку, в сервис доставки, в соседний микросервис, в S3. Каждый такой вызов это `http.Client`, и почти все инциденты «сервис лёг, хотя мы ничего не меняли» начинаются здесь. Внешний API стал отвечать за минуту вместо 100 мс, а у клиента нет таймаута: горутины копятся, память растёт, сервис умирает. Или забыли закрыть тело ответа: каждое соединение утекает, через час кончаются файловые дескрипторы. Или ретраи без паузы добили и без того перегруженный сервис.

В браузере `fetch` многое решает за тебя: пул соединений, лимиты, отмена при уходе со страницы. В Go всё это есть, но включается осознанно. В этом уроке разбираем, как устроен клиент (Client и Transport), какие бывают таймауты, почему тело ответа надо закрывать, как отменять запросы контекстом, как правильно повторять и как тестировать клиентский код через `httptest`.

Вопросы на собесе: «что не так с `http.Get` в продакшене», «зачем закрывать `resp.Body`», «какие таймауты бывают у HTTP-клиента», «какие запросы можно ретраить».

## Client и Transport

```go
resp, err := http.Get("https://api.example.com/items")
```

`http.Get` это `http.DefaultClient.Get`. А `DefaultClient` это `&http.Client{}` без единой настройки. Клиент устроен в два слоя:

- **`http.Client`**: политика уровня запроса. Общий таймаут, редиректы, куки.
- **`http.Transport`** (реализует интерфейс `http.RoundTripper`): механика. Пул соединений, дозвон, TLS, прокси, HTTP/2. Транспорт по умолчанию `http.DefaultTransport`.

```go
type RoundTripper interface {
	RoundTrip(*http.Request) (*http.Response, error)
}
```

`RoundTripper` это клиентский двойник `http.Handler`: один метод, «отправь запрос, получи ответ». Его оборачивают так же, как хендлеры middleware-ами: для логирования, метрик, подстановки токена, а в тестах подменяют фейком.

**Правило: один клиент (точнее, один транспорт) на всё приложение или на каждый внешний сервис, переиспользуемый всеми горутинами.** `Client` и `Transport` потокобезопасны. Пул соединений живёт в транспорте, и новый транспорт на каждый запрос означает новое TCP- и TLS-рукопожатие на каждый запрос, а соединения брошенных пулов остаются открытыми: у `&http.Transport{}` `IdleConnTimeout` равен нулю, и закроет их только сервер.

::: trap Клиент на каждый запрос
```go
// плохо
func fetch(url string) (*http.Response, error) {
	c := &http.Client{Transport: &http.Transport{}, Timeout: 5 * time.Second}
	return c.Get(url)
}
```

Каждый вызов создаёт новый пул, соединение из которого больше никто не использует. Под нагрузкой это тысячи открытых соединений (утекают файловые дескрипторы у тебя, копятся соединения на сервере) и лишние 2 RTT на каждый запрос.

```go
// хорошо: создать один раз, передавать как зависимость
type PaymentsClient struct {
	http *http.Client
	base string
}

func NewPaymentsClient(base string) *PaymentsClient {
	return &PaymentsClient{http: &http.Client{Timeout: 5 * time.Second}, base: base}
}
```
:::

## Таймауты: что ограничивает что

```go
fmt.Println(http.DefaultClient.Timeout, http.DefaultMaxIdleConnsPerHost)
```

<Predict :options="['30s 100', '0s 2', '0s 100', '1m30s 2']" :correct="1">

У `DefaultClient` таймаута нет вовсе (`0` значит «бесконечно»): если сервер принял соединение и молчит, `http.Get` будет ждать вечно, а с ним и горутина хендлера, который его вызвал. А `MaxIdleConnsPerHost` по умолчанию `2`: в пуле держится не больше двух простаивающих соединений к одному хосту. При 50 параллельных запросах к одному API остальные 48 соединений после ответа закрываются и потом открываются заново.

</Predict>

Таймауты бывают на разных стадиях запроса:

```text
 дозвон (TCP)    TLS        отправка   ждём заголовки   читаем тело
|--Dialer.Timeout--|--TLSHandshakeTimeout--|   |--ResponseHeaderTimeout--|
|<----------------------- Client.Timeout ------------------------------------------>|
|<--------------------- context deadline (на конкретный запрос) ------------------->|
```

- **`Client.Timeout`**: всё целиком, от дозвона до **дочитывания тела**. Самый простой и самый важный. Ставь всегда.
- **`net.Dialer.Timeout`** (в `DefaultTransport` 30 секунд): установка TCP-соединения.
- **`Transport.TLSHandshakeTimeout`** (10 секунд по умолчанию).
- **`Transport.ResponseHeaderTimeout`**: сколько ждать заголовков ответа после отправки запроса. По умолчанию нет. Полезен, когда тело может стримиться долго, но ответ должен начаться быстро.
- **`Transport.IdleConnTimeout`** (90 секунд): сколько держать простаивающее соединение в пуле.
- **Дедлайн контекста**: таймаут для конкретного вызова. Если `Client.Timeout` 10 секунд, а этот запрос нужен за 300 мс, `context.WithTimeout` на запрос.

Удобный способ получить хороший транспорт: склонировать дефолтный и поправить нужное.

```go
tr := http.DefaultTransport.(*http.Transport).Clone()
tr.MaxIdleConnsPerHost = 32
tr.ResponseHeaderTimeout = 5 * time.Second
client := &http.Client{Timeout: 10 * time.Second, Transport: tr}
```

`Clone` сохраняет прокси из окружения (`HTTP_PROXY`), таймауты дозвона и TLS и `ForceAttemptHTTP2: true`. Транспорт с нуля `&http.Transport{}` не знает про прокси из окружения и не имеет таймаутов дозвона и TLS. HTTP/2 он включает сам, но только пока ты не задал свой `DialContext` или `TLSClientConfig`: тогда автоматический HTTP/2 пропадает, если не выставить `ForceAttemptHTTP2: true`, как у `DefaultTransport`.

::: trap Правка DefaultTransport
```go
// плохо: глобальная настройка для всех
http.DefaultTransport.(*http.Transport).MaxIdleConnsPerHost = 100
http.DefaultClient.Timeout = 2 * time.Second
```

`DefaultTransport` и `DefaultClient` общие для всей программы, включая библиотеки (SDK облаков, OpenTelemetry-экспортёры), которые ты не писал. Ты поменял поведение чужого кода, а гонка данных, если правишь во время работы, бонусом. Делай свою копию через `Clone()` и свой `http.Client`.
:::

## Тело ответа надо закрыть

```go
resp, err := client.Do(req)
if err != nil {
	return err
}
defer resp.Body.Close()
```

Почему это важно: `resp.Body` держит **соединение**. Пока тело не прочитано до конца или не закрыто, транспорт не может вернуть соединение в пул и отдать следующему запросу. Не закрыл: соединение висит, а следующий запрос открывает новое. Через какое-то время упираешься в лимит файловых дескрипторов (`too many open files`) или в лимит соединений на стороне сервера.

Правила:

- `defer resp.Body.Close()` **сразу после проверки ошибки**, до любых других проверок и ранних `return`.
- Когда `err != nil`, `resp` обычно `nil`, закрывать нечего. Исключение описано в документации `Client.Do`: ответ вместе с ошибкой бывает, только когда упал `CheckRedirect`, и тело тогда уже закрыто.
- Тело надо закрывать **всегда**, даже если статус не 2xx и тело тебе не нужно.
- Соединение возвращается в пул, когда тело прочитано до `EOF` или закрыто. Если закрыть, не дочитав, раньше соединение просто закрывалось. С **Go 1.27** `Close` у HTTP/1-ответа сам дочитывает небольшой непрочитанный остаток (до консервативного лимита), чтобы соединение можно было переиспользовать. Большой непрочитанный ответ всё равно приведёт к закрытию соединения, поэтому, если тело может быть большим, а оно не нужно, дочитай с ограничением: `io.Copy(io.Discard, io.LimitReader(resp.Body, 64<<10))`.

```go
srv := httptest.NewServer(http.NotFoundHandler())
defer srv.Close()

resp, err := http.Get(srv.URL)
if err != nil {
	fmt.Println("ошибка:", err)
	return
}
defer resp.Body.Close()
fmt.Println("ok", resp.StatusCode)
```

<Predict :options="['ошибка: 404 Not Found', 'ok 404', 'ok 200', 'паника']" :correct="1">

Для `http.Client` любой полученный ответ это успех, включая 404 и 500. `err` бывает только при сетевых проблемах, таймаутах, отмене, ошибках редиректа. Проверять статус обязан ты: `if resp.StatusCode != http.StatusOK`. Многие с фронтенда ждут поведения axios, который бросает исключение на 4xx/5xx; `fetch`, кстати, тоже не бросает.

</Predict>

Типичная функция похода во внешний API:

```go
func (c *Client) GetItem(ctx context.Context, id string) (Item, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.base+"/items/"+url.PathEscape(id), nil)
	if err != nil {
		return Item{}, err
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return Item{}, fmt.Errorf("get item %s: %w", id, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		snippet, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return Item{}, fmt.Errorf("get item %s: status %d: %s", id, resp.StatusCode, snippet)
	}
	var it Item
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&it); err != nil {
		return Item{}, fmt.Errorf("get item %s: decode: %w", id, err)
	}
	return it, nil
}
```

Обрати внимание: `url.PathEscape` для подстановки в путь (иначе `id = "../admin"` уйдёт куда не надо), лимит на чтение тела ошибки и тела ответа (чужой сервис тоже может прислать гигабайт), контекст в запросе.

::: hood Как устроен пул соединений
`Transport` хранит простаивающие соединения в мапе по ключу «схема + хост + порт (+ прокси)». Запрос сначала ищет свободное соединение в пуле; если его нет, начинает дозвон, а параллельно ждёт, не освободится ли существующее, и берёт то, что случится раньше. Ответ читается отдельной горутиной соединения (`readLoop`), а тело, которое ты получаешь, это обёртка, которая при `EOF` или `Close` сигнализирует этой горутине: «соединение свободно». Тогда оно возвращается в пул, если там меньше `MaxIdleConnsPerHost` соединений к этому хосту и меньше `MaxIdleConns` всего; иначе закрывается.

Для HTTP/2 всё иначе: к хосту обычно одно соединение, по которому идут параллельные потоки, и `MaxIdleConnsPerHost` влияет мало. Но закрывать тело всё равно нужно: незакрытый поток держит ресурсы и окно управления потоком.
:::

## Контекст и отмена

Запрос, созданный через `http.NewRequestWithContext(ctx, ...)`, отменяется, когда отменяется `ctx`: транспорт прерывает дозвон, отправку или чтение, а `Do` (или чтение `Body`) вернёт ошибку, для которой `errors.Is(err, context.Canceled)` или `context.DeadlineExceeded` истинно.

Внутри HTTP-хендлера передавай дальше `r.Context()`: клиент твоего сервиса ушёл, значит и походы во внешние сервисы ради него можно прекратить.

```go
func handler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 800*time.Millisecond)
	defer cancel()
	item, err := payments.GetItem(ctx, r.PathValue("id"))
	// ...
}
```

Параллельные запросы с отменой (например, «спроси три зеркала, возьми первый ответ») требуют аккуратности: проигравшие запросы надо отменить, их горутины дождаться, а тела закрыть. Если вернуться сразу после первого ответа и не дождаться остальных, горутины проигравших продолжат жить, а с небуферизованным каналом результатов зависнут на отправке навсегда. Это задание в домашке.

## Редиректы

`Client` сам идёт по редиректам, до 10 подряд (потом ошибка `stopped after 10 redirects`). При этом:

- на `301`, `302`, `303` метод `POST` превращается в `GET`, тело выбрасывается (так исторически делают браузеры);
- на `307` и `308` метод и тело сохраняются, если тело можно перечитать (`req.GetBody` есть; `http.NewRequest` выставляет его для `strings.Reader`, `bytes.Reader`, `bytes.Buffer`).

```go
mux.HandleFunc("/old", func(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/new", http.StatusFound) // 302
})
mux.HandleFunc("/new", func(w http.ResponseWriter, r *http.Request) {
	b, _ := io.ReadAll(r.Body)
	fmt.Fprintf(w, "%s %q", r.Method, b)
})
// ...
resp, _ := http.Post(srv.URL+"/old", "text/plain", strings.NewReader("data"))
b, _ := io.ReadAll(resp.Body)
resp.Body.Close()
fmt.Println(resp.StatusCode, string(b))
```

<Predict :options="['302 (пусто)', '200 POST \'data\'', '200 GET \'\'', 'ошибка: redirect']" :correct="2">

Клиент сам перешёл по 302 и вернул финальный ответ 200. По правилам для 301/302/303 POST стал GET без тела. С `http.StatusTemporaryRedirect` (307) был бы `POST "data"`. Если редиректы не нужны (например, ты сам хочешь увидеть `Location`, как в сокращателе ссылок), задай `CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }`.

</Predict>

## Ретраи

Сеть ненадёжна, и короткие сбои (обрыв соединения, 503 при деплое соседа) лечатся повтором. Но ретраи это и способ добить перегруженный сервис. Правила:

1. **Повторяй только то, что безопасно повторять.** Идемпотентные методы (`GET`, `HEAD`, `OPTIONS`, `PUT`, `DELETE`) или запросы с `Idempotency-Key`, который сервер умеет обрабатывать. `POST` «создать платёж» без ключа повторять нельзя: таймаут не значит, что платёж не прошёл.
2. **Повторяй только временные ошибки.** Сетевые ошибки, `429`, `502`, `503`, `504`. Не повторяй `4xx` (запрос неверный, повтор ничего не изменит) и обычно не `500` (баг на той стороне, скорее всего, повторится).
3. **Не повторяй отменённое.** Если истёк контекст запроса, повтор бессмыслен: вызывающему уже не нужен результат.
4. **Пауза растёт экспоненциально**: 100 мс, 200, 400, 800... с потолком. И **с джиттером** (случайной добавкой): иначе тысяча клиентов, упавших одновременно, будут повторять синхронно и снова положат сервис. Популярный вариант full jitter: пауза случайная от 0 до `min(cap, base * 2^n)`.
5. **Уважай `Retry-After`**: сервер сам говорит, когда приходить.
6. **Ограничивай число попыток** и общий бюджет времени (контекст). Каждая попытка закрывает тело предыдущего ответа.
7. **Тело запроса для повтора** должно перечитываться. Главное правило: нет `req.GetBody`, значит тело не повторить, и запрос не ретраят. Если `GetBody` есть, транспорт при повторной отправке того же `*Request` сам перемотает тело через него. Явно звать `req.GetBody()` нужно, когда ты собираешь новый запрос (`req.Clone`) или отправляешь тело своим кодом.

```go
for attempt := 1; ; attempt++ {
	resp, err := client.Do(req)
	if err == nil && !temporary(resp.StatusCode) || attempt == maxAttempts {
		return resp, err
	}
	if resp != nil {
		resp.Body.Close() // промежуточный ответ не нужен
	}
	delay := min(base<<(attempt-1), maxDelay)
	select {
	case <-time.After(rand.N(delay)): // full jitter
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}
```

Это эскиз: в нём нет проверки идемпотентности и перечитывания тела. Полную версию пишешь в домашке, а паузу там подменяет функция `Sleep`, чтобы тест не ждал по-настоящему.

::: trap Ретраи на каждом уровне
```text
// плохо: фронт делает 3 попытки, API-шлюз 3, сервис A 3 попытки в сервис B
// одна ошибка B превращается в 27 запросов к B
```

Ретраи перемножаются по цепочке вызовов. Под нагрузкой это «шторм ретраев», который не даёт упавшему сервису подняться. Хорошо: повторять на одном уровне (обычно ближайшем к источнику проблемы), ограничивать долю повторов (retry budget, например не больше 10% от успешных), а при длительных отказах включать circuit breaker: перестать ходить в сервис на время и сразу отдавать ошибку.
:::

## Тестирование клиента

Ходить в настоящий внешний API из тестов нельзя: медленно, нестабильно, платно. Два инструмента из стандартной библиотеки.

**`httptest.NewServer`** поднимает настоящий HTTP-сервер на `127.0.0.1` со случайным портом. Ты пишешь хендлер, который изображает внешний API, и направляешь клиента на `srv.URL`:

```go
func TestGetItem(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/items/42" {
			t.Errorf("путь %q, want /items/42", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"id":"42","name":"чай"}`)
	}))
	defer srv.Close()

	c := &Client{http: srv.Client(), base: srv.URL}
	got, err := c.GetItem(context.Background(), "42")
	// проверки
}
```

Так проверяются и пограничные случаи, которые с реальным API не воспроизвести: сервер отвечает 503 два раза подряд, зависает, рвёт соединение (`http.NewResponseController(w).Hijack()` и `Close`), шлёт гигантское тело. `srv.Client()` отдаёт клиент, настроенный на этот сервер (для `NewTLSServer` он доверяет его сертификату).

**Свой `RoundTripper`** подменяет транспорт без сети вообще или оборачивает настоящий, чтобы подсмотреть: сколько было попыток, закрыты ли тела.

```go
type countingTransport struct {
	base  http.RoundTripper
	calls atomic.Int64
}

func (t *countingTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	t.calls.Add(1)
	return t.base.RoundTrip(r)
}
```

С Go 1.27 есть ещё `httptest.NewTestServer`: сервер на фейковой сети в памяти для использования вместе с `testing/synctest`, где время тоже фейковое. Тест «запрос с таймаутом в 10 секунд» выполняется мгновенно и детерминированно. Это пригодится, когда тестов с таймаутами станет много.

## Что выведет программа?

```go
ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
defer cancel()
req, _ := http.NewRequestWithContext(ctx, "GET", slowServerURL, nil) // сервер отвечает через 1 с
_, err := http.DefaultClient.Do(req)
fmt.Println(errors.Is(err, context.DeadlineExceeded))
```

<Predict :options="['false: у DefaultClient нет таймаута', 'true', 'программа ждёт секунду и печатает false', 'паника: nil response']" :correct="1">

У `DefaultClient` нет своего таймаута, но дедлайн контекста запроса работает независимо от него. Через 50 мс транспорт прерывает ожидание, `Do` возвращает ошибку, и `errors.Is` находит в цепочке `context.DeadlineExceeded`. Поэтому контекст в запросе нужен всегда, даже если у клиента есть `Timeout`.

</Predict>

::: interview
<details><summary>Что не так с http.Get в продакшен-коде?</summary>

Он использует `http.DefaultClient` без таймаута: зависший сервер повесит вызов навсегда. Нет контекста, значит нельзя отменить по уходу клиента или по дедлайну. Транспорт по умолчанию держит лишь 2 простаивающих соединения на хост. Правильно: свой `http.Client` с `Timeout`, транспорт через `DefaultTransport.Clone()` с нужными лимитами, запросы через `NewRequestWithContext`.

</details>
<details><summary>Зачем закрывать resp.Body? Что будет, если не закрыть?</summary>

Тело держит соединение. Пока оно не прочитано до конца или не закрыто, транспорт не может вернуть соединение в пул. Незакрытые тела копятся: на каждый запрос новое соединение, утечка горутин чтения и файловых дескрипторов, в итоге `too many open files` или исчерпание лимита соединений. `defer resp.Body.Close()` сразу после проверки `err`. С Go 1.27 `Close` сам дочитывает небольшой остаток для переиспользования соединения, но закрывать всё равно обязательно.

</details>
<details><summary>Какие таймауты есть у HTTP-клиента?</summary>

`Client.Timeout`: весь запрос, включая чтение тела. В транспорте: `Dialer.Timeout` (TCP), `TLSHandshakeTimeout`, `ResponseHeaderTimeout` (ожидание заголовков), `ExpectContinueTimeout`, `IdleConnTimeout` (простой в пуле). Плюс дедлайн контекста на конкретный запрос. Минимум: `Client.Timeout` и контекст.

</details>
<details><summary>Возвращает ли http.Client ошибку на 500?</summary>

Нет. Любой полученный HTTP-ответ, включая 4xx и 5xx, это `err == nil`. Ошибка бывает при сетевых сбоях, таймаутах, отмене, проблемах TLS, превышении числа редиректов. Статус проверяют руками.

</details>
<details><summary>Какие запросы можно автоматически повторять и как?</summary>

Идемпотентные (GET, HEAD, OPTIONS, PUT, DELETE) или с `Idempotency-Key`, на временных ошибках: сетевые сбои, 429, 502, 503, 504. Экспоненциальная пауза с джиттером и потолком, уважение `Retry-After`, ограничение попыток и общего времени через контекст, закрытие тела промежуточных ответов, перечитывание тела запроса через `GetBody`. Не повторять 4xx и отменённые запросы. Учитывать перемножение ретраев по цепочке сервисов.

</details>
<details><summary>Зачем джиттер в backoff?</summary>

Без него клиенты, получившие ошибку одновременно (например, при перезапуске сервиса), повторяют синхронно: волны нагрузки в моменты 100 мс, 300 мс, 700 мс снова кладут сервис. Случайная пауза размазывает повторы по времени. Подробный разбор вариантов (full, equal, decorrelated jitter) в статье AWS.

</details>
<details><summary>Как протестировать код, который ходит во внешний HTTP API?</summary>

`httptest.NewServer` с хендлером, изображающим API (в том числе ошибки, медленные ответы, обрывы), и клиент, направленный на `srv.URL`. Или свой `RoundTripper`-фейк в `Client.Transport` без сети. Обёртка-`RoundTripper` поверх настоящего транспорта помогает считать попытки и проверять, что тела закрыты. Для тестов с временем в Go 1.27 есть `testing/synctest` и `httptest.NewTestServer` на фейковой сети.

</details>
<details><summary>Что будет с методом POST при редиректе?</summary>

На 301, 302 и 303 `http.Client`, как и браузеры, меняет POST на GET и выбрасывает тело. На 307 и 308 сохраняет метод и тело, если его можно перечитать (`GetBody`). Не больше 10 редиректов; поведение настраивается `CheckRedirect`, а `http.ErrUseLastResponse` останавливает переходы и возвращает сам ответ с редиректом.

</details>
:::

## Источники

- `RU` [Настройка инструментов стандартной библиотеки Golang для production, Хабр (iSpring)](https://habr.com/ru/companies/ispring/articles/560032/): таймауты клиента и сервера, почему `http.Get` не подходит.
- `RU` [Чиним проблемы нагрузок в Go с помощью настройки пула HTTP-соединений, Хабр (Ozon)](https://habr.com/ru/companies/ozontech/articles/561790/): `MaxIdleConnsPerHost` на практике, как его нехватка выглядит под нагрузкой.
- `RU` [50 оттенков Go, Хабр (VK)](https://habr.com/ru/companies/vk/articles/314804/): классический список ловушек; найди раздел про закрытие тела HTTP-ответа (статья 2016 года: про автоматическое дочитывание в Go 1.27 там, конечно, нет).
- `EN` [The complete guide to Go net/http timeouts](https://blog.cloudflare.com/the-complete-guide-to-golang-net-http-timeouts/): Cloudflare со схемами, какой таймаут на какой стадии срабатывает.
- `EN` [Exponential Backoff And Jitter](https://aws.amazon.com/blogs/architecture/exponential-backoff-and-jitter/): AWS сравнивает стратегии джиттера на симуляции.
- `EN` [Документация net/http](https://pkg.go.dev/net/http): разделы `Client`, `Transport` и `Response.Body` (кто и когда закрывает).

## Домашка

<Hw id="ex1" title="FetchJSON" level="база">

Файл `homework/m06/l04-http-client/ex1_fetch.go`: GET с контекстом, декодирование JSON, `*StatusError` с началом тела для не-2xx и закрытие тела на всех путях. Тест считает TCP-соединения и закрытые тела.

```bash
cd homework && go test -v ./m06/l04-http-client/ -run TestFetchJSON
```

</Hw>

<Hw id="ex2" title="Настроенный клиент" level="база">

Файл `ex2_client.go`, функция `NewClient`: таймаут, копия `DefaultTransport` с увеличенным пулом и `ResponseHeaderTimeout`, сам `DefaultTransport` не трогать.

```bash
cd homework && go test -v ./m06/l04-http-client/ -run TestNewClient
```

</Hw>

<Hw id="ex3" title="Ретраи с backoff" level="посложнее">

Файл `ex3_retry.go`, функция `Do`: повтор только идемпотентного, только временных ошибок, экспонента с потолком, `Retry-After`, закрытие промежуточных тел, свежее тело через `GetBody`, отмена контекстом.

```bash
cd homework && go test -v ./m06/l04-http-client/ -run TestDo
```

</Hw>

<Hw id="ex4" title="Первый ответ из нескольких" level="со звёздочкой">

Файл `ex4_first.go`, функция `First`: параллельные запросы к зеркалам, первый 200 побеждает, остальные отменяются, функция возвращается только после завершения всех горутин. Тест проверяет, что медленные серверы увидели отмену и что после возврата не осталось запросов в полёте.

```bash
cd homework && go test -race -v ./m06/l04-http-client/ -run TestFirst
```

</Hw>

<Hw id="ex5" title="Утечка соединений вживую" level="посложнее">

Напиши программу, которая 200 раз подряд делает `GET` к `httptest.NewServer` и **не закрывает** тело. На сервере повесь `srv.Config.ConnState` и посчитай `http.StateNew`. Сравни с вариантом, где тело закрыто, и с вариантом, где оно закрыто без чтения, а сервер отдаёт 1 МБ. Объясни разницу, опираясь на раздел про пул и на изменение в Go 1.27.

</Hw>
