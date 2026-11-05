# Go и БД: database/sql и pgx

До этого урока ты общался с Postgres через `psql`. Сервис делает то же самое из Go: открывает соединения, отправляет запросы, читает строки в структуры, управляет транзакциями. Код вокруг запроса кажется рутиной, но именно в нём живут самые частые продовые баги: пул соединений кончился и сервис встал, `NULL` уронил обработчик, транзакция не откатилась и держит блокировку, конкатенация строки в SQL открыла инъекцию.

В Go два основных пути к Postgres: стандартный интерфейс `database/sql` с драйвером и библиотека **pgx** в «родном» режиме. В курсе используем pgx v5, но `database/sql` нужно знать: на нём написана половина существующего кода, на нём работают goose и многие ORM, и про него спрашивают на собесе.

Все примеры запущены против учебной базы (`cd homework && docker compose up -d`), строка подключения:

```go
const url = "postgres://goschool:goschool@localhost:55432/goschool?sslmode=disable"
```

## database/sql против pgx

`database/sql` это не драйвер, а **общий интерфейс** стандартной библиотеки: пул соединений, `Query`, `Scan`, транзакции. Конкретную базу подключает драйвер, который регистрируется под именем при импорте:

```go
import (
	"database/sql"

	_ "github.com/jackc/pgx/v5/stdlib" // регистрирует драйвер "pgx"
)

db, err := sql.Open("pgx", url)
if err != nil {
	return err
}
defer db.Close()
if err := db.PingContext(ctx); err != nil { // Open соединений не открывает, проверяем явно
	return err
}
```

pgx можно использовать двумя способами: как драйвер для `database/sql` (выше) или напрямую через свой API:

```go
import "github.com/jackc/pgx/v5/pgxpool"

pool, err := pgxpool.New(ctx, url)
if err != nil {
	return err
}
defer pool.Close()
```

| | `database/sql` + драйвер | pgx напрямую |
|---|---|---|
| Базы | любые, код переносим | только PostgreSQL |
| Типы | через `driver.Value`: `sql.NullString`, `[]byte` | все типы Postgres: массивы, `jsonb`, `uuid`, диапазоны, `numeric` |
| Скорость | лишний слой конвертации | бинарный протокол, кеш подготовленных запросов |
| Возможности | общий знаменатель | `COPY`, пакеты запросов (batch), `LISTEN/NOTIFY`, трейсинг |
| Экосистема | goose, sqlx, GORM, большинство библиотек | sqlc (режим pgx/v5), River |

Правило: пишешь сервис под Postgres, бери `pgxpool`. Нужна библиотека, которая принимает `*sql.DB`, открой его через `pgx/v5/stdlib`: драйвер тот же, просто другой фасад. Старый драйвер `lib/pq` ещё встречается в легаси, новые проекты обычно берут pgx.

## Пул соединений

Соединение с Postgres дорогое: TCP (часто TLS), аутентификация, и на сервере под каждое соединение отдельный **процесс** с несколькими мегабайтами памяти. Открывать соединение на каждый запрос нельзя. Пул держит несколько открытых соединений и раздаёт их запросам: взял, выполнил, вернул.

И `*sql.DB`, и `*pgxpool.Pool` это пулы, безопасные для использования из многих горутин. Создаются **один раз** при старте сервиса и передаются в репозитории. Создавать пул в обработчике запроса это классическая ошибка новичка.

Настройки `pgxpool` (значения по умолчанию из исходников pgx v5.12):

| Поле `pgxpool.Config` | По умолчанию | Смысл |
|---|---|---|
| `MaxConns` | `max(4, runtime.NumCPU())` | потолок соединений |
| `MinConns` | 0 | сколько держать открытыми даже без нагрузки |
| `MaxConnLifetime` | 1 час | соединение старше пересоздаётся (балансировка, утечки памяти на сервере) |
| `MaxConnIdleTime` | 30 минут | простаивающее дольше закрывается |
| `HealthCheckPeriod` | 1 минута | как часто проверять простаивающие соединения |

