# sqlc, ORM, миграции

В прошлом уроке ты писал SQL строками и сканировал результат руками. Для пяти запросов это нормально, для двухсот уже утомительно: опечатка в имени колонки находится только в рантайме, порядок полей в `Scan` легко перепутать, а меняющаяся схема ломает код молча. Экосистема Go предлагает несколько способов с этим жить: query builder, ORM и кодогенерацию. Споры «ORM или чистый SQL» бесконечны, и на собесе важен не «правильный» ответ, а понимание компромиссов.

Вторая половина урока про **миграции**: как схема базы меняется вместе с кодом и почему безобидный `ALTER TABLE` может положить продакшен на несколько минут. Это то, что джуниор делает в первые недели работы, и то, где ошибка стоит дороже всего.

## Четыре способа работать с SQL из Go

### Сырой SQL

То, что ты уже умеешь: pgx, запрос строкой, `CollectRows`.

```go
rows, _ := pool.Query(ctx, `SELECT id, title, done FROM tasks WHERE project_id = $1 ORDER BY id`, projectID)
tasks, err := pgx.CollectRows(rows, pgx.RowToStructByName[Task])
```

Плюсы: полный контроль, видно, какой запрос уходит в базу, доступны все возможности Postgres (CTE, оконные функции, `ON CONFLICT`, `jsonb`). Минусы: ошибки в SQL и несовпадение колонок с полями находятся только при выполнении, много повторяющегося кода, сложно собирать запросы с необязательными фильтрами.

### Query builder: squirrel

Query builder собирает текст SQL из вызовов методов. Он не знает твою схему и не прячет SQL, а просто избавляет от склейки строк:

```go
import sq "github.com/Masterminds/squirrel"

psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar) // $1 вместо ?

q := psql.Select("id", "title", "done").From("tasks").OrderBy("id")
if projectID != 0 {
	q = q.Where(sq.Eq{"project_id": projectID})
}
if search != "" {
	q = q.Where("title ILIKE ?", "%"+search+"%")
}
sql, args, err := q.Limit(20).ToSql()
fmt.Println(sql)
fmt.Println(args, err)
// SELECT id, title, done FROM tasks WHERE project_id = $1 AND title ILIKE $2 ORDER BY id LIMIT 20
// [5 %отчёт%] <nil>
```

Дальше `sql` и `args` идут в обычный `pool.Query`. Удобно `sq.Eq` со срезом: `sq.Eq{"id": []int64{1, 2, 3}}` даёт `id IN ($1,$2,$3)`. Ниша query builder это динамические запросы: фильтры, сортировки и пагинация из параметров API. Типобезопасности он не добавляет: опечатка в `"titel"` всё так же всплывёт только в рантайме. В домашке ты напишешь такой мини-builder сам, без зависимостей.

### ORM: GORM

ORM (Object-Relational Mapping) отображает таблицы на структуры и генерирует SQL за тебя. Самая популярная в Go это **GORM**:

```go
type Project struct {
	ID    int64
	Name  string
	Tasks []Task // связь has many по полю ProjectID
}

type Task struct {
	ID        int64
	ProjectID int64
	Title     string
	Done      bool
}

db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
var projects []Project
db.Preload("Tasks").Find(&projects)
```

Плюсы: CRUD без SQL, связи, хуки, быстрый старт. Минусы, за которые ORM в Go недолюбливают:
- SQL скрыт, и легко написать код, который делает сотни запросов (N+1, ниже);
- рефлексия и `interface{}` повсюду: ошибки типов только в рантайме;
- неочевидное поведение. Например, `db.Model(&u).Updates(User{Name: "", Age: 0})` **не обновит** поля с нулевыми значениями: GORM считает их «не заданными» (это поведение описано в документации). `db.First(&t, 99)` при отсутствии записи вернёт ошибку `record not found` (`gorm.ErrRecordNotFound`), а `db.Find` не вернёт;
- сложный запрос (оконные функции, CTE) всё равно пишется строкой через `db.Raw`.

