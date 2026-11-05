# Аутентификация и безопасность API

До этого урока твой API верил всем: кто угодно мог создать задачу, прочитать чужую и удалить всё подряд. В реальном сервисе первым делом появляются вопросы «кто ты?» и «что тебе можно?». На них отвечают пароли, токены, middleware и проверки владельца. И именно здесь чаще всего ломаются настоящие продукты: в рейтинге рисков API от OWASP первое место занимает не хитрая криптография, а банальное «поменял id в URL и прочитал чужое».

Ты пришёл из фронтенда, поэтому половина темы тебе знакома с другой стороны: ты уже клал токен в `localStorage`, ловил 401 в интерсепторе axios и делал «тихий» refresh. Теперь посмотрим, что происходит на сервере, почему одни решения безопасны, а другие только выглядят такими, и что из этого спрашивают на собесе.

План: храним пароли правильно (bcrypt, argon2id), честно сравниваем сессии и JWT, разбираем JWT по байтам и его атаки, строим пару access и refresh с ротацией, пишем middleware аутентификации и авторизации на Gin, проходим OWASP API Top 10 и заканчиваем rate limiting и сравнением за постоянное время.

## Аутентификация и авторизация

Два слова похожи, но отвечают на разные вопросы:

- **Аутентификация** (authentication, authn): кто ты? Проверяем пароль, токен, сертификат. Провал: **401 Unauthorized**. Название кода историческое и сбивает с толку: по смыслу это «не аутентифицирован».
- **Авторизация** (authorization, authz): что тебе можно? Пользователь известен, но у него нет роли или это не его ресурс. Провал: **403 Forbidden**.

Правило для ответа на 401 из RFC 6750: сервер добавляет заголовок `WWW-Authenticate: Bearer`, так клиент понимает, какой способ входа ожидается. На 403 повторный вход не поможет, и фронтенд не должен в этом случае бросать пользователя на страницу логина.

Есть третий случай, важный для безопасности: пользователь просит **чужой** объект. Ответить 403 значит подтвердить, что объект с таким id существует. Поэтому для чужих ресурсов обычно отвечают **404 Not Found**, как будто объекта нет вовсе. GitHub делает именно так с приватными репозиториями.

## Хранение паролей

### Почему не SHA-256

Первая мысль: «захеширую пароль SHA-256, и всё». Проблема в том, что SHA-256 создан быть **быстрым**. Современная видеокарта считает миллиарды SHA-256 в секунду, поэтому утёкшую таблицу `users` с такими хешами перебирают по словарям за часы. Соль (случайная строка, добавленная к паролю перед хешированием) спасает от заранее посчитанных радужных таблиц и от одинаковых хешей у одинаковых паролей, но не от перебора: перебирать каждый хеш отдельно на GPU всё ещё дёшево.

Для паролей нужен **медленный** хеш с настраиваемой стоимостью: bcrypt, scrypt или argon2id. Каждая проверка пароля стоит десятки миллисекунд. Для одного входа это незаметно, а для перебора миллиардов вариантов неподъёмно.

### bcrypt

bcrypt живёт в `golang.org/x/crypto/bcrypt` и делает всё сам: генерирует соль, кладёт её в результат, сравнивает за постоянное время.

```go
import "golang.org/x/crypto/bcrypt"

hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost) // cost 10
if err != nil {
	return err // например, bcrypt.ErrPasswordTooLong
}
// в базу кладём string(hash)

// при входе:
if err := bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(password)); err != nil {
	return ErrUnauthorized
}
```

**Cost** это логарифм числа раундов: +1 к cost удваивает время. На M-серии Mac cost 10 занимает около 60 мс, cost 12 около 200 мс (замерено на машине автора, у тебя будет иначе). OWASP рекомендует минимум 10 и «столько, сколько выдержит сервер». Не путай `bcrypt.MinCost` (4) с минимумом для продакшена: это просто нижняя граница алгоритма, годится только для тестов.

```go
package main

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	h1, _ := bcrypt.GenerateFromPassword([]byte("qwerty"), 10)
	h2, _ := bcrypt.GenerateFromPassword([]byte("qwerty"), 10)
	fmt.Println(string(h1) == string(h2))
	fmt.Println(bcrypt.CompareHashAndPassword(h1, []byte("qwerty")))
	fmt.Println(bcrypt.CompareHashAndPassword(h2, []byte("qwerty")))
}
```

<Predict :options="['true, <nil>, <nil>', 'false, <nil>, <nil>', 'false, <nil>, ошибка mismatched', 'true, <nil>, ошибка mismatched']" :correct="1">

Хеши разные, потому что у каждого своя случайная соль. Но проверка проходит для обоих: `CompareHashAndPassword` достаёт соль и cost из самого хеша, хеширует пароль заново с ними и сравнивает. Поэтому сравнивать хеши через `==` бессмысленно, а хранить соль в отдельной колонке не нужно.

</Predict>

::: hood Что лежит в bcrypt-хеше
Хеш выглядит так: `$2a$10$N9qo8uLOickgx2ZMRZoMye...`, всего 60 символов.

```text
$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy
 |  |  |<------- соль ------->|<------- хеш (184 бита) ------>|
 |  |  22 символа (128 бит)   31 символ
 |  cost = 10, значит 2^10 раундов расширения ключа
 версия алгоритма
```

Кодировка base64 своя, bcrypt-овская (алфавит `./A-Za-z0-9`). Всё, что нужно для проверки, хранится в одной строке, поэтому можно постепенно поднимать cost: при успешном входе посмотри `bcrypt.Cost(hash)`, и если он меньше текущего, перехешируй пароль (он как раз у тебя в руках) и обнови запись.
:::