```go
cfg, err := pgxpool.ParseConfig(url)
if err != nil {
	return nil, err
}
cfg.MaxConns = 20
cfg.MinConns = 2
cfg.MaxConnLifetime = 30 * time.Minute
pool, err := pgxpool.NewWithConfig(ctx, cfg)
```

То же можно задать в строке подключения: `...?pool_max_conns=20&pool_min_conns=2`. У `database/sql` аналоги: `db.SetMaxOpenConns`, `SetMaxIdleConns`, `SetConnMaxLifetime`, `SetConnMaxIdleTime`. Опасный дефолт: `MaxOpenConns = 0` значит «без ограничения», и под нагрузкой сервис может открыть больше соединений, чем разрешает `max_connections` сервера (по умолчанию 100).

### Что происходит, когда пул кончился

Запрос, которому не досталось соединения, **ждёт**, пока кто-то вернёт своё. Ждёт столько, сколько позволяет контекст:

```go
cfg.MaxConns = 2
small, _ := pgxpool.NewWithConfig(ctx, cfg)
c1, _ := small.Acquire(ctx) // заняли оба соединения
c2, _ := small.Acquire(ctx)

tctx, cancel := context.WithTimeout(ctx, 200*time.Millisecond)
defer cancel()
err := small.QueryRow(tctx, `SELECT 1`).Scan(&n)
fmt.Println(err) // через 200 мс: context deadline exceeded

s := small.Stat()
fmt.Printf("total=%d acquired=%d idle=%d max=%d emptyAcquire=%d canceledAcquire=%d\n",
	s.TotalConns(), s.AcquiredConns(), s.IdleConns(), s.MaxConns(),
	s.EmptyAcquireCount(), s.CanceledAcquireCount())
// total=2 acquired=2 idle=0 max=2 emptyAcquire=2 canceledAcquire=1
```

Без таймаута в контексте запрос висел бы вечно, а за ним выстраивались бы все новые HTTP-запросы. Поэтому метрики пула (`Stat()` у pgx, `DB.Stats()` у `database/sql`) выводят в мониторинг: растущий `EmptyAcquireCount` или `WaitCount` означает, что соединений не хватает или их кто-то не возвращает.

Сколько ставить `MaxConns`? Больше не значит быстрее: Postgres упирается в CPU и диски, и сотни одновременных запросов работают медленнее десятков. Отправная точка: несколько соединений на ядро сервера базы, поделённых на число инстансов сервиса. Если инстансов много, перед базой ставят **PgBouncer**, внешний пул соединений.

::: hood Почему у Postgres дорогие соединения
Postgres использует модель «процесс на соединение»: главный процесс `postmaster` на каждое подключение делает `fork`, и дальше все запросы этого клиента выполняет отдельный процесс ОС со своей памятью (кеш каталога, кеш планов, `work_mem` под сортировки). Тысяча соединений это тысяча процессов, переключения контекста и гигабайты памяти, даже если 990 из них простаивают. Поэтому пул на стороне приложения обязателен, а при большом числе инстансов нужен ещё и PgBouncer в режиме `transaction`: он отдаёт серверное соединение клиенту только на время транзакции.

Подвох PgBouncer в режиме `transaction`: подготовленные запросы и сессионные настройки (`SET`, сессионные advisory locks) живут на серверном соединении, а следующая транзакция может попасть на другое. pgx по умолчанию кеширует подготовленные запросы (`QueryExecModeCacheStatement`); со старым PgBouncer его переключают в `QueryExecModeExec` или `QueryExecModeSimpleProtocol` через `cfg.ConnConfig.DefaultQueryExecMode`. PgBouncer начиная с версии 1.21 умеет сам отслеживать подготовленные запросы протокольного уровня.
:::

## Контекст и таймауты

Каждый метод pgx принимает `context.Context`, в `database/sql` есть версии с суффиксом `Context`. Контекст делает две вещи: ограничивает ожидание соединения из пула и **отменяет запрос на сервере**:

```go
tctx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
defer cancel()
_, err := pool.Exec(tctx, `SELECT pg_sleep(5)`)
fmt.Println(err, errors.Is(err, context.DeadlineExceeded))
// timeout: context deadline exceeded true    (через 100 мс, а не через 5 с)
```

При отмене контекста pgx отправляет серверу отдельный запрос отмены (cancel request), и Postgres прерывает выполнение. Без этого «брошенный» клиентом запрос продолжал бы жрать ресурсы базы.

На практике контекст приходит из HTTP-запроса (`r.Context()`): клиент отвалился, запрос в базу отменился. Сверху добавляют свой таймаут на запрос к базе, а на сервере страхуются `statement_timeout`, чтобы ни один запрос не работал вечно.

::: trap Методы без контекста
Плохо: `db.Query(...)`, `db.Exec(...)`, `db.Begin()` из `database/sql`. Они используют `context.Background()`: ни отмены, ни таймаута.

Хорошо: только `QueryContext`, `ExecContext`, `BeginTx` с контекстом запроса. В pgx версий без контекста нет вообще, и это одна из причин его любить.
:::

## Чтение результатов

### Одна строка

```go
var name string
err := pool.QueryRow(ctx, `SELECT name FROM users WHERE id = $1`, id).Scan(&name)
if errors.Is(err, pgx.ErrNoRows) {
	return ErrNotFound
}
```

Если строк нет, `Scan` вернёт `pgx.ErrNoRows` (в `database/sql` это `sql.ErrNoRows`). В pgx v5 `pgx.ErrNoRows` оборачивает `sql.ErrNoRows`, поэтому `errors.Is(err, sql.ErrNoRows)` тоже `true`. Наружу из репозитория такие ошибки не выпускают: превращают в свою `ErrNotFound`, иначе обработчик HTTP начнёт зависеть от драйвера.

### Много строк: database/sql

```go
rows, err := db.QueryContext(ctx, `SELECT name, phone FROM users ORDER BY id`)
if err != nil {
	return nil, err
}
defer rows.Close() // вернуть соединение в пул
var out []Contact
for rows.Next() {
	var c Contact
	if err := rows.Scan(&c.Name, &c.Phone); err != nil {
		return nil, err
	}
	out = append(out, c)
}
if err := rows.Err(); err != nil { // ошибка посреди выборки видна только здесь
	return nil, err
}
```

Три обязательных детали: `defer rows.Close()`, проверка ошибки `Scan`, проверка `rows.Err()` после цикла. `rows.Next()` возвращает `false` и при конце данных, и при ошибке (например, обрыв соединения на середине), различить их можно только через `rows.Err()`.

::: trap Забытый rows.Close
`Rows` держит соединение, пока не дочитан до конца или не закрыт. Забыл `Close` при раннем выходе из цикла, и соединение не вернётся в пул:

```go
db.SetMaxOpenConns(2)
for range 2 {
	db.QueryContext(ctx, `SELECT id FROM users`) // rows не закрыты
}
err := db.QueryRowContext(lctx, `SELECT 1`).Scan(&n) // lctx с таймаутом 300 мс
fmt.Println(err)                                    // context deadline exceeded
fmt.Println(db.Stats().InUse, db.Stats().WaitCount) // 2 1
```

Два «утёкших» соединения, и сервис встал. Хорошо: `defer rows.Close()` сразу после проверки ошибки `Query`, всегда. `QueryRow` закрывает за собой сам.
:::

### Много строк: pgx

В pgx есть помощники, которые делают цикл, `Close` и `Err` за тебя:

```go
type User struct {
	ID    int64   `db:"id"`
	Email string  `db:"email"`
	Name  string  `db:"name"`
	Phone *string `db:"phone"`
}

rows, _ := pool.Query(ctx, `SELECT id, email, name, phone FROM users ORDER BY id`)
users, err := pgx.CollectRows(rows, pgx.RowToStructByName[User])
```

Ошибку `Query` можно не проверять отдельно: pgx отложит её в `rows`, и `CollectRows` её вернёт. Варианты:

- `pgx.RowToStructByName[T]`: колонки к полям по тегу `db` (или по имени поля без учёта регистра). Если колонки для поля нет, ошибка: `cannot find field name in returned row`. Есть `RowToStructByNameLax`, который такое прощает.
- `pgx.RowToStructByPos[T]`: по порядку колонок.
- `pgx.RowTo[int64]`: одна колонка в простой тип.
- `pgx.CollectOneRow`: первая строка (остальные молча игнорируются), `pgx.ErrNoRows`, если строк нет. `pgx.CollectExactlyOneRow` дополнительно вернёт `pgx.ErrTooManyRows`, если строк больше одной: его бери, когда лишняя строка означает баг.
- `pgx.ForEachRow(rows, []any{&a, &b}, func() error {...})`: обработать каждую строку без накопления в срез.

## NULL

`NULL` в SQL это «значения нет». В Go у `string` и `int64` нет такого состояния, и `Scan` честно отказывается:

```go
var phone string
err := db.QueryRowContext(ctx, `SELECT phone FROM users WHERE id = 2`).Scan(&phone)
// sql: Scan error on column index 0, name "phone": converting NULL to string is unsupported
err = pool.QueryRow(ctx, `SELECT phone FROM users WHERE id = 2`).Scan(&phone)
// can't scan into dest[0] (col: phone): cannot scan NULL into *string
```

Три способа принять `NULL`:

| Способ | Пример | Когда |
|---|---|---|
| указатель | `var p *string` -> `nil` | проще всего, хорошо ложится в JSON (`null`) |
| `sql.Null*` | `sql.NullString{String: "", Valid: false}` | `database/sql`, есть и `sql.Null[T]` (Go 1.22) |
| `pgtype.*` | `pgtype.Text{String: "", Valid: false}` | pgx, так генерирует sqlc |

Все три работают в обе стороны: при записи `nil`-указатель или `Valid: false` превратится в `NULL`. Если `NULL` для колонки не бывает, сделай её `NOT NULL` в схеме, тогда и в Go хватит обычного `string`. Альтернатива для чтения: `COALESCE(phone, '')` в запросе, но тогда ты теряешь разницу между «пусто» и «не указано».

## Ошибки базы

Ошибка, пришедшая от сервера, в pgx имеет тип `*pgconn.PgError` с кодом SQLSTATE:

```go
_, err := pool.Exec(ctx, `INSERT INTO users (email, name) VALUES ('anna@example.com', 'Анна 2')`)
var pgErr *pgconn.PgError
if errors.As(err, &pgErr) {
	fmt.Println(pgErr.Code, pgErr.ConstraintName)
	fmt.Println(pgErr.Message)
	fmt.Println(pgErr.Detail)
}
// 23505 users_email_key
// duplicate key value violates unique constraint "users_email_key"
// Key (email)=(anna@example.com) already exists.
```

Коды, которые нужно знать наизусть:

| Код | Имя | Что делать |
|---|---|---|
| `23505` | unique_violation | «email занят», HTTP 409 |
| `23503` | foreign_key_violation | ссылка на несуществующую запись, обычно 404 или 422 |
| `23502` | not_null_violation | баг валидации, 400 или 500 |
| `23514` | check_violation | нарушено бизнес-правило из `CHECK` |
| `40001` | serialization_failure | повторить транзакцию |
| `40P01` | deadlock_detected | повторить транзакцию |
| `55P03` | lock_not_available | `NOWAIT` не дождался или сработал `lock_timeout` |
| `57014` | query_canceled | запрос отменён сервером: `statement_timeout` или запрос отмены |

Константы есть в пакете `github.com/jackc/pgerrcode`, но строковых литералов достаточно. Не делай «проверить, что email свободен, потом вставить»: между SELECT и INSERT другой запрос успеет зарегистрировать тот же email. Вставляй сразу и распознавай `23505`: уникальность гарантирует база, а не код.

## SQL-инъекции и плейсхолдеры