### Кодогенерация: sqlc

**sqlc** переворачивает подход: ты пишешь обычный SQL, а sqlc разбирает его вместе со схемой и **генерирует типобезопасный Go-код**. Опечатка в колонке или несовпадение типа это ошибка генерации, а не продовый баг.

Конфиг `sqlc.yaml`:

```yaml
version: "2"
sql:
  - engine: "postgresql"
    schema: "db/migrations"     # sqlc понимает миграции goose и читает только Up
    queries: "db/queries"
    gen:
      go:
        package: "store"
        out: "internal/store"
        sql_package: "pgx/v5"   # генерировать код под pgx, а не database/sql
```

Запросы с аннотациями, `db/queries/tasks.sql`:

```sql
-- name: GetTask :one
SELECT * FROM tasks WHERE id = $1;

-- name: ListProjectTasks :many
SELECT id, title, done FROM tasks
WHERE project_id = $1
ORDER BY id;

-- name: CreateTask :one
INSERT INTO tasks (project_id, title, due_at)
VALUES ($1, $2, $3)
RETURNING *;

-- name: CompleteTask :execrows
UPDATE tasks SET done = true WHERE id = $1 AND NOT done;
```

Аннотация после имени задаёт форму результата: `:one` (одна строка), `:many` (срез), `:exec` (только ошибка), `:execrows` (число затронутых строк). Генерация, **ничего не ставя глобально**:

```bash
go run github.com/sqlc-dev/sqlc/cmd/sqlc@v1.31.1 generate
```

Первый запуск компилирует sqlc (у меня около 40 секунд: внутри парсер PostgreSQL на C), дальше берётся из кеша. Чтобы у всей команды была одна версия, закрепи sqlc в `go.mod` как инструмент (Go 1.24+):

```bash
go get -tool github.com/sqlc-dev/sqlc/cmd/sqlc@v1.31.1
go tool sqlc generate
```

Учти, что это добавит в `go.mod` заметное число косвенных зависимостей sqlc; их не будет в бинарнике сервиса, но в `go.mod` они видны.

Что получится в `internal/store/tasks.sql.go` (фрагмент сгенерированного кода):

```go
type CreateTaskParams struct {
	ProjectID int64
	Title     string
	DueAt     pgtype.Timestamptz
}

func (q *Queries) CreateTask(ctx context.Context, arg CreateTaskParams) (Task, error) {
	row := q.db.QueryRow(ctx, createTask, arg.ProjectID, arg.Title, arg.DueAt)
	var i Task
	err := row.Scan(
		&i.ID,
		&i.ProjectID,
		&i.Title,
		&i.Done,
		&i.DueAt,
	)
	return i, err
}

type ListProjectTasksRow struct {
	ID    int64
	Title string
	Done  bool
}

func (q *Queries) ListProjectTasks(ctx context.Context, projectID int64) ([]ListProjectTasksRow, error) {
```

`due_at` в схеме допускает `NULL`, поэтому sqlc выбрал `pgtype.Timestamptz` (с полем `Valid`). Для `SELECT *` он сам раскрыл список колонок, а для выборки части колонок создал отдельную структуру `ListProjectTasksRow`. Использование:

```go
q := store.New(pool) // принимает и пул, и pgx.Tx: те же методы работают в транзакции
task, err := q.CreateTask(ctx, store.CreateTaskParams{ProjectID: 1, Title: "Написать тесты"})
```

Слабое место sqlc: **динамические запросы**. Запрос с пятью необязательными фильтрами в sqlc превращается в условия вида `(sqlc.narg('done')::bool IS NULL OR done = sqlc.narg('done'))`, которые хуже используют индексы. Для таких мест сочетают sqlc с ручным builder.

### Что выбрать