### Ограничение в 72 байта

bcrypt использует только первые 72 **байта** пароля. Раньше многие реализации молча обрезали хвост: пароли `aaa...a1` и `aaa...a2` длиной 80 байт считались одинаковыми. В `x/crypto/bcrypt` теперь есть `ErrPasswordTooLong`: пароль длиннее 72 байт даёт ошибку, а не тихую обрезку.

Для кириллицы это особенно коварно: буква в UTF-8 занимает 2 байта, поэтому лимит наступает уже на 37 символах. Проверяй `len(password)`, это байты, а не `utf8.RuneCountInString`.

### argon2id

Argon2 победил в Password Hashing Competition в 2015 году, а его вариант argon2id (закреплён в RFC 9106) сейчас первая рекомендация OWASP. Главное отличие от bcrypt: он **memory-hard**, то есть требует много памяти на каждую проверку. Видеокарты и ASIC сильны в вычислениях, но память на ядро у них дорогая, поэтому параллельный перебор становится намного дороже.

В Go он есть в `golang.org/x/crypto/argon2`, но это низкоуровневая функция: соль, параметры и формат хранения на тебе.

```go
salt := make([]byte, 16)
rand.Read(salt) // crypto/rand
key := argon2.IDKey([]byte(password), salt, 2, 19*1024, 1, 32) // time=2, memory=19 МиБ, threads=1, длина 32 байта

// храним параметры вместе с солью и ключом, как это делает bcrypt:
encoded := fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
	argon2.Version, 19*1024, 2, 1,
	base64.RawStdEncoding.EncodeToString(salt),
	base64.RawStdEncoding.EncodeToString(key))
// $argon2id$v=19$m=19456,t=2,p=1$...
```

При проверке разбираешь строку, считаешь `IDKey` с теми же параметрами и солью и сравниваешь через `subtle.ConstantTimeCompare`. Параметры `m=19456 (19 МиБ), t=2, p=1` это минимальная конфигурация из OWASP Password Storage Cheat Sheet.

| | bcrypt | argon2id |
|---|---|---|
| Стойкость к GPU | хорошая | лучше (нужна память) |
| API в Go | одна функция, формат готов | низкоуровневый, формат пишешь сам |
| Ограничения | 72 байта | нет |
| Риск | минимальный | ошибиться в формате и сравнении |

Честный вывод для джуна: bcrypt с cost 10-12 остаётся нормальным выбором и проще в правильном использовании. argon2id лучше по стойкости, но требует аккуратности и памяти на сервере: 19 МиБ на каждый одновременный вход это 1.9 ГБ при сотне параллельных логинов.

::: trap Пароль сравнивают через == или хранят «как есть»
```go
// плохо: пароль в базе открытым текстом, сравнение обычным ==
if user.Password == req.Password { ... }

// плохо: хеш без соли и быстрый
if user.Hash == fmt.Sprintf("%x", sha256.Sum256([]byte(req.Password))) { ... }
```

```go
// хорошо
if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)) != nil {
	return ErrUnauthorized
}
```

И ещё: сообщение об ошибке входа одинаковое для «нет такого email» и «неверный пароль». Иначе по разным ответам злоумышленник узнаёт, какие email зарегистрированы.

Одинакового текста мало. Если email не найден, наивный код сразу возвращает 401, не вызывая bcrypt, и ответ приходит на те же 60 мс быстрее. Такую разницу видно по сети без всякой статистики. Поэтому для несуществующего пользователя тоже сравнивают пароль, только с заранее посчитанным фиктивным хешем того же cost:

```go
var dummyHash, _ = bcrypt.GenerateFromPassword([]byte("dummy-password"), bcrypt.DefaultCost)

func (s *Auth) Login(ctx context.Context, email, password string) (User, error) {
	u, err := s.users.GetByEmail(ctx, email)
	if errors.Is(err, ErrNotFound) {
		_ = bcrypt.CompareHashAndPassword(dummyHash, []byte(password)) // выравниваем время ответа
		return User{}, ErrUnauthorized
	}
	if err != nil {
		return User{}, err
	}
	if bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)) != nil {
		return User{}, ErrUnauthorized
	}
	return u, nil
}
```
:::

## Сессии против JWT

Есть два способа помнить, что пользователь вошёл.

**Сессия.** Сервер создаёт случайный идентификатор, хранит его в базе или Redis вместе с `user_id` и отдаёт клиенту в cookie. На каждый запрос сервер ищет сессию по id. **JWT.** Сервер подписывает токен с `user_id` и ролью внутри и отдаёт клиенту. На каждый запрос сервер проверяет подпись и верит содержимому, в базу не ходит.

| | Сессия в хранилище | JWT |
|---|---|---|
| Проверка запроса | поход в Redis или базу | только проверка подписи |
| Мгновенный выход, бан | удалил сессию, готово | токен живёт до `exp`, нужен denylist |
| Масштабирование | общее хранилище сессий | любой сервис с ключом проверит сам |
| Размер | 32 байта id | сотни байт в каждом запросе |
| Ошибки реализации | мало | много (алгоритм, exp, хранение) |

Честно: для монолита с одним фронтендом сессии в httpOnly cookie часто проще и безопаснее. JWT выигрывает, когда токен проверяют **несколько сервисов** без общего хранилища или когда клиент не браузер (мобильное приложение, другой бэкенд). Популярный компромисс, который мы и сделаем: короткий JWT (access) плюс длинный refresh-токен в базе. Access проверяется без базы, а refresh можно отозвать.

## Как устроен JWT