```go
input := "' OR '1'='1"
q := fmt.Sprintf(`SELECT count(*) FROM users WHERE email = '%s'`, input)
pool.QueryRow(ctx, q).Scan(&n)
// SELECT count(*) FROM users WHERE email = '' OR '1'='1' -> 2
pool.QueryRow(ctx, `SELECT count(*) FROM users WHERE email = $1`, input).Scan(&n)
// placeholder -> 0
```

Склейка превратила данные пользователя в код: условие стало «пустой email ИЛИ истина». Через такую дыру читают чужие данные, обходят логин, удаляют таблицы.

**Плейсхолдеры** (`$1`, `$2` в Postgres, `?` в MySQL) передают значения отдельно от текста запроса. В расширенном протоколе Postgres текст и параметры вообще уходят разными сообщениями: сервер разбирает SQL без значений, и никакое значение не может стать частью синтаксиса. Это не экранирование, а разделение каналов.

Плейсхолдером нельзя передать **идентификатор**: имя таблицы, колонки, направление сортировки. Если сортировка приходит из query-параметра, выбирай из белого списка:

```go
orderBy := map[string]string{"name": "name", "created": "created_at"}[r.URL.Query().Get("sort")]
if orderBy == "" {
	orderBy = "id"
}
q := `SELECT id, name FROM users ORDER BY ` + orderBy // в строку попадает только наш литерал
```

Отдельная тонкость: `LIKE` и `ILIKE`. Плейсхолдер защищает от инъекции, но символы `%` и `_` внутри значения остаются **шаблонами**. Поиск по `"%"` вернёт всё. Их экранируют обратным слэшем перед передачей (это задача из домашки).

## Транзакции в Go

Правильный паттерн один, запомни его как рефлекс:

```go
tx, err := pool.Begin(ctx)
if err != nil {
	return err
}
defer tx.Rollback(ctx) // на любом раннем return; после Commit ничего не делает

if _, err := tx.Exec(ctx, `UPDATE accounts SET balance = balance - $2 WHERE id = $1`, from, amount); err != nil {
	return err
}
if _, err := tx.Exec(ctx, `UPDATE accounts SET balance = balance + $2 WHERE id = $1`, to, amount); err != nil {
	return err
}
return tx.Commit(ctx)
```

`defer tx.Rollback` гарантирует, что при любой ошибке, панике или забытом `return` транзакция закроется и соединение вернётся в пул. После успешного `Commit` вызов `Rollback` вернёт `pgx.ErrTxClosed`, который мы игнорируем. Внутри транзакции все запросы идут через `tx`, а не через `pool`: запрос через `pool` возьмёт **другое** соединение и окажется вне транзакции.

Уровень изоляции задаётся при старте:

```go
tx, err := pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.Serializable})
// database/sql:
tx, err := db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
```

В pgx есть помощник, который делает `Commit` при `nil` и `Rollback` при ошибке:

```go
err := pgx.BeginTxFunc(ctx, pool, pgx.TxOptions{IsoLevel: pgx.Serializable}, func(tx pgx.Tx) error {
	// работа через tx
	return nil
})
```

::: trap Транзакция без defer Rollback
Плохо:

```go
tx, _ := pool.Begin(ctx)
if _, err := tx.Exec(ctx, q1); err != nil {
	return err // транзакция открыта, соединение не вернулось в пул, блокировки висят
}
tx.Commit(ctx)
```

Хорошо: `defer tx.Rollback(ctx)` сразу после успешного `Begin`, и `return tx.Commit(ctx)` в конце, чтобы не потерять ошибку коммита (на Serializable она бывает именно там).
:::

### Ретрай serialization_failure

На Repeatable Read и Serializable Postgres отменяет конфликтующую транзакцию с `40001` (из прошлого урока). Повторять надо **всю** транзакцию: открыть заново, перечитать данные, принять решение заново. Удобно завернуть это в функцию. Каркас (целиком ты напишешь её в домашке):