| | Сырой SQL | squirrel | GORM | sqlc |
|---|---|---|---|---|
| Видно, какой SQL уходит | да | почти | нет | да |
| Ошибки колонок и типов | рантайм | рантайм | рантайм | генерация |
| Динамические фильтры | неудобно | отлично | хорошо | неудобно |
| Сложный SQL (CTE, окна) | да | частично | через Raw | да |
| Шаблонный код | много | средне | мало | генерируется |
| Риск N+1 | низкий | низкий | высокий | низкий |

Типичный ответ на собесе, который звучит взросло: «в Go-сообществе предпочитают явный SQL. Я бы взял sqlc для основной массы запросов, а для эндпоинтов со сложной фильтрацией собирал запрос builder-ом. GORM оправдан для быстрого прототипа или админки с простым CRUD». Важно уметь объяснить, **почему**, а не просто назвать инструмент.

## Проблема N+1

Нужно показать список проектов с задачами. Очевидный код:

```go
var projects []Project
db.Find(&projects)
for i := range projects {
	db.Where("project_id = ?", projects[i].ID).Find(&projects[i].Tasks)
}
```

Лог GORM для трёх проектов:

```
SELECT * FROM "projects"
SELECT * FROM "tasks" WHERE project_id = 1
SELECT * FROM "tasks" WHERE project_id = 2
SELECT * FROM "tasks" WHERE project_id = 3
```

Один запрос за списком и **N** запросов за деталями. На трёх проектах незаметно, на тысяче это тысяча сетевых походов в базу по 0.3-1 мс: секунда на ровном месте. Каждый запрос по отдельности быстрый, поэтому в логах медленных запросов N+1 не видно. Особенно легко его получить в ORM с ленивой загрузкой связей и в GraphQL, где резолвер поля вызывается для каждого объекта.

Решения:

1. **Второй запрос со списком id** (так делает `Preload`):

```
SELECT * FROM "tasks" WHERE "tasks"."project_id" IN (1,2,3)
SELECT * FROM "projects"
```

GORM выполнил запрос за проектами и один за задачами всех проектов, а потом разложил задачи по проектам в памяти (лог печатает строку после завершения запроса, поэтому порядок в логе обратный). В pgx то же самое: `WHERE project_id = ANY($1)` со срезом `[]int64` и раскладка через `map[int64]*Project`.

2. **JOIN** одним запросом и группировка строк в Go. Меньше походов, но данные проекта повторяются в каждой строке задачи.

3. **Агрегация в базе**: `json_agg` или `array_agg` по задачам в подзапросе, на выходе одна строка на проект.

Как ловить N+1:
- логировать запросы в dev-окружении и смотреть глазами;
- в тестах считать запросы. В pgx для этого есть **трейсер**: интерфейс, который pgx вызывает перед и после каждого запроса.

```go
type queryCounter struct{ n atomic.Int64 }

func (c *queryCounter) TraceQueryStart(ctx context.Context, _ *pgx.Conn, _ pgx.TraceQueryStartData) context.Context {
	c.n.Add(1)
	return ctx
}

func (c *queryCounter) TraceQueryEnd(context.Context, *pgx.Conn, pgx.TraceQueryEndData) {}

cfg, _ := pgxpool.ParseConfig(url)
cfg.ConnConfig.Tracer = &queryCounter{}
```

Тест домашки устроен именно так. Тот же механизм используют для логирования запросов и для трейсинга OpenTelemetry;
- на проде в `pg_stat_statements`: запрос с огромным `calls` и маленьким `mean_exec_time` это частый признак N+1.

::: trap Запрос в цикле
Плохо: запрос к базе внутри цикла по результатам другого запроса.

```go
for i := range projects {
	rows, _ := pool.Query(ctx, `SELECT id, title, done FROM tasks WHERE project_id = $1`, projects[i].ID)
	projects[i].Tasks, _ = pgx.CollectRows(rows, pgx.RowToStructByPos[Task])
}
```

Хорошо: собрать id и сделать один запрос.

```go
rows, _ := pool.Query(ctx, `SELECT project_id, id, title, done FROM tasks WHERE project_id = ANY($1) ORDER BY id`, ids)
```
:::

## Миграции