JWT (RFC 7519) это три части через точку: `header.payload.signature`. Каждая часть закодирована base64url без паддинга.

```text
eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9 . eyJzdWIiOiI0MiIsInJvbGUiOiJ1c2VyIn0 . 3q2-7w...
{"alg":"HS256","typ":"JWT"}            {"sub":"42","role":"user"}            HMAC-SHA256(header.payload, secret)
```

Подпись защищает от **подмены**, но не от **чтения**: payload это просто base64, а не шифрование.

```go
package main

import (
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func main() {
	claims := jwt.MapClaims{"sub": "42", "role": "admin", "exp": time.Date(2026, 11, 5, 12, 0, 0, 0, time.UTC).Unix()}
	tok, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte("super-secret-key-super-secret-key"))
	payload := strings.Split(tok, ".")[1]
	b, _ := base64.RawURLEncoding.DecodeString(payload)
	fmt.Println(string(b))
}
```

<Predict :options="['ошибка: payload зашифрован', 'бинарный мусор', 'JSON с полями exp, role (admin) и sub (42)', 'пустая строка: без секрета не прочитать']" :correct="2">

Любой, у кого есть токен, читает payload без всякого секрета. Сам вывод: `{"exp":1793880000,"role":"admin","sub":"42"}`. Ключи отсортированы, потому что `MapClaims` это map, а `encoding/json` пишет ключи map по алфавиту. Вывод: в JWT не кладут ничего секретного (пароль, номер карты, внутренние флаги). Только идентификатор и роль.

</Predict>

### Стандартные поля (claims)

| Поле | Смысл | Что проверять |
|---|---|---|
| `sub` | subject, кто владелец (id пользователя) | не пустой |
| `exp` | expiration, после этого момента токен недействителен | обязателен, не наступил |
| `nbf` | not before, раньше этого момента недействителен | не в будущем |
| `iat` | issued at, когда выпущен | не в будущем |
| `iss` | issuer, кто выпустил | равен твоему сервису |
| `aud` | audience, для кого выпущен | содержит твой API |
| `jti` | уникальный id токена | нужен для denylist |

`iss` и `aud` защищают от случая, когда два сервиса компании подписывают токены одним ключом: токен от сервиса биллинга не должен открыть админку трекера.

### HS256, RS256, EdDSA

- **HS256** (HMAC-SHA256): один общий секрет и подписывает, и проверяет. Просто и быстро. Минус: каждый, кто может проверить токен, может и выпустить его. Годится, когда токены выпускает и проверяет один сервис. Секрет длинный и случайный, минимум 32 байта: короткий секрет перебирают офлайн по одному перехваченному токену.
- **RS256** (RSA) и **EdDSA** (Ed25519): асимметричные. Подписывает только сервис аутентификации приватным ключом, остальные проверяют публичным (часто публикуется по адресу JWKS). EdDSA даёт короткие ключи и подписи и быстро подписывает (а вот проверка RSA-2048 обычно не медленнее).

### Атаки на алгоритм

Алгоритм указан **в самом токене**, то есть его выбирает тот, кто прислал токен. Отсюда две классические атаки.

1. **`alg: none`.** Спецификация разрешает неподписанные токены. Злоумышленник ставит `"alg":"none"`, убирает подпись и пишет в payload `"role":"admin"`. Если библиотека слепо следует заголовку, токен принят. В `golang-jwt/jwt/v5` `none` по умолчанию запрещён: он работает только если ключ равен специальной константе `jwt.UnsafeAllowNoneSignatureType`.
2. **Подмена алгоритма (key confusion).** Сервер ждёт RS256 и проверяет публичным ключом. Злоумышленник берёт этот публичный ключ (он же публичный!) и подписывает токен как **HS256**, используя ключ как HMAC-секрет. Наивная библиотека возьмёт «ключ проверки» и посчитает HMAC с ним. В Go-библиотеке ключ типизирован: HMAC ждёт `[]byte`, а RSA `*rsa.PublicKey`, поэтому такая атака упрётся в `ErrInvalidKeyType`. Но только пока ты сам не отдаёшь `[]byte` для любого алгоритма.

Защита одна и обязательная: **явно разреши ровно свой алгоритм**.

```go
type Claims struct {
	Role string `json:"role"`
	jwt.RegisteredClaims
}

func Issue(secret []byte, userID, role string, now time.Time) (string, error) {
	c := Claims{
		Role: role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			Issuer:    "tasktracker",
			Audience:  jwt.ClaimStrings{"tasktracker-api"},
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(15 * time.Minute)),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, c).SignedString(secret)
}

func Parse(secret []byte, token string) (*Claims, error) {
	var c Claims
	_, err := jwt.ParseWithClaims(token, &c,
		func(*jwt.Token) (any, error) { return secret, nil },
		jwt.WithValidMethods([]string{"HS256"}), // только наш алгоритм
		jwt.WithExpirationRequired(),           // без exp токен вечный
		jwt.WithIssuer("tasktracker"),
		jwt.WithAudience("tasktracker-api"),
		jwt.WithIssuedAt(),
	)
	if err != nil {
		return nil, err
	}
	return &c, nil
}
```

`ParseWithClaims` по умолчанию проверяет `exp` и `nbf` **только если они есть**, а `iat` не проверяет вовсе: для него нужна опция `jwt.WithIssuedAt()`. Токен без `exp` пройдёт проверку и будет действовать вечно, поэтому `WithExpirationRequired` нужен, а токен с `iat` из будущего без `WithIssuedAt` тоже будет принят. В тестах время подменяют опцией `jwt.WithTimeFunc`.