```go
func WithRetry(ctx context.Context, pool *pgxpool.Pool, fn func(tx pgx.Tx) error) error {
	var err error
	for attempt := range MaxAttempts {
		// 1. ctx отменён: вернуть ctx.Err(), новую попытку не начинать
		// 2. транзакция Serializable: fn(tx), потом Commit; ошибка fn: Rollback
		//    (40001 приходит и из запросов внутри fn, и из самого COMMIT)
		// 3. ошибка не 40001 и не 40P01 (или nil): вернуть как есть
		// 4. пауза, зависящая от attempt, со случайным разбросом; ждать через select с ctx.Done()
	}
	return err // попытки кончились: последняя ошибка
}

func retryable(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && (pgErr.Code == "40001" || pgErr.Code == "40P01")
}
```

Детали, которые отличают рабочий ретрай от игрушечного: ограничение числа попыток, пауза со случайным разбросом (иначе конкуренты снова столкнутся), уважение к контексту, `errors.As` вместо сравнения строк, и `fn` без побочных эффектов вне базы (письмо, отправленное в первой попытке, не отзовёшь).

## Что выведет программа?

В таблице `users` у Бориса (`id = 2`) телефон `NULL`.

```go
var phone string
err := pool.QueryRow(ctx, `SELECT phone FROM users WHERE id = 2`).Scan(&phone)
fmt.Printf("%q %v\n", phone, err)
```

<Predict :options="['\'\' <nil>', '\'NULL\' <nil>', '\'\' can\'t scan into dest[0] (col: phone): cannot scan NULL into *string', 'паника']" :correct="2">

`string` не умеет хранить `NULL`, и pgx возвращает ошибку, а не пустую строку: так ты не перепутаешь «телефона нет» с «телефон пустой». Нужен `*string`, `pgtype.Text` или `sql.NullString`.

</Predict>

```go
err := pool.QueryRow(ctx, `SELECT phone FROM users WHERE id = 99`).Scan(&ps)
fmt.Println(errors.Is(err, pgx.ErrNoRows), errors.Is(err, sql.ErrNoRows))
```

<Predict :options="['true false', 'true true', 'false false', 'false true']" :correct="1">

В pgx v5 `ErrNoRows` оборачивает `sql.ErrNoRows`, поэтому код, написанный под `database/sql`, работает и с pgx. Наружу из репозитория лучше отдавать свою `ErrNotFound`.

</Predict>

```go
cfg.MaxConns = 2
small, _ := pgxpool.NewWithConfig(ctx, cfg)
c1, _ := small.Acquire(ctx)
c2, _ := small.Acquire(ctx)
err := small.QueryRow(ctx, `SELECT 1`).Scan(&n) // ctx без таймаута
```

<Predict :options="['ошибка: pool exhausted', 'откроется третье соединение', 'запрос будет ждать вечно', 'вернёт 1']" :correct="2">

Пул не открывает соединений сверх `MaxConns` и не возвращает ошибку сразу: запрос ждёт, пока кто-нибудь вызовет `Release`. Ограничить ожидание может только контекст: с `context.WithTimeout` на 200 мс будет `context deadline exceeded` через 200 мс.

</Predict>

```go
input := "' OR '1'='1"
q := fmt.Sprintf(`SELECT count(*) FROM users WHERE email = '%s'`, input)
pool.QueryRow(ctx, q).Scan(&n)
fmt.Println(n) // в таблице 2 пользователя
```

<Predict :options="['0', '2', 'ошибка синтаксиса', '1']" :correct="1">

Запрос превращается в `WHERE email = '' OR '1'='1'`, условие истинно для всех строк. С плейсхолдером `$1` ответ 0: строка `' OR '1'='1'` ищется как обычный email.

</Predict>

::: interview
<details><summary>Чем database/sql отличается от pgx? Что выберешь?</summary>

`database/sql` это универсальный интерфейс стандартной библиотеки с пулом, драйвер подключается отдельно. pgx это драйвер и самостоятельная библиотека только для Postgres: поддержка всех типов, бинарный протокол, кеш подготовленных запросов, COPY, batch, LISTEN/NOTIFY, трейсинг. Для сервиса на Postgres выбираю `pgxpool`, а для библиотек, которые требуют `*sql.DB` (goose), открываю тот же pgx через `pgx/v5/stdlib`.