Схема базы меняется вместе с кодом: новая таблица, колонка, индекс. Менять её руками в `psql` на каждом окружении нельзя: забудешь, ошибёшься, и staging будет отличаться от прода. **Миграции** это пронумерованные файлы с изменениями схемы, которые хранятся в git рядом с кодом и применяются инструментом в строгом порядке. Инструмент помнит в служебной таблице, какие миграции уже применены.

### goose

**goose** (уже в `homework/go.mod`) поддерживает миграции на SQL и на Go. SQL-миграция это файл `00002_create_tasks.sql`:

```sql
-- +goose Up
CREATE TABLE tasks (
	id         bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
	project_id bigint  NOT NULL REFERENCES projects (id) ON DELETE CASCADE,
	title      text    NOT NULL,
	done       boolean NOT NULL DEFAULT false
);
CREATE INDEX tasks_project_id_idx ON tasks (project_id);

-- +goose Down
DROP TABLE tasks;
```

- **Up** применяет изменение, **Down** отменяет его. Down должен возвращать схему ровно в состояние до Up.
- Номер в имени задаёт порядок. Вместо порядковых номеров можно использовать время (`20261105120000_create_tasks.sql`), меньше конфликтов в команде, но порядок мерджа становится важнее.
- Каждая миграция по умолчанию выполняется в **транзакции**: в Postgres DDL транзакционный, и упавшая посередине миграция откатится целиком. Это большой плюс Postgres перед MySQL.
- Функции и блоки с `;` внутри оборачивают в `-- +goose StatementBegin` и `-- +goose StatementEnd`.
- Команды, которые нельзя выполнять в транзакции (`CREATE INDEX CONCURRENTLY`), требуют `-- +goose NO TRANSACTION` в начале файла.

Запускать можно CLI (`go run github.com/pressly/goose/v3/cmd/goose@v3.29.0 -dir db/migrations postgres "$DATABASE_URL" up`), но в Go-сервисе удобнее встроить миграции в бинарник через `embed` и применять библиотекой:

```go
package migrations

import "embed"

//go:embed *.sql
var FS embed.FS
```

```go
db, err := sql.Open("pgx", url) // goose работает через database/sql
p, err := goose.NewProvider(goose.DialectPostgres, db, migrations.FS)
results, err := p.Up(ctx)
for _, r := range results {
	fmt.Println(r)
}
```

```
OK    up 00001_create_projects.sql (6.35ms)
OK    up 00002_create_tasks.sql (6.51ms)
OK    up 00003_backfill.go (2.75ms)
```

Состояние хранится в таблице `goose_db_version` (`version_id`, `is_applied`, `tstamp`). Методы провайдера: `Up`, `UpTo(v)`, `UpByOne`, `Down`, `DownTo(v)`, `Status`, `GetDBVersion`.

**Go-миграции** нужны, когда изменение не выражается SQL: перенести данные с вызовом Go-логики, перешифровать поле. Они регистрируются в `init` и получают транзакцию:

```go
func init() {
	goose.AddMigrationContext(upBackfill, downBackfill)
}

func upBackfill(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.ExecContext(ctx, `UPDATE projects SET name = trim(name)`)
	return err
}

func downBackfill(ctx context.Context, tx *sql.Tx) error { return nil }
```

Версия берётся из имени файла (`00003_backfill.go`). Если несколько инстансов сервиса стартуют одновременно и каждый запускает миграции, включи блокировку: `goose.WithSessionLocker(locker)`, где `locker` из `lock.NewPostgresSessionLocker()`. Она использует advisory lock из урока про транзакции.

### Миграции только вперёд

Down удобен локально, но на проде миграции почти никогда не откатывают:
- `DROP COLUMN` в Down безвозвратно удаляет данные, записанные после Up;
- новый код к моменту отката уже мог записать данные, которые старая схема не вмещает;
- откат схемы требует отката кода, и наоборот: появляется окно несовместимости.