::: trap Проверка без WithValidMethods
```go
// плохо: принимаем любой HMAC-алгоритм, который назвал клиент
jwt.Parse(tok, func(t *jwt.Token) (any, error) { return secret, nil })
```

Токен, подписанный тем же секретом алгоритмом HS512 или HS384, пройдёт. Сегодня это «не страшно», а завтра кто-то добавит второй ключ, и путаница алгоритмов станет дырой. Хорошо: `jwt.WithValidMethods([]string{"HS256"})` и `jwt.WithExpirationRequired()` в каждом вызове. Удобнее один раз создать `jwt.NewParser(...)` с опциями и переиспользовать.
:::

## Access и refresh токены

Один долгоживущий JWT плох: украли его, и злоумышленник работает неделями, а отозвать нельзя. Один короткий тоже плох: пользователь логинится каждые 15 минут. Решение из двух токенов:

- **Access**: JWT, живёт 5-15 минут, ходит в каждом запросе в `Authorization: Bearer ...`. Проверяется без базы.
- **Refresh**: случайная строка (не JWT), живёт дни или недели, хранится в базе, используется только для получения новой пары на `POST /auth/refresh`.

```text
клиент                            сервер
  | POST /auth/login ------------->  проверить пароль
  | <------- access(15м) + refresh(30д), refresh -> БД (хеш)
  | GET /tasks (Bearer access) --->  проверить подпись, без БД
  | ... через 15 минут 401 ...
  | POST /auth/refresh (refresh) ->  найти хеш в БД, пометить использованным,
  | <------- новый access + НОВЫЙ refresh       выдать новую пару (ротация)
```

### Ротация и обнаружение кражи

**Ротация**: каждый refresh-токен одноразовый. Обменял его, получил новый, старый больше не работает. Что это даёт? Если злоумышленник украл refresh и воспользовался им, настоящий клиент придёт со **старым** (уже использованным) токеном. Сервер видит повторное использование и понимает: случилась кража. Правильная реакция: отозвать **всё семейство** токенов этого входа (цепочку ротаций одного устройства), чтобы и вор, и клиент потеряли доступ, и пользователь залогинился заново.

### Отзыв

- Refresh отзывается просто: удалить или пометить запись в базе. «Выйти на всех устройствах» это `DELETE FROM refresh_tokens WHERE user_id = $1`.
- Access отозвать нельзя, он проверяется без базы. Варианты: держать его коротким (15 минут неприятны, но терпимы); хранить denylist `jti` отозванных токенов в Redis до их `exp`; держать в пользователе `token_version` и класть её в токен (бан пользователя увеличивает версию, но тогда снова нужен поход в базу или кеш).

::: hood Почему refresh хешируют SHA-256, а не bcrypt
В базе лежит не сам refresh-токен, а его SHA-256. Утечка таблицы не даёт готовых токенов. Но почему не bcrypt, как для паролей? Медленный хеш нужен, когда исходное значение **слабое**: пароль можно угадать по словарю. Refresh-токен это 32 случайных байта из `crypto/rand`, 256 бит энтропии, перебрать нечего. Быстрый хеш тут достаточен и вдобавок позволяет искать запись по индексу: `WHERE token_hash = $1`. С bcrypt так не выйдет, у каждого хеша своя соль.
:::

## Где хранить токены на клиенте

Тут ты как фронтендер видел все варианты. Посмотрим на них глазами атакующего.

**localStorage или sessionStorage.** Удобно: `axios.defaults.headers.Authorization = 'Bearer ' + token`. Но любой JavaScript на странице читает localStorage. Одна **XSS** (внедрённый скрипт через уязвимый пакет из npm, неэкранированный комментарий, скомпрометированный CDN) и токены улетают на сервер атакующего. Там ими пользуются, даже когда вкладка закрыта.

**Cookie с флагами `HttpOnly; Secure; SameSite`.** `HttpOnly` запрещает JavaScript читать cookie, поэтому XSS не может его **украсть** (но может делать запросы от имени пользователя, пока страница открыта). Браузер сам прикладывает cookie к запросам, а это открывает **CSRF**: чужой сайт отправляет форму на твой API, браузер прикладывает cookie. Защита: `SameSite=Lax` или `Strict` (браузер не шлёт cookie с чужих сайтов на большинство запросов), CSRF-токен для изменяющих запросов, проверка `Origin`.

```go
c.SetSameSite(http.SameSiteStrictMode)
c.SetCookie("refresh_token", token, int((30 * 24 * time.Hour).Seconds()),
	"/v1/auth", "", true, true) // path только для эндпоинтов auth, Secure, HttpOnly
```

Частая рабочая схема для SPA: access-токен в памяти JS (переменная, не storage, пропадает при перезагрузке), refresh-токен в httpOnly cookie с узким `Path=/v1/auth`. После перезагрузки страница вызывает `/auth/refresh` и получает новый access. Ещё надёжнее паттерн BFF (backend for frontend): токены вообще не попадают в браузер, их держит маленький сервер рядом с фронтендом, а браузер общается с ним по сессионной cookie.

Никакой вариант не спасает, если XSS уже есть: это главный вывод. Хранение в cookie уменьшает ущерб (нельзя унести токен), но не отменяет необходимость экранирования и Content Security Policy.

## Middleware аутентификации и авторизации

В Gin проверка токена естественно ложится в middleware (как работают цепочки обработчиков, ты видел в [уроке про middleware](/m08/l03-middleware)). Хорошая практика: middleware не знает про JWT, а получает функцию проверки. Так его легко тестировать с фейком.