</details>
<details><summary>Как работает пул соединений и что будет, если он исчерпан?</summary>

Пул держит до `MaxConns` открытых соединений и выдаёт их запросам на время выполнения. Если все заняты, запрос ждёт освобождения, пока не истечёт контекст. Без таймаута сервис зависает. Причины исчерпания: медленные запросы, долгие транзакции, утечки (незакрытые `rows`, транзакции без `Rollback`), слишком маленький пул. Мониторится через `pool.Stat()` и `db.Stats()`.

</details>
<details><summary>Почему нельзя ставить MaxConns = 1000?</summary>

На сервере каждое соединение это процесс с памятью, `max_connections` по умолчанию 100. Даже если лимит поднять, Postgres упирается в CPU и диск, и тысяча одновременных запросов выполняется медленнее, чем несколько десятков в очереди. Размер пула подбирают от ядер базы и числа инстансов, при многих инстансах ставят PgBouncer.

</details>
<details><summary>Как прочитать NULL в Go?</summary>

В указатель (`*string`, `nil` для NULL), в `sql.NullString`/`sql.Null[T]` для database/sql или в `pgtype.Text` и другие типы pgtype для pgx. Обычный `string` вернёт ошибку сканирования. Если NULL по смыслу невозможен, лучше `NOT NULL` в схеме.

</details>
<details><summary>Напиши правильный каркас транзакции в Go.</summary>

`tx, err := pool.Begin(ctx)`, проверка ошибки, сразу `defer tx.Rollback(ctx)`, все запросы через `tx`, в конце `return tx.Commit(ctx)`. Rollback после Commit безвреден, а при любом раннем выходе транзакция гарантированно закроется и вернёт соединение. Альтернатива: `pgx.BeginTxFunc`.

</details>
<details><summary>Как защититься от SQL-инъекций? Чем плейсхолдер лучше экранирования?</summary>

Передавать значения только через плейсхолдеры (`$1`). При этом текст запроса и параметры уходят серверу раздельно, и значение физически не может стать частью синтаксиса, в отличие от экранирования, где ошибка в одном месте ломает всё. Идентификаторы (имя колонки для сортировки) плейсхолдером не передаются, их выбирают из белого списка. Для LIKE дополнительно экранируют `%` и `_`.

</details>
<details><summary>Как отличить ошибку уникальности от других ошибок базы?</summary>

`errors.As(err, &pgErr)` с `pgErr *pgconn.PgError` и проверка `pgErr.Code == "23505"`, при необходимости `pgErr.ConstraintName`. Сравнивать текст ошибки нельзя: он зависит от версии и локали сервера.

</details>
<details><summary>Как правильно повторять транзакцию при serialization_failure?</summary>

Повторять всю транзакцию с начала (новое `Begin`, перечитать данные), только на кодах 40001 и 40P01, ограничить число попыток, делать паузу с jitter, прекращать при отмене контекста. Побочные эффекты вне базы внутри повторяемой функции недопустимы.

</details>
<details><summary>Зачем передавать context в запросы к базе?</summary>

Контекст ограничивает и ожидание соединения из пула, и выполнение запроса: при отмене pgx шлёт серверу cancel request, и Postgres прерывает запрос. Когда клиент HTTP отвалился или истёк таймаут, база не тратит ресурсы на ненужную работу, а горутина не висит.

</details>
:::

## Источники