Поэтому ошибку в применённой миграции исправляют **новой** миграцией. И главное правило: **никогда не редактируй уже применённую миграцию**. Она выполнилась на проде и у коллег, изменённый файл никто не перезапустит, и окружения разойдутся.

## Миграции без простоя

На проде таблица живая: в неё пишут и из неё читают сотни запросов в секунду. Почти любой `ALTER TABLE` берёт блокировку **ACCESS EXCLUSIVE**, несовместимую ни с чем, даже с `SELECT`. Если сам `ALTER` выполняется миллисекунды, это не страшно. Страшно ожидание.

Сессия A читает таблицу в долгой транзакции (отчёт, забытый `BEGIN`). Сессия B выполняет миграцию. Сессия C это обычный запрос пользователя:

```sql
A> BEGIN;
A> SELECT count(*) FROM accounts;            -- держит ACCESS SHARE до конца транзакции
B> ALTER TABLE accounts ADD COLUMN note text; -- ждёт A
C> SELECT balance FROM accounts WHERE id = 1; -- ждёт... B!
```

Кто кого ждёт (из третьего окна, через `pg_blocking_pids`):

```
 сессия | ждёт | запрос
--------+------+-------------------------------------------
 B      | {A}  | ALTER TABLE accounts ADD COLUMN note text
 C      | {B}  | SELECT balance FROM accounts WHERE id = 1
```

`SELECT` из C совместим с блокировкой A, но **встаёт в очередь за ALTER**: Postgres не пропускает новые запросы вперёд ожидающей эксклюзивной блокировки, иначе она никогда бы её не получила. Пока A не закончит, все запросы к таблице стоят. Для сервиса это выглядит как полный отказ.

Защита: **короткий `lock_timeout`** перед DDL и повтор при неудаче.

```sql
B> SET lock_timeout = '100ms';
B> ALTER TABLE accounts ADD COLUMN note text;
   ERROR:  canceling statement due to lock timeout
```

Миграция упала через 100 мс, запросы пользователей простояли не дольше этого. Инструмент повторит её через несколько секунд, когда долгая транзакция закончится. В goose это `SET LOCAL lock_timeout` в начале SQL-миграции: миграция идёт в транзакции, и `LOCAL` ограничивает настройку этой транзакцией. Простой `SET` пережил бы `COMMIT` и остался на соединении: если goose работает на общем `*sql.DB` сервиса, соединение вернулось бы в пул с таймаутом блокировок для всех будущих запросов.

::: hood Какие изменения схемы дешёвые
Дешёвые (меняют только каталог, миллисекунды под блокировкой):
- `ADD COLUMN` без `DEFAULT` или с **константным** `DEFAULT` (с PostgreSQL 11 значение хранится в каталоге и подставляется при чтении старых строк);
- `DROP COLUMN` (колонка помечается удалённой, место освобождается позже);
- `ADD CONSTRAINT ... NOT VALID`, переименования, `DROP DEFAULT`.

Дорогие (переписывают таблицу или читают её целиком под блокировкой):
- `ADD COLUMN` с **изменчивым** `DEFAULT` (`random()`, `gen_random_uuid()`);
- `ALTER COLUMN TYPE` с несовместимым представлением (`int` в `bigint`, `text` в `int`);
- `SET NOT NULL` и `ADD CONSTRAINT CHECK`/`FOREIGN KEY` без `NOT VALID`: проверка всех строк;
- `CREATE INDEX` без `CONCURRENTLY`: блокирует запись на всё построение.

Проверка на таблице в миллион строк:

```sql
ALTER TABLE big ADD COLUMN priority int NOT NULL DEFAULT 0;                 -- 1.1 ms, файл тот же
ALTER TABLE big ADD COLUMN token int NOT NULL DEFAULT (random() * 100)::int; -- 557 ms, таблица переписана
```

После второй команды у таблицы новый `pg_relation_filenode`: Postgres создал новый файл и переписал в него все строки, держа ACCESS EXCLUSIVE.
:::

Безопасные рецепты:

- **Индекс**: `CREATE INDEX CONCURRENTLY` в отдельной миграции с `-- +goose NO TRANSACTION`. В транзакции он не работает: `ERROR: CREATE INDEX CONCURRENTLY cannot run inside a transaction block`. Если упал, остаётся невалидный индекс: `DROP INDEX CONCURRENTLY` и заново.
- **NOT NULL или CHECK на большой таблице**: `ADD CONSTRAINT ... CHECK (col IS NOT NULL) NOT VALID` (мгновенно, проверяет только новые строки), потом отдельной миграцией `VALIDATE CONSTRAINT` (читает таблицу, но под слабой блокировкой, не мешая записи). С PostgreSQL 12 после валидного `CHECK (col IS NOT NULL)` команда `SET NOT NULL` не перепроверяет таблицу.
- **Заполнение данных** (backfill): не одним `UPDATE` на миллионы строк (долгая транзакция, блокировки, bloat), а пачками по несколько тысяч строк в отдельных транзакциях.
- **Переименование или смена типа колонки**: схема **expand/contract**. Добавить новую колонку, писать в обе, перенести данные, переключить чтение на новую, перестать писать в старую, удалить старую. Каждый шаг отдельным релизом, и в любой момент старая и новая версии кода совместимы со схемой.

::: trap Правка применённой миграции
Плохо: заметить ошибку в `00005_add_status.sql`, который уже ушёл на прод, и исправить файл. У тебя локально всё хорошо (ты пересоздал базу), а на проде goose считает версию 5 применённой и новый текст не выполнит никогда.

Хорошо: новая миграция `00006_fix_status.sql`, которая доводит схему до нужного состояния.
:::

::: trap Тяжёлый DDL без lock_timeout
Плохо:

```sql
-- +goose Up
ALTER TABLE orders ADD COLUMN source text;
```

Безобидно, пока на `orders` нет долгой транзакции. Если есть, миграция встанет в очередь и остановит все запросы к таблице.

Хорошо:

```sql
-- +goose Up
SET LOCAL lock_timeout = '3s';
ALTER TABLE orders ADD COLUMN source text;
```

Не получилось за 3 секунды, миграция упадёт, деплой повторит её позже. Пользователи заметят пару секунд задержки, а не минуты простоя.
:::

## Что выведет программа?

```go
psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
sql, args, _ := psql.Select("id").From("tasks").Where(sq.Eq{"id": []int64{1, 2, 3}}).ToSql()
fmt.Println(sql, args)
```

<Predict :options="['SELECT id FROM tasks WHERE id = ANY($1) [[1 2 3]]', 'SELECT id FROM tasks WHERE id IN ($1,$2,$3) [1 2 3]', 'SELECT id FROM tasks WHERE id IN (1,2,3) []', 'SELECT id FROM tasks WHERE id = $1 [[1 2 3]]']" :correct="1">

`sq.Eq` со срезом раскрывается в `IN` с отдельным плейсхолдером на каждый элемент. Это безопасно, но на тысяче id получится тысяча параметров (у Postgres лимит 65535). В pgx для больших списков лучше `= ANY($1)` с одним параметром-массивом.

</Predict>

```go
// в базе 3 проекта
var projects []Project
db.Find(&projects)
for i := range projects {
	db.Where("project_id = ?", projects[i].ID).Find(&projects[i].Tasks)
}
// сколько SQL-запросов выполнит GORM?
```

<Predict :options="['1', '2', '4', '6']" :correct="2">

Один `SELECT * FROM "projects"` и по одному запросу на каждый из трёх проектов: N+1 = 4. `db.Preload("Tasks").Find(&projects)` сделает 2 запроса при любом числе проектов.

</Predict>

```sql
-- в big 1 000 000 строк, сессия A держит открытую транзакцию после SELECT из big
B> ALTER TABLE big ADD COLUMN priority int NOT NULL DEFAULT 0;
C> SELECT * FROM big WHERE id = 1;
```