```go
type User struct{ ID, Role string }

type Verifier func(token string) (User, error)

func Authenticate(verify Verifier) gin.HandlerFunc {
	return func(c *gin.Context) {
		scheme, token, ok := strings.Cut(c.GetHeader("Authorization"), " ")
		if !ok || !strings.EqualFold(scheme, "Bearer") || token == "" {
			c.Header("WWW-Authenticate", "Bearer")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		u, err := verify(token)
		if err != nil {
			c.Header("WWW-Authenticate", "Bearer")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		c.Set("user", u)
		c.Next()
	}
}

func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		v, ok := c.Get("user")
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		if !slices.Contains(roles, v.(User).Role) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
		c.Next()
	}
}

// подключение
api := r.Group("/v1", Authenticate(verifyJWT))
api.GET("/tasks/:id", getTask)
admin := api.Group("/admin", RequireRole("admin"))
admin.DELETE("/users/:id", deleteUser)
```

Тела ошибок здесь и в домашке упрощены до `{"error":"unauthorized"}`, чтобы не отвлекать от авторизации. В проекте используй единый формат `{"error":{"code","message","fields"}}` из [урока про binding](/m08/l02-binding).

Схема по RFC 9110 сравнивается без учёта регистра (`bearer` тоже валиден), отсюда `strings.EqualFold`. `AbortWithStatusJSON` останавливает цепочку, но код **после** него в этом же middleware всё равно выполнится, поэтому `return` обязателен.

Роли проверяет middleware, а **владельца ресурса** только обработчик или сервис: middleware не знает, чья задача номер 7. Об этом следующий раздел.

::: trap Middleware, который пропускает без токена
```go
// плохо: «если токена нет, пусть обработчик сам разберётся»
func Authenticate(c *gin.Context) {
	h := c.GetHeader("Authorization")
	if h == "" {
		c.Next()
		return
	}
	...
}
```

Обработчик получит пустого пользователя, `userID` станет нулём, и запрос `WHERE owner_id = 0` либо вернёт пусто, либо (с `LEFT JOIN` и `OR`) что-то лишнее. Для защищённых групп отсутствие токена всегда 401. Если нужны публичные маршруты, вешай на них другой middleware (`OptionalAuth`) явно.
:::

## OWASP API Security Top 10

OWASP (некоммерческий проект по безопасности веба) ведёт отдельный рейтинг рисков для API. Редакция 2023 года, кратко и с тем, как это выглядит в Go:

1. **API1 Broken Object Level Authorization (BOLA, она же IDOR).** `GET /tasks/7` отдаёт задачу, не проверив, что она твоя. Самая частая уязвимость.
2. **API2 Broken Authentication.** Слабые пароли без лимита попыток, JWT без проверки алгоритма или `exp`, токены в URL.
3. **API3 Broken Object Property Level Authorization.** Отдаём лишние поля (`password_hash` в JSON) или принимаем лишние (mass assignment: `{"title":"x","owner_id":1}` и `ShouldBindJSON` прямо в модель).
4. **API4 Unrestricted Resource Consumption.** Нет лимитов: `?limit=1000000`, загрузка файла на 10 ГБ, бесконечный перебор пароля.
5. **API5 Broken Function Level Authorization.** Обычный пользователь вызывает `DELETE /admin/users/1`, потому что на группе забыли `RequireRole`.
6. **API6 Unrestricted Access to Sensitive Business Flows.** Боты выкупают все билеты или регистрируют тысячи аккаунтов: бизнес-процесс не защищён от автоматизации.
7. **API7 Server Side Request Forgery.** API скачивает картинку по URL пользователя, а тот передаёт `http://169.254.169.254/` (метаданные облака) или внутренний сервис.
8. **API8 Security Misconfiguration.** CORS `*` с credentials, стектрейсы в ответах, отладочный режим в проде.
9. **API9 Improper Inventory Management.** Забытая `/v0` без проверок, тестовый стенд с боевыми данными.
10. **API10 Unsafe Consumption of APIs.** Доверяем ответу стороннего API без валидации.

BOLA в коде выглядит невинно:

```go
// плохо: проверяем только существование
func (r *Repo) GetTask(ctx context.Context, id int64) (Task, error) {
	return scan(r.pool.QueryRow(ctx, `SELECT ... FROM tasks WHERE id = $1`, id))
}

// хорошо: владелец часть запроса, чужая задача неотличима от несуществующей
func (r *Repo) GetTask(ctx context.Context, ownerID, id int64) (Task, error) {
	return scan(r.pool.QueryRow(ctx,
		`SELECT ... FROM tasks WHERE id = $1 AND owner_id = $2`, id, ownerID))
}
```

Обрати внимание: `ownerID` берётся **только** из проверенного токена (`c.Get("user")`), никогда из тела, query или заголовка вроде `X-User-ID`. Проверка владельца в самом SQL лучше, чем «достал задачу и сравнил в Go»: нет окна гонки, и невозможно забыть сравнение в одной из веток.

::: trap Mass assignment через binding в модель
```go
// плохо: клиент прислал {"title":"x","owner_id":1,"role":"admin"}
var t Task
c.ShouldBindJSON(&t) // owner_id перезаписан из тела
repo.Update(ctx, t)
```

```go
// хорошо: отдельная структура запроса только с разрешёнными полями
var in struct {
	Title string `json:"title" binding:"required,max=200"`
}
if err := c.ShouldBindJSON(&in); err != nil { ... }
t.Title = in.Title // владелец остаётся прежним
```

То же в обратную сторону: модель с `PasswordHash` не отдают в `c.JSON` напрямую, либо отдельная структура ответа, либо тег `json:"-"`.
:::

## Rate limiting