- `RU` [Работаем с PostgreSQL в Go. Опыт Авито](https://habr.com/ru/companies/avito/articles/716516/): pgx, размер пула, PgBouncer и грабли с отменой контекста из продакшена.
- `RU` [Как работать с Postgres в Go: практики, особенности, нюансы](https://habr.com/ru/companies/avito/articles/461935/): более ранняя статья Авито про database/sql и пул, полезна разделами про таймауты и метрики пула.
- `RU` [Документация PostgreSQL: Коды ошибок](https://postgrespro.ru/docs/postgresql/17/errcodes-appendix): полный список SQLSTATE.
- `EN` [pgx v5 на pkg.go.dev](https://pkg.go.dev/github.com/jackc/pgx/v5): читай раздел о CollectRows и типах, и отдельно [pgxpool](https://pkg.go.dev/github.com/jackc/pgx/v5/pgxpool) про настройки пула.
- `EN` [Accessing relational databases](https://go.dev/doc/database/): официальное руководство по database/sql, особенно [Managing connections](https://go.dev/doc/database/manage-connections) и [Avoiding SQL injection risk](https://go.dev/doc/database/sql-injection).

## Домашка

Пакет `homework/m07/l07-go-db/`: репозиторий пользователей и заказов. Тесты идут в настоящую базу, каждый тест в своей схеме. Перед запуском `cd homework && docker compose up -d`.

<Hw id="ex1" title="Пользователь по email и NULL" level="база">

`ex1_user.go`, метод `UserByEmail`: `CollectOneRow` + `RowToStructByName`, `NULL` в телефоне, своя `ErrNotFound` вместо `pgx.ErrNoRows`.

```bash
cd homework && go test -v ./m07/l07-go-db/ -run TestUserByEmail
```

</Hw>

<Hw id="ex2" title="Регистрация и 23505" level="база">

`ex2_create.go`, метод `CreateUser`: вставка с `RETURNING id`, распознавание нарушения уникальности через `*pgconn.PgError`, остальные ошибки обёрнуты. Отдельный тест регистрирует тот же email в соседней незакоммиченной транзакции: «проверить SELECT-ом, потом вставить» на нём падает.

```bash
cd homework && go test -v ./m07/l07-go-db/ -run TestCreateUser
```

</Hw>

<Hw id="ex3" title="Поиск без инъекций" level="посложнее">

`ex3_search.go`, метод `SearchUsers`: `ILIKE` с плейсхолдером и экранированием `%`, `_` и `\`. Тест пробует инъекции и шаблонные символы.

```bash
cd homework && go test -v ./m07/l07-go-db/ -run TestSearchUsers
```

</Hw>

<Hw id="ex4" title="Заказ в одной транзакции" level="база">

`ex4_order.go`, метод `CreateOrder`: заказ и позиции атомарно, `defer tx.Rollback`, `23503` в `ErrNotFound`. Тест проверяет, что после неудачной вставки в базе не остаётся следов и соединения вернулись в пул.

```bash
cd homework && go test -v ./m07/l07-go-db/ -run TestCreateOrder
```

</Hw>

<Hw id="ex5" title="Ретрай 40001" level="посложнее">

`ex5_retry.go`, функция `WithRetry`. Тесты устраивают настоящие конфликты сериализации: в одном между чтением и записью счётчика его меняет соседнее соединение (40001 на UPDATE), в другом write skew с соседней SERIALIZABLE-транзакцией (40001 на COMMIT).

```bash
cd homework && go test -race -v ./m07/l07-go-db/ -run TestWithRetry
```

</Hw>

<Hw id="ex6" title="То же на database/sql" level="база">

`ex6_database_sql.go`, функция `ListContacts`: классический цикл `rows.Next`/`Scan`/`rows.Err` и `sql.NullString`. Тест проверяет и то, что соединение вернулось в пул.

```bash
cd homework && go test -v ./m07/l07-go-db/ -run TestListContacts
```

</Hw>

<Hw id="ex7" title="Исчерпание пула своими глазами" level="со звёздочкой">

Без тестов. Напиши программу: пул с `MaxConns = 3`, 20 горутин, каждая делает `SELECT pg_sleep(0.5)` с таймаутом контекста 1 секунда. Сколько запросов успеет, сколько упадёт по таймауту? Выведи `pool.Stat()` до и после. Потом добавь в каждую горутину `pool.Begin` без `Rollback` и `Commit` и посмотри, что станет со следующими запросами.

</Hw>