<Predict :options="['C выполнится сразу: SELECT совместим с блокировкой A', 'C ждёт, пока закончится A и пройдёт ALTER из B', 'B переписывает миллион строк, C ждёт этого', 'B получит ошибку deadlock']" :correct="1">

`ALTER` ждёт ACCESS EXCLUSIVE, потому что A держит ACCESS SHARE. C совместим с A, но встаёт в очередь за ожидающим B. Сам `ADD COLUMN ... DEFAULT 0` мгновенный (константный default), проблема только в очереди. Лечится `SET LOCAL lock_timeout` в миграции.

</Predict>

::: interview
<details><summary>ORM, query builder, sqlc, чистый SQL: в чём компромиссы?</summary>

Чистый SQL даёт контроль и прозрачность, но ошибки видны в рантайме и много шаблонного кода. Query builder (squirrel) удобен для динамических запросов, но типобезопасности не добавляет. ORM (GORM) экономит код на CRUD, но прячет SQL, провоцирует N+1 и полагается на рефлексию. sqlc генерирует типобезопасный код из SQL и проверяет его по схеме при генерации, но неудобен для динамических фильтров. Частая комбинация: sqlc плюс builder для сложных фильтров.

</details>
<details><summary>Что такое N+1 и как его обнаружить и исправить?</summary>

Один запрос за списком и по запросу на каждый элемент за связанными данными. Каждый запрос быстрый, но их число растёт с данными. Обнаружить: логи запросов, счётчик запросов в тестах (трейсер pgx), `pg_stat_statements` (много вызовов маленьких запросов). Исправить: второй запрос с `IN`/`ANY` по списку id (Preload), JOIN, агрегация в базе.

</details>
<details><summary>Как работает sqlc?</summary>

Читает схему (DDL или миграции) и файлы запросов с аннотациями `-- name: Имя :one/:many/:exec`, парсит SQL парсером Postgres, выводит типы параметров и колонок и генерирует Go-структуры и методы. Опечатка в имени колонки или несовпадение типов это ошибка генерации. Сгенерированные методы принимают интерфейс, которому удовлетворяют и пул, и транзакция.

</details>
<details><summary>Зачем нужны миграции и как goose понимает, какие уже применены?</summary>

Чтобы схема менялась воспроизводимо и одинаково на всех окружениях, вместе с кодом и через ревью. goose хранит применённые версии в таблице `goose_db_version` и при `up` применяет недостающие по возрастанию номера, каждую по умолчанию в транзакции.

</details>
<details><summary>Почему на проде миграции делают только вперёд?</summary>

Down часто необратимо теряет данные (DROP COLUMN), откат схемы требует синхронного отката кода, а данные, записанные новым кодом, могут не поместиться в старую схему. Поэтому ошибки чинят новой миграцией, а применённые миграции не редактируют.

</details>
<details><summary>Почему ALTER TABLE ADD COLUMN может положить прод, хотя он мгновенный?</summary>

Он берёт ACCESS EXCLUSIVE. Если на таблице есть долгая транзакция, ALTER ждёт её, а все новые запросы к таблице, включая SELECT, встают в очередь за ALTER. Защита: `SET LOCAL lock_timeout` на короткое время и повтор миграции, плюс отсутствие долгих транзакций.

</details>
<details><summary>Как добавить индекс на большую таблицу без простоя?</summary>

`CREATE INDEX CONCURRENTLY` в отдельной миграции вне транзакции (`-- +goose NO TRANSACTION`). Он строится дольше и в несколько проходов, но не блокирует запись. Если упал, остаётся невалидный индекс, его удаляют `DROP INDEX CONCURRENTLY` и создают заново.

</details>
<details><summary>Как переименовать колонку без простоя?</summary>

Через expand/contract: добавить новую колонку, писать в обе, перенести старые данные пачками, переключить чтение на новую, перестать писать в старую, удалить её. Каждый шаг отдельным релизом, чтобы старая и новая версии приложения в любой момент работали со схемой. Прямой `RENAME COLUMN` мгновенный, но ломает работающую старую версию кода.

</details>
:::

## Источники