Ограничение частоты запросов защищает от перебора паролей (API2), от перегрузки (API4) и от ботов (API6). Классический алгоритм **token bucket**: у клиента ведро на `b` жетонов, жетоны добавляются со скоростью `r` в секунду, каждый запрос тратит один. Пустое ведро значит отказ с кодом **429 Too Many Requests** и заголовком `Retry-After`.

В Go он готов: `golang.org/x/time/rate`.

```go
package main

import (
	"fmt"
	"time"

	"golang.org/x/time/rate"
)

func main() {
	lim := rate.NewLimiter(1, 3) // 1 токен в секунду, ведро на 3
	now := time.Now()
	for range 5 {
		fmt.Print(lim.AllowN(now, 1), " ")
	}
	fmt.Println(lim.AllowN(now.Add(2*time.Second), 1))
}
```

Разбор этой программы в разделе «Что выведет программа?» ниже. Метод `AllowN(t, n)` принимает время явно, что удобно в тестах: не нужно ждать секунды.

Один лимитер на весь сервер бессмыслен: один бот съест лимит всех пользователей. Лимит считают **по ключу**: по IP для анонимных запросов (логин, регистрация), по `user_id` для аутентифицированных. Значит, нужна карта лимитеров под мьютексом:

```go
type KeyedLimiter struct {
	mu       sync.Mutex
	limiters map[string]*rate.Limiter
	r        rate.Limit
	b        int
}

func (l *KeyedLimiter) Allow(key string) bool {
	l.mu.Lock()
	lim, ok := l.limiters[key]
	if !ok {
		lim = rate.NewLimiter(l.r, l.b)
		l.limiters[key] = lim
	}
	l.mu.Unlock()
	return lim.Allow() // сам Limiter потокобезопасен
}
```

У этой версии есть потолок: карта растёт бесконечно, ключ на каждый IP. В проде старые ключи вытесняют (LRU, периодическая чистка по времени последнего доступа), а при нескольких репликах сервиса лимит держат в Redis, иначе у каждой реплики свой счётчик и реальный лимит умножается на число реплик.

::: trap ClientIP и X-Forwarded-For
Лимит по IP в Gin делают через `c.ClientIP()`. По умолчанию Gin **доверяет всем прокси** и берёт IP из заголовка `X-Forwarded-For`, а его клиент может прислать сам. Злоумышленник ставит случайный `X-Forwarded-For` в каждый запрос и обходит лимит.

```go
r := gin.New()
// хорошо: никому не доверяем, IP берётся из TCP-соединения
_ = r.SetTrustedProxies(nil)
// или доверяем только своему балансировщику
_ = r.SetTrustedProxies([]string{"10.0.0.0/8"})
```

Разбор на примере в «Что выведет программа?».
:::

## Сравнение за постоянное время

Обычное сравнение строк `a == b` останавливается на первом несовпавшем байте. Время ответа чуть-чуть зависит от того, сколько первых байт угадано. На этом строится **timing attack**: измеряя время ответа на миллионы запросов, можно подбирать секрет (API-ключ, подпись вебхука) по байту. По сети это трудно, но в одной облачной зоне измеримо.

Секреты сравнивают функцией из `crypto/subtle`:

```go
// плохо
if r.Header.Get("X-Api-Key") == cfg.APIKey { ... }

// хорошо
if subtle.ConstantTimeCompare([]byte(r.Header.Get("X-Api-Key")), []byte(cfg.APIKey)) == 1 { ... }
```

Тонкость из документации: при разной длине `ConstantTimeCompare` возвращает 0 **сразу**, то есть длина секрета утекает. Если это важно, сравнивай хеши фиксированной длины: `sha256.Sum256` от обеих строк, затем `ConstantTimeCompare`. Для HMAC-подписей есть готовая `hmac.Equal`. А bcrypt и `jwt` сравнивают за постоянное время внутри, отдельно ничего делать не нужно.

Побайтная утечка через `==` тонкая и требует миллионов измерений. Куда грубее утечка на уровне логики: вход с несуществующим email отвечает на десятки миллисекунд быстрее, потому что bcrypt не вызывался. Это лечится фиктивным сравнением с заранее посчитанным хешем, как в ловушке из раздела про пароли: время ответа не должно зависеть от того, существует ли пользователь.

## Что выведет программа?

Программа с лимитером из раздела выше:

```go
lim := rate.NewLimiter(1, 3) // 1 токен в секунду, ведро на 3
now := time.Now()
for range 5 {
	fmt.Print(lim.AllowN(now, 1), " ")
}
fmt.Println(lim.AllowN(now.Add(2*time.Second), 1))
```

<Predict :options="['true true true true true true', 'true false false false false true', 'true true true false false true', 'true true true false false false']" :correct="2">

Ведро изначально полное: 3 жетона, поэтому первые три запроса в один и тот же момент проходят (это burst), четвёртый и пятый нет. Через 2 секунды накопилось 2 новых жетона, шестой запрос проходит. Обрати внимание, burst это не «три запроса в секунду», а «сколько можно сразу после простоя».

</Predict>

```go
r := gin.New()
r.GET("/ip", func(c *gin.Context) { c.String(200, c.ClientIP()) })

req := httptest.NewRequest("GET", "/ip", nil)
req.RemoteAddr = "203.0.113.5:51000"
req.Header.Set("X-Forwarded-For", "1.2.3.4")
w := httptest.NewRecorder()
r.ServeHTTP(w, req)
fmt.Println(w.Body.String())
```

<Predict :options="['203.0.113.5', '1.2.3.4', '203.0.113.5:51000', 'пустая строка']" :correct="1">