- `RU` [Инструкции по GORM](https://gorm.io/ru_RU/docs/): официальная документация на русском; посмотри разделы «Preload» и «Обновление» (про нулевые значения).
- `RU` [SQL миграции в Postgres. Часть 1](https://habr.com/ru/articles/540500/): `lock_timeout`, очередь блокировок и безопасные приёмы на практике.
- `RU` [Распространённые ошибки изменения схемы базы данных PostgreSQL](https://habr.com/ru/articles/582698/): Николай Самохвалов о том, как миграции кладут прод, с примерами и рецептами.
- `RU` [Документация PostgreSQL: ALTER TABLE](https://postgrespro.ru/docs/postgresql/17/sql-altertable): какие формы команды какую блокировку берут и когда переписывают таблицу (раздел «Замечания»).
- `EN` [sqlc Documentation](https://docs.sqlc.dev/en/latest/): Getting started с pgx и справочник по аннотациям и `sqlc.arg`/`sqlc.narg`.
- `EN` [goose](https://pressly.github.io/goose/): документация по аннотациям SQL-миграций, Go-миграциям и Provider API.
- `EN` [squirrel](https://github.com/Masterminds/squirrel): README с примерами builder-а.

## Домашка

Пакет `homework/m07/l08-sqlc-orm-migrations/`. Перед запуском `cd homework && docker compose up -d`.

<Hw id="ex1" title="Миграции goose" level="база">

В `migrations/` три файла с пустыми `Up`/`Down`: таблица проектов, таблица задач с внешним ключом и индексом, колонка приоритета с `CHECK`. Тест через `goose.Provider` накатывает до версии 2, проверяет схему и поведение (UNIQUE, каскадное удаление), накатывает 3 на таблицу с данными, откатывает по одной до нуля и накатывает снова.

```bash
cd homework && go test -v ./m07/l08-sqlc-orm-migrations/ -run TestMigrations
```

</Hw>

<Hw id="ex2" title="Убери N+1" level="база">

`ex2_nplus1.go`: функция работает правильно, но делает 21 запрос. Перепиши так, чтобы запросов было не больше двух. Тест считает запросы через трейсер pgx.

```bash
cd homework && go test -v ./m07/l08-sqlc-orm-migrations/ -run TestProjectsWithTasks
```

</Hw>

<Hw id="ex3" title="Свой query builder" level="посложнее">

`ex3_filter.go`, функция `BuildTaskQuery`: необязательные фильтры, нумерация плейсхолдеров, буквальный поиск, лимит. Тест выполняет твой запрос в базе и сравнивает с эталоном.

```bash
cd homework && go test -v ./m07/l08-sqlc-orm-migrations/ -run TestBuildTaskQuery
```

</Hw>

<Hw id="ex4" title="sqlc вживую" level="посложнее">

Без тестов. В отдельной папке вне репозитория курса создай модуль, положи туда миграции из задачи 1 и `sqlc.yaml` из урока. Напиши запросы `CreateProject :one`, `ListProjects :many`, `CreateTask :one`, `ListTasksByProject :many`, `SetPriority :execrows`. Сгенерируй код через `go run github.com/sqlc-dev/sqlc/cmd/sqlc@v1.31.1 generate`, посмотри, какие типы sqlc выбрал для `priority` и `created_at`. Сделай опечатку в имени колонки и посмотри на ошибку генерации. Напиши `main.go`, который применяет миграции goose через `embed` и вызывает сгенерированные методы.

</Hw>

<Hw id="ex5" title="Очередь блокировок" level="со звёздочкой">

Без тестов. Воспроизведи сценарий из раздела «Миграции без простоя» в трёх окнах `psql`, посмотри `pg_blocking_pids` из четвёртого. Потом повтори с `SET lock_timeout = '100ms'` в окне миграции. Создай таблицу на миллион строк и замерь (`\timing on`) `ADD COLUMN` с константным и с изменчивым `DEFAULT`, проверь `pg_relation_filenode` до и после.

</Hw>