Gin по умолчанию доверяет всем прокси, поэтому `ClientIP` берёт адрес из `X-Forwarded-For`, который прислал сам клиент. Лимит по IP обходится одной строкой заголовка. После `r.SetTrustedProxies(nil)` та же программа печатает `203.0.113.5`.

</Predict>

```go
long := strings.Repeat("я", 37)
_, err := bcrypt.GenerateFromPassword([]byte(long), 10)
fmt.Println(len(long), err)
```

<Predict :options="['37 <nil>', '74 <nil>', '74 bcrypt: password length exceeds 72 bytes', '37 bcrypt: password length exceeds 72 bytes']" :correct="2">

`len` строки считает байты, а «я» в UTF-8 занимает 2 байта: 37 символов это 74 байта. Это больше лимита bcrypt, и `x/crypto/bcrypt` возвращает `ErrPasswordTooLong` вместо молчаливой обрезки. Ограничение пароля в форме «не длиннее 72 символов» для кириллицы не работает.

</Predict>

::: interview
<details><summary>Чем аутентификация отличается от авторизации? Когда 401, когда 403, когда 404?</summary>

Аутентификация отвечает «кто ты» (проверка пароля или токена), авторизация «что тебе можно» (роль, владение ресурсом). Нет токена или он невалиден: 401 с заголовком `WWW-Authenticate`. Пользователь известен, но действие запрещено ролью: 403. Объект существует, но чужой: обычно 404, чтобы не подтверждать существование объекта и не помогать перебору id.

</details>
<details><summary>Как правильно хранить пароли? Почему не SHA-256 с солью?</summary>

Медленной функцией с настраиваемой стоимостью: bcrypt (cost от 10), argon2id (минимум по OWASP 19 МиБ памяти, 2 итерации), scrypt. SHA-256 быстрый: GPU считает миллиарды хешей в секунду, и соль не мешает перебору каждого хеша отдельно. У bcrypt соль и cost хранятся внутри хеша, сравнение через `CompareHashAndPassword` за постоянное время. Ограничение bcrypt: 72 байта пароля.

</details>
<details><summary>Сессии или JWT? Плюсы и минусы.</summary>

Сессия: случайный id в cookie, состояние в Redis или базе. Мгновенный отзыв, мало способов ошибиться, но поход в хранилище на каждый запрос. JWT: подписанный токен с данными, проверяется без хранилища, удобно для нескольких сервисов и не браузерных клиентов. Минусы: нельзя мгновенно отозвать, payload читается всеми, много мест для ошибки (алгоритм, `exp`). На практике часто комбинируют: короткий access JWT плюс refresh в базе.

</details>
<details><summary>Как устроен JWT и что в нём нельзя хранить?</summary>

Три части base64url через точку: заголовок с алгоритмом, payload с claims, подпись от первых двух частей. Подпись защищает от изменения, но не от чтения, поэтому секреты (пароли, персональные данные сверх необходимого, внутренние флаги) в payload не кладут. Проверять надо подпись, конкретный алгоритм, `exp` (и его наличие), `nbf`, `iss`, `aud`, а `iat` в golang-jwt проверяется только с `WithIssuedAt`.

</details>
<details><summary>Что за атаки alg=none и подмена алгоритма? Как защититься в Go?</summary>

Алгоритм указан в токене, то есть выбран клиентом. `none` означает «без подписи»: уязвимая библиотека примет любой payload. Подмена RS256 на HS256: злоумышленник подписывает токен публичным ключом как HMAC-секретом, наивная проверка его примет. Защита: явно разрешить свой алгоритм (`jwt.WithValidMethods([]string{"HS256"})`), не отдавать из keyfunc ключ для чужих алгоритмов. В golang-jwt `none` по умолчанию отключён, а HMAC требует ключ типа `[]byte`.

</details>
<details><summary>Зачем refresh-токен и что такое ротация?</summary>

Access живёт минуты, чтобы кража была недолгой, refresh живёт дни и нужен только для получения новой пары. Ротация: каждый refresh одноразовый, при обмене выдаётся новый. Если старый токен пришёл повторно, значит его украли: отзываем всё семейство токенов этого входа. В базе refresh хранится как SHA-256 хеш.

</details>
<details><summary>Где хранить токены в браузере?</summary>

localStorage читается любым JavaScript, поэтому XSS уносит токены. httpOnly cookie недоступна скриптам, но открывает CSRF, его закрывают `SameSite`, CSRF-токеном и проверкой `Origin`. Частая схема: access в памяти, refresh в httpOnly Secure SameSite cookie с узким `Path`. Ещё лучше BFF, где токены не попадают в браузер. Ни одна схема не спасает от уже существующей XSS полностью.

</details>
<details><summary>Что такое BOLA/IDOR и как от неё защищаются?</summary>

Broken Object Level Authorization: API отдаёт или меняет объект по id, не проверив, что он принадлежит пользователю. Первое место в OWASP API Top 10. Защита: владелец берётся только из проверенного токена и входит в сам SQL-запрос (`WHERE id = $1 AND owner_id = $2`), чужой объект возвращает 404. Плюс тесты, где пользователь B запрашивает объект пользователя A.

</details>
<details><summary>Как сделать rate limiting и какие подводные камни?</summary>

Token bucket из `golang.org/x/time/rate`, отдельный лимитер на ключ (IP или user id), ответ 429 с `Retry-After`. Камни: общий лимитер на всех; IP из `X-Forwarded-For` без настройки доверенных прокси; карта ключей растёт без очистки; при нескольких репликах лимит в памяти умножается на число реплик, нужен Redis.

</details>
<details><summary>Зачем subtle.ConstantTimeCompare, если есть ==?</summary>

`==` прекращает сравнение на первом отличии, и по времени ответа можно подбирать секрет побайтно (timing attack). `ConstantTimeCompare` тратит время, зависящее только от длины. При разной длине он возвращает 0 сразу, поэтому длину секрета можно скрыть, сравнивая хеши фиксированной длины. Для HMAC есть `hmac.Equal`.

</details>
:::

## Источники

- `RU` [Пять простых шагов для понимания JSON Web Tokens (JWT), Хабр](https://habr.com/ru/articles/340146/): устройство заголовка, payload и подписи на пальцах.
- `RU` [Добавляем Refresh Token, Хабр](https://habr.com/ru/post/466929/): зачем нужен второй токен и как выглядит обмен.
- `RU` [Вы кто такие, я вас не знаю, или Как мы делаем JWT-аутентификацию, Хабр](https://habr.com/ru/articles/764424/): практика с access и refresh, отзыв токенов, несколько устройств; комментарии к статье тоже полезны.
- `EN` [OWASP Password Storage Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Password_Storage_Cheat_Sheet.html): параметры argon2id, bcrypt, scrypt и PBKDF2, про лимит 72 байта и pepper.
- `EN` [OWASP API Security Top 10, 2023](https://api-security.owasp.org/editions/2023/en/0x11-t10): все десять рисков с примерами атак, начни с API1 и API3.
- `EN` [RFC 8725: JSON Web Token Best Current Practices](https://www.rfc-editor.org/rfc/rfc8725): официальный список ошибок с JWT, раздел 2 про атаки.
- `EN` [golang-jwt/jwt v5](https://pkg.go.dev/github.com/golang-jwt/jwt/v5): `ParseWithClaims`, опции парсера `WithValidMethods`, `WithExpirationRequired`, `WithTimeFunc`.
- `EN` [golang.org/x/time/rate](https://pkg.go.dev/golang.org/x/time/rate): `Limiter`, `Allow`, `AllowN`, `Reserve` и как устроен token bucket.

## Домашка

Пакет `homework/m08/l05-auth/`. Тесты с bcrypt идут несколько секунд: так и задумано, bcrypt медленный.

<Hw id="ex1" title="Хеширование паролей" level="база">

Файл `ex1_password.go`: `HashPassword` и `CheckPassword` на bcrypt. Пустой пароль, слабый cost и пароль длиннее 72 байт дают ошибку, сравнение не через `==`.

```bash
cd homework && go test -v ./m08/l05-auth/ -run 'TestHashPassword|TestCheckPassword'
```

</Hw>

<Hw id="ex2" title="Выпуск и проверка JWT" level="посложнее">

Файл `ex2_jwt.go`: `IssueAccess` и `ParseAccess` на `golang-jwt/jwt/v5`. Тесты подсовывают просроченный токен, `alg=none`, HS512 тем же секретом, подменённый payload, чужие `iss` и `aud`, токен без `exp`. Всё это должно отклоняться с `ErrInvalidToken`.

```bash
cd homework && go test -v ./m08/l05-auth/ -run 'TestIssueAccess|TestParseAccess'
```

</Hw>

<Hw id="ex3" title="Middleware: 401 и 403" level="база">

Файл `ex3_middleware.go`: `Authenticate` и `RequireRole` для Gin. Без токена 401 с `WWW-Authenticate`, чужая роль 403, `RequireRole` без `Authenticate` тоже 401, цепочка прерывается.

Формат тела ошибки упрощён, в проекте используй единый формат из [урока про binding](/m08/l02-binding).

```bash
cd homework && go test -v ./m08/l05-auth/ -run TestAuthMiddleware
```

</Hw>

<Hw id="ex4" title="Закрываем IDOR" level="посложнее">

Файл `ex4_idor.go`: обработчики `GET` и `PATCH /tasks/:id`. Чужая задача даёт 404, admin может читать, но не править чужое, владельца нельзя сменить через тело запроса.

Формат тела ошибки упрощён, в проекте используй единый формат из [урока про binding](/m08/l02-binding).

```bash
cd homework && go test -v ./m08/l05-auth/ -run TestTaskRoutesIDOR
```

</Hw>

<Hw id="ex5" title="Rate limiter по ключу" level="посложнее">

Файл `ex5_ratelimit.go`: `KeyedLimiter` с отдельным ведром на ключ, middleware `RateLimit` с 429 и `Retry-After`, функция ключа `KeyByUserOrIP`. Тесты проверяют разделение по ключам и гонки (`-race`).

Формат тела ошибки упрощён, в проекте используй единый формат из [урока про binding](/m08/l02-binding).

```bash
cd homework && go test -race -v ./m08/l05-auth/ -run 'Limit|KeyBy'
```

</Hw>

<Hw id="ex6" title="Refresh-токены с ротацией" level="со звёздочкой">

Файл `ex6_refresh.go`: `RefreshStore` выдаёт непрозрачные токены, хранит только их хеши, ротирует при обмене, при повторном использовании отзывает всё семейство и умеет «выйти на всех устройствах». Два одновременных обмена одного токена: успешен ровно один.

```bash
cd homework && go test -race -v ./m08/l05-auth/ -run TestRefresh
```

</Hw>

<Hw id="ex7" title="Сколько стоит cost" level="база">

Без тестов. Напиши бенчмарк с `b.Loop()` (Go 1.24+), который хеширует пароль bcrypt с cost 10, 11, 12, 13 и 14 (подтесты `b.Run`). Посмотри, как растёт время, и выбери cost, при котором один хеш занимает 100-300 мс на твоей машине. Подумай, сколько входов в секунду выдержит одно ядро с таким cost и что это значит для лимита на `/auth/login`.

</Hw>
