# River: основы

В прошлом уроке ты собрал свою очередь на `SKIP LOCKED` и увидел, сколько всего вокруг неё нужно: ретраи с backoff, возврат упавших задач, отложенный запуск, очистка старых строк, периодические задачи, уведомление воркеров о новой работе. **River** это готовая очередь задач для Go поверх PostgreSQL, где всё это уже сделано, а главная идея та же: задача живёт в твоей базе и ставится в **той же транзакции**, что и бизнес-данные.

Почему не Redis-очередь (asynq) и не брокер? Потому что именно транзакционная постановка убирает целый класс багов: «пользователь создан, а письмо не поставлено», «письмо отправлено пользователю, которого откатили». В River задача появляется в очереди тогда и только тогда, когда закоммитилась твоя транзакция, и не видна воркерам до коммита. На собеседовании это главный аргумент, его стоит уметь объяснить.

В этом уроке: как River устроен внутри, миграции, аргументы и воркеры, клиент и очереди, `Insert` и `InsertTx`, остановка. Версия в курсе: River **v0.51.0** (`homework/go.mod`). API River менялся между версиями, поэтому сверяйся с `go doc github.com/riverqueue/river` для своей версии, а не со старыми статьями.

## Как River устроен

Всё состояние River лежит в нескольких таблицах в твоей базе:

| Таблица | Что хранит |
|---|---|
| `river_job` | сами задачи: `kind`, `args` (jsonb), `state`, `attempt`, `max_attempts`, `priority`, `queue`, `scheduled_at`, `errors[]`, `unique_key` |
| `river_queue` | известные очереди и их пауза |
| `river_leader` | кто сейчас лидер кластера (одна строка) |
| `river_migration` | применённые версии схемы River |

### Состояния задачи

```text
          Insert                 Insert с ScheduledAt в будущем
            |                              |
            v                              v
       [available] <-- планировщик -- [scheduled]
            |                              ^
         забрал воркер                     | JobSnooze
            v                              |
        [running] --------------------------
         |   |   \______ ошибка, попытки есть ______> [retryable] --> (по времени) --> available
         |   |
         |   +-- JobCancel -----------------------> [cancelled]
         |   +-- ошибка, попытки кончились --------> [discarded]
         v
     [completed]
```

Ещё есть `pending`: задача вставлена, но не готова к работе, пока её не переведут явно (нужно для продвинутых сценариев). Финальные состояния (`completed`, `cancelled`, `discarded`) со временем удаляет чистильщик: по умолчанию завершённые и отменённые через 24 часа, отброшенные через 7 дней (`CompletedJobRetentionPeriod` и соседние поля конфига).

### Как воркер узнаёт о новой задаче

Каждый клиент с очередями запускает **producer** на очередь. Он забирает задачи тем же приёмом, что ты писал руками:

```sql
-- упрощённо, из исходников riverpgxv5
WITH locked_jobs AS (
	SELECT * FROM river_job
	WHERE state = 'available' AND queue = $1 AND scheduled_at <= now()
	ORDER BY priority ASC, scheduled_at ASC, id ASC
	LIMIT $2
	FOR UPDATE SKIP LOCKED
)
UPDATE river_job SET state = 'running', attempt = river_job.attempt + 1, attempted_at = now(), ...
FROM locked_jobs WHERE river_job.id = locked_jobs.id
RETURNING ...
```

Чтобы не опрашивать таблицу каждые 10 мс, River использует **LISTEN/NOTIFY** PostgreSQL. При вставке задачи клиент в той же транзакции делает `pg_notify` в канал вида `<схема>.river_insert`. Postgres доставляет уведомления только **после коммита** (откат их выбрасывает), поэтому воркеры узнают о задаче ровно тогда, когда она стала видимой. Каждый клиент держит одно соединение, которое слушает каналы. На случай пропущенных уведомлений есть **поллинг**: раз в `FetchPollInterval` (по умолчанию 1 с) producer всё равно заглядывает в таблицу. Между выборками выдерживается `FetchCooldown` (100 мс), это потолок частоты запросов.

::: hood Лидер и обслуживающие процессы
Часть работы нужна в кластере ровно в одном экземпляре: переводить `scheduled` и `retryable` в `available`, когда подошло время (scheduler, раз в 5 секунд); возвращать «застрявшие» `running`-задачи (rescuer, по умолчанию через час работы); удалять старые завершённые (cleaner); ставить периодические задачи; переиндексировать таблицы. Это делает **лидер**.

Выборы устроены на одной строке `river_leader`: каждый клиент периодически пытается `INSERT ... ON CONFLICT (name) DO NOTHING` со сроком `expires_at`. Кто вставил, тот лидер, и он продлевает срок `UPDATE ... WHERE leader_id = мой AND expires_at >= now()`. Лидер умер и не продлил: строка протухает, её удаляют, и следующий клиент становится лидером. Корректно уходящий лидер удаляет строку сам и шлёт NOTIFY, чтобы остальные не ждали. Отсюда практическое следствие: периодические задачи и ретраи работают, только пока жив хотя бы один клиент River, которому разрешено быть лидером.
:::

## Миграции

Таблицы River создаются его собственными миграциями. Версии ведёт River сам в `river_migration`, независимо от goose.

Через CLI:

```bash
go install github.com/riverqueue/river/cmd/river@latest
river migrate-up --line main --database-url "$DATABASE_URL"
river migrate-list --line main --database-url "$DATABASE_URL"
```

Из Go, например на старте сервиса или в тестах:

```go
migrator, err := rivermigrate.New(riverpgxv5.New(pool), nil)
if err != nil {
	return err
}
res, err := migrator.Migrate(ctx, rivermigrate.DirectionUp, nil)
```

Если миграции в проекте ведёт goose, можно выгрузить SQL River и положить его отдельными файлами: `river migrate-get --line main --all --exclude-version 1 --up > river_up.sql` (версия 1 это таблица учёта миграций самого River, goose она не нужна). Документация River просит сохранять границы версий: каждую в своей транзакции.

Схему River берёт из `search_path` соединения, поэтому в тестах домашки миграции идут в отдельную схему каждого теста (`testdb.New` выставляет `search_path`), и параллельные тесты не мешают друг другу.

## Аргументы и Kind

Задача описывается структурой аргументов. River кодирует её `encoding/json` в колонку `args`:

```go
type WelcomeEmailArgs struct {
	UserID int64  `json:"user_id"`
	Email  string `json:"email"`
}

func (WelcomeEmailArgs) Kind() string { return "welcome_email" }
```

`Kind()` это стабильное имя типа задачи. По нему воркер находит задачу в таблице, поэтому:

- **не меняй Kind после выкатки**: в таблице останутся задачи со старым именем, и их никто не возьмёт (для переименования есть `JobArgsWithKindAliases`);
- имя проверяется регулярным выражением: без пробелов и запятых, минимум два символа.

Аргументы это **данные, а не объекты**: только то, что переживёт JSON. Клади id, а не всю сущность: к моменту выполнения сущность может измениться, и воркер должен прочитать актуальную версию из базы.

Настройки по умолчанию для всех задач типа задаются методом `InsertOpts`:

```go
func (WelcomeEmailArgs) InsertOpts() river.InsertOpts {
	return river.InsertOpts{Queue: "email", MaxAttempts: 5}
}
```

Опции, переданные при вставке, перекрывают их. Без опций задача попадает в очередь `default` с приоритетом 1 (самый высокий из 1-4) и 25 попытками.

```go
type SendArgs struct {
	UserID int64  `json:"user_id"`
	token  string // неэкспортируемое поле
	Secret string `json:"-"`
}

b, _ := json.Marshal(SendArgs{UserID: 7, token: "t", Secret: "s"})
var back SendArgs
_ = json.Unmarshal(b, &back)
fmt.Println(string(b), back.UserID, back.token == "", back.Secret == "")
```

<Predict :options="['{&quot;user_id&quot;:7,&quot;token&quot;:&quot;t&quot;} 7 false true', '{&quot;user_id&quot;:7} 7 true true', '{&quot;user_id&quot;:7,&quot;Secret&quot;:&quot;s&quot;} 7 true false', 'ошибка: неэкспортируемое поле']" :correct="1">

`encoding/json` видит только экспортируемые поля, а `json:"-"` исключает поле явно. Воркер получит задачу, где `token` и `Secret` пустые, и никакой ошибки не будет. Отсюда правило: аргументы River это открытая JSON-структура, не прячь туда «служебное» в маленьких буквах. И не клади секреты: `args` лежат в базе открытым текстом и видны в River UI.

</Predict>

## Worker

Воркер это тип, который умеет выполнять задачи одного Kind:

```go
type WelcomeEmailWorker struct {
	river.WorkerDefaults[WelcomeEmailArgs] // NextRetry, Timeout, Middleware по умолчанию

	Pool   *pgxpool.Pool
	Mailer Mailer
}

func (w *WelcomeEmailWorker) Work(ctx context.Context, job *river.Job[WelcomeEmailArgs]) error {
	// job.Args типизирован, job.JobRow даёт ID, Attempt, Errors, Queue...
	return w.Mailer.Send(ctx, job.Args.Email, "Добро пожаловать", "...")
}
```

Что значит результат `Work`:

| Вернул | Что будет |
|---|---|
| `nil` | `completed` |
| ошибку | `retryable` с backoff, а если попытки кончились, `discarded` |
| `river.JobCancel(err)` | `cancelled` сразу, сколько бы попыток ни осталось |
| `river.JobSnooze(d)` | `scheduled` через `d`, попытка не тратится (урок 3) |
| паника | как ошибка: River перехватит её и запишет стек в `errors` |

Уточнение к таблице: если до повтора или конца snooze не больше 5 секунд (первые ретраи по умолчанию, короткий snooze), River ставит задачу сразу в `available` со временем запуска в будущем, а не в `retryable` или `scheduled`. Почему так, разберём в уроке 3.

Зависимости (пул, почтовый клиент) кладутся в поля структуры воркера, это обычный DI руками. Для маленьких задач без зависимостей есть `river.WorkFunc`:

```go
river.AddWorker(workers, river.WorkFunc(func(ctx context.Context, job *river.Job[PingArgs]) error {
	return nil
}))
```

Воркеры регистрируются в наборе `Workers`, по одному на Kind:

```go
workers := river.NewWorkers()
river.AddWorker(workers, &WelcomeEmailWorker{Pool: pool, Mailer: mailer}) // паникует на дубле Kind
```

`river.AddWorkerSafely` делает то же, но возвращает ошибку вместо паники.

::: trap Work не слушает ctx
```go
// плохо: при остановке клиента или таймауте задача висит до конца sleep
func (w *ExportWorker) Work(ctx context.Context, job *river.Job[ExportArgs]) error {
	time.Sleep(5 * time.Minute)
	return w.upload(job.Args.ID)
}
```

```go
// хорошо: ждём либо работу, либо отмену
func (w *ExportWorker) Work(ctx context.Context, job *river.Job[ExportArgs]) error {
	select {
	case <-time.After(5 * time.Minute):
	case <-ctx.Done():
		return ctx.Err()
	}
	return w.upload(ctx, job.Args.ID) // и ctx дальше во все вызовы
}
```

River отменяет `ctx` при таймауте задачи (по умолчанию 1 минута, `JobTimeout`) и при жёсткой остановке клиента. Код, который игнорирует `ctx`, нельзя остановить ничем, кроме убийства процесса. И верни ошибку: если вернуть `nil` после отмены, River запишет задачу как выполненную.
:::

## Client и очереди

Клиент связывает драйвер, воркеры и очереди:

```go
client, err := river.NewClient(riverpgxv5.New(pool), &river.Config{
	Queues: map[string]river.QueueConfig{
		river.QueueDefault: {MaxWorkers: 10},
		"email":            {MaxWorkers: 2},
	},
	Workers:         workers,
	SoftStopTimeout: 10 * time.Second,
})
if err != nil {
	return err
}
if err := client.Start(ctx); err != nil {
	return err
}
```

- `TTx` у клиента это тип транзакции драйвера: для `riverpgxv5` это `pgx.Tx`, поэтому тип клиента `*river.Client[pgx.Tx]`.
- `MaxWorkers` это число горутин на очередь **в этом процессе**. Три реплики сервиса по 2 воркера дают 6 параллельных писем.
- Очередь, которой нет в `Queues`, этот клиент не обрабатывает. Задачи туда вставить можно, они будут ждать клиента, у которого такая очередь есть.
- **Клиент только для вставки**: без `Queues` и без `Start`. Его используют API-процессы, когда воркеры живут отдельно. `Workers` всё равно стоит передать: тогда River при вставке проверит, что для Kind есть воркер.

Отдельные очереди нужны для изоляции: тысяча медленных отчётов в `default` не должна задерживать письмо о сбросе пароля. Внутри очереди порядок выдачи задаётся приоритетом (1 раньше 4), потом временем.

```go
type ShortArgs struct{}

func (ShortArgs) Kind() string { return "x" }

// ... воркер для ShortArgs зарегистрирован через AddWorkerSafely без ошибки
_, err = river.NewClient(riverpgxv5.New(nil), &river.Config{Workers: workers})
fmt.Println("client:", err)
```

<Predict :options="['client: <nil>', 'client: job kind &quot;x&quot; should match regex ...', 'паника в AddWorkerSafely', 'ошибка появится только при Insert']" :correct="1">

`NewClient` проверяет формат Kind всех зарегистрированных воркеров и отказывается создаваться: имя должно быть не короче двух символов и без пробелов и запятых. Полный текст в v0.51.0: `job kind "x" should match regex \A[\w][\w\-\[\]<>\/.·:+]+\z`. Ошибка всплывает на старте сервиса, а не через неделю на первой вставке, и это хорошо.

</Predict>

## Insert и InsertTx

```go
// вне транзакции: задача видна воркерам сразу
res, err := client.Insert(ctx, WelcomeEmailArgs{UserID: id, Email: email}, nil)

// в транзакции: задача станет видна после tx.Commit, исчезнет при tx.Rollback
res, err := client.InsertTx(ctx, tx, WelcomeEmailArgs{UserID: id, Email: email}, nil)
```

`res` имеет тип `*rivertype.JobInsertResult`: в `res.Job` вставленная строка (ID, состояние, очередь), а `res.UniqueSkippedAsDuplicate` пригодится в уроке про уникальность. Для пачки есть `InsertMany` и `InsertManyTx`, а для очень больших пачек `InsertManyFast`, который пишет через `COPY FROM`.

Регистрация пользователя целиком:

```go
func Register(ctx context.Context, pool *pgxpool.Pool, client *river.Client[pgx.Tx], email string) (int64, error) {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx) // после Commit это no-op

	var id int64
	if err := tx.QueryRow(ctx,
		`INSERT INTO users (email) VALUES ($1) RETURNING id`, email).Scan(&id); err != nil {
		return 0, err
	}
	if _, err := client.InsertTx(ctx, tx, WelcomeEmailArgs{UserID: id, Email: email}, nil); err != nil {
		return 0, err
	}
	return id, tx.Commit(ctx)
}
```

Сравни с тем, что было бы с брокером: вставка пользователя и публикация в брокер это две системы, нужен outbox и ретранслятор. Здесь outbox и очередь это одна и та же таблица, ретранслятор не нужен.

::: hood Почему задача не видна до коммита
`river_job` это обычная таблица. Строка, вставленная в незакоммиченной транзакции, по правилам MVCC невидима для других транзакций, включая producer-ов River. `pg_notify` в транзакции тоже копится и доставляется слушателям только после коммита. Откат стирает и строку, и уведомление. Никакой магии River тут нет: всё обеспечивает PostgreSQL, River лишь не мешает этим пользоваться.
:::

::: trap Insert вместо InsertTx
```go
// плохо: задача вставлена в отдельной транзакции через пул
tx, _ := pool.Begin(ctx)
id := createUser(ctx, tx, email)
client.Insert(ctx, WelcomeEmailArgs{UserID: id}, nil) // уже закоммичена!
if err := chargeTrial(ctx, tx, id); err != nil {
	tx.Rollback(ctx) // пользователя нет, а письмо уйдёт
	return err
}
```

Хуже того, воркер может взять задачу раньше, чем закоммитится пользователь, не найти его и отменить письмо. `Insert` внутри открытой транзакции почти всегда баг: если рядом есть `tx`, задача ставится через `InsertTx(ctx, tx, ...)`. В тестах это ловится так: вставь в транзакции, откати, `rivertest.RequireNotInserted` или `SELECT count(*) FROM river_job`.
:::

## Остановка клиента

При деплое процесс получает SIGTERM. Нельзя просто выйти: задачи посреди выполнения останутся в `running` и вернутся в работу только через час (rescuer). River умеет останавливаться в два этапа:

1. **Мягкая остановка** (`client.Stop(ctx)`): новые задачи не берутся, текущим даётся доработать.
2. **Жёсткая**: контексты выполняющихся задач отменяются. Сразу при `client.StopAndCancel(ctx)`, а при `Stop` только если задан `Config.SoftStopTimeout` и задачи за это время не закончились. Без `SoftStopTimeout` мягкая остановка ждёт задачи сколько угодно (пока не отменится `ctx`, переданный в `Stop`).

Отдельная ловушка: **отмена контекста, переданного в `Start`**. Документация `Client.Start` и исходник `client.go` v0.51.0 говорят прямо: если `SoftStopTimeout` не задан, отмена этого контекста равна `StopAndCancel`, то есть жёсткой остановке, потому что контексты задач наследуются от него. Если `SoftStopTimeout` задан (появился в v0.38.0), контексты задач от него отвязаны, и отмена запускает мягкую остановку с переходом в жёсткую по таймауту. Поэтому `signal.NotifyContext`, переданный прямо в `Start`, без `SoftStopTimeout` обрывает письма на полуслове при каждом деплое.

Что происходит с прерванной задачей. Хорошо написанный `Work` на отмене возвращает `ctx.Err()`. Если отмену вызвала остановка клиента (`Stop` с эскалацией, `StopAndCancel`), River v0.44+ не считает это ошибкой: задача сразу возвращается в `available` без траты попытки и без записи в `errors` (событие `job_interrupted`), и её подхватит этот или другой процесс. Исключение: отмена контекста `Start` без `SoftStopTimeout`. Контекст задачи отменяется вместе с ним, причина отмены не помечена как остановка клиента, и River записывает обычную ошибку с тратой попытки. Если же задача вернула `nil`, River запишет её как выполненную.

```go
ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
defer stop()

// Start получает контекст, который НЕ отменяется сигналом: остановкой
// управляем сами, ниже, после HTTP-сервера.
if err := client.Start(context.WithoutCancel(ctx)); err != nil {
	return err
}
// ... HTTP-сервер и прочее

<-ctx.Done() // пришёл сигнал, сначала останавливаем HTTP

stopCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()
// мягкая остановка; с SoftStopTimeout 10 с через 10 с River отменит ctx задач
if err := client.Stop(stopCtx); err != nil {
	return err
}
// пул закрываем только после остановки River
```

`Stop` блокируется до полной остановки или до отмены своего `ctx`. Канал `client.Stopped()` закрывается, когда клиент остановился, это удобно, если остановку инициировала отмена контекста `Start`.

::: trap pool.Close раньше client.Stop
```go
// плохо: defer-ы выполняются в обратном порядке
client.Start(ctx)
defer client.Stop(context.Background())
defer pool.Close() // выполнится ПЕРВЫМ
```

`pool.Close()` ждёт, пока все соединения вернутся в пул, а River держит соединения (слушатель LISTEN, выполняющиеся задачи). Процесс повиснет на выходе. Порядок: сначала HTTP-сервер (чтобы запросы могли дописать задачи), потом River, потом пул. В проекте трекера это сделано defer-ом, объявленным сразу после `Start` и, значит, после `defer pool.Close()`.
:::

## Подписка на события

Клиент умеет сообщать о том, что произошло с задачами. Это нужно для метрик, логов и, главное, для тестов без `time.Sleep`:

```go
events, cancel := client.Subscribe(river.EventKindJobCompleted, river.EventKindJobFailed)
defer cancel()

select {
case ev := <-events:
	fmt.Println(ev.Kind, ev.Job.ID, ev.Job.State)
case <-time.After(10 * time.Second):
	t.Fatal("задача не выполнилась за 10 секунд")
}
```

Так устроен сквозной тест в домашке: настоящий клиент, настоящая очередь, ожидание события с таймаутом вместо «поспим секунду и проверим».

## Что выведет программа?

```go
tx, _ := pool.Begin(ctx)
_, _ = client.InsertTx(ctx, tx, WelcomeEmailArgs{UserID: 1}, nil)
_, _ = client.Insert(ctx, WelcomeEmailArgs{UserID: 2}, nil)
_ = tx.Rollback(ctx)

var n int
_ = pool.QueryRow(ctx, `SELECT count(*) FROM river_job`).Scan(&n)
fmt.Println(n)
```

<Predict :options="['0', '1', '2', 'зависит от того, запущен ли клиент']" :correct="1">

Одна задача. `InsertTx` писал в транзакцию `tx`, откат её стёр. `Insert` работал через пул в своей короткой транзакции и закоммитился сразу, на него `tx.Rollback` не влияет. Запущен ли клиент, неважно: вставка работает и у клиента без `Start`.

</Predict>

```go
res, _ := client.Insert(ctx, PingArgs{}, nil) // PingArgs без метода InsertOpts
fmt.Println(res.Job.State, res.Job.Queue, res.Job.Priority, res.Job.MaxAttempts)
```

<Predict :options="['scheduled default 4 3', 'available default 1 25', 'pending default 2 25', 'available ping 1 10']" :correct="1">

Задача без опций сразу доступна (`available`), в очереди `default` (`river.QueueDefault`), с приоритетом 1 (`river.PriorityDefault`, самый высокий) и 25 попытками (`river.MaxAttemptsDefault`). 25 попыток при backoff `попытка^4` секунд растягиваются примерно на 20 дней, это стоит помнить, когда выбираешь `MaxAttempts` для своих задач.

</Predict>

::: interview
<details><summary>Почему River хранит задачи в PostgreSQL и в чём главный плюс?</summary>

Задачу можно вставить в той же транзакции, что и бизнес-данные (`InsertTx`). Она появится в очереди тогда и только тогда, когда закоммитится транзакция, и не будет видна воркерам до коммита. Это убирает проблему двойной записи без outbox и ретранслятора, не нужна отдельная инфраструктура. Цена: нагрузка на базу и потолок пропускной способности ниже, чем у брокера.

</details>
<details><summary>Как воркеры River узнают о новых задачах?</summary>

Через LISTEN/NOTIFY: при вставке клиент делает `pg_notify` в той же транзакции, уведомление доставляется после коммита, producer нужной очереди забирает задачи через `FOR UPDATE SKIP LOCKED`. Плюс поллинг раз в `FetchPollInterval` (1 с по умолчанию) на случай пропущенных уведомлений. Для окружений без LISTEN/NOTIFY (PgBouncer в transaction pooling) есть режим `PollOnly`.

</details>
<details><summary>Какие состояния есть у задачи River?</summary>

`available` (ждёт воркера), `scheduled` (запланирована на будущее, сюда же уходит snooze), `pending` (не готова, пока не переведут явно), `running`, `retryable` (упала, ждёт повтора), `completed`, `cancelled` (JobCancel или отмена вручную), `discarded` (попытки кончились). Финальные состояния удаляет чистильщик по сроку хранения.

</details>
<details><summary>Зачем River выбирает лидера?</summary>

Обслуживающие процессы должны работать в одном экземпляре на кластер: перевод запланированных и повторяемых задач в `available`, возврат застрявших задач, удаление старых, постановка периодических задач. Лидер выбирается через одну строку `river_leader` с `INSERT ... ON CONFLICT DO NOTHING` и сроком, который лидер продлевает. Умер лидер: срок истёк, выбирается другой.

</details>
<details><summary>Чем Insert отличается от InsertTx и когда что использовать?</summary>

`Insert` вставляет задачу в своей транзакции через пул, она видна сразу. `InsertTx` вставляет в переданную транзакцию: видна после коммита, исчезает при откате. Если задача связана с изменением данных, всегда `InsertTx` в той же транзакции. `Insert` уместен, когда задача ни с чем не связана: ручной запуск отчёта из админки, постановка из cron.

</details>
<details><summary>Что значит, что Work вернул ошибку, nil, JobCancel?</summary>

`nil`: задача выполнена. Ошибка: повтор по backoff, пока не кончатся `MaxAttempts`, потом `discarded`. `JobCancel(err)`: задача сразу `cancelled`, повторов не будет, для постоянных ошибок. `JobSnooze(d)`: отложить на `d` без траты попытки. Паника перехватывается и считается ошибкой.

</details>
<details><summary>Как корректно остановить клиент River?</summary>

`Stop(ctx)`: перестаём брать новые задачи и ждём текущие; если задан `SoftStopTimeout`, по его истечении контексты задач отменяются (жёсткая остановка), без него ждём сколько угодно. `StopAndCancel` отменяет сразу. Отмена контекста, переданного в `Start`, без `SoftStopTimeout` равна `StopAndCancel`, а с ним запускает мягкую остановку. Задачи, прерванные через `Stop` (по таймауту) или `StopAndCancel` и вернувшие `ctx.Err()`, River возвращает в `available` без траты попытки. А при отмене контекста `Start` без `SoftStopTimeout` контекст задачи отменяется вместе с ним, River видит обычную ошибку отмены, и попытка тратится, как при любой ошибке. Останавливать River после HTTP-сервера и до закрытия пула. Задачи, которые игнорируют `ctx`, не остановить ничем.

</details>
<details><summary>Почему в аргументах задачи хранят id, а не объект целиком?</summary>

Между постановкой и выполнением может пройти время, объект изменится или будет удалён. Воркер должен прочитать актуальное состояние из базы и решить, нужна ли работа. Плюс аргументы хранятся в JSON в таблице и видны в UI: большие объекты раздувают таблицу, а секреты в них утекают.

</details>
:::

## Источники

- `RU` [Паттерн Transactional Outbox: от теории до продакшена](https://habr.com/ru/articles/967974/): то, что River делает «бесплатно», если смотреть со стороны брокера; полезно сравнить.
- `EN` [River: Getting started](https://riverqueue.com/docs): установка, миграции, первый воркер и клиент.
- `EN` [River: Transactional enqueueing](https://riverqueue.com/docs/transactional-enqueueing): почему InsertTx главный аргумент за очередь в базе.
- `EN` [River: Migrations](https://riverqueue.com/docs/migrations): CLI, `rivermigrate`, выгрузка SQL для goose.
- `EN` [River: Graceful shutdown](https://riverqueue.com/docs/graceful-shutdown): мягкая и жёсткая остановка, `SoftStopTimeout`.
- `EN` [Документация пакета river на pkg.go.dev](https://pkg.go.dev/github.com/riverqueue/river): `Config`, `Client`, `Worker`, примеры `Example...`.
- `EN` [Transactionally Staged Job Drains in Postgres](https://brandur.org/job-drain): статья одного из авторов River о транзакционной постановке задач.

## Домашка

Пакет `homework/m09/l02-river-basics/`. Тесты сами мигрируют таблицы River в схему каждого теста (`helpers_test.go`). Перед запуском `cd homework && docker compose up -d`.

<Hw id="ex1" title="Регистрация с письмом в одной транзакции" level="база">

`ex1_register.go`: `Kind` и `InsertOpts` для `WelcomeEmailArgs` (очередь `email`, 5 попыток) и `RegisterUserTx`. Тесты проверяют задачу внутри транзакции через `rivertest.RequireInsertedTx`, её невидимость снаружи до коммита и отсутствие после отката.

```bash
cd homework && go test -v ./m09/l02-river-basics/ -run 'TestWelcomeEmailArgs|TestRegisterUserTx'
```

</Hw>

<Hw id="ex2" title="Идемпотентный воркер" level="база">

`ex2_worker.go`, метод `WelcomeEmailWorker.Work`: отмена для удалённого пользователя, повтор при ошибке почты, отметка `welcome_sent_at` строго после отправки, второй запуск без второго письма. Тест выполняет воркер синхронно через `rivertest.Worker`.

```bash
cd homework && go test -v ./m09/l02-river-basics/ -run TestWelcomeEmailWorker
```

</Hw>

<Hw id="ex3" title="Клиент с очередями" level="посложнее">

`ex3_client.go`, функция `NewClient`: воркер, очереди `default` и `email`, `SoftStopTimeout`. Сквозной тест запускает клиент, регистрирует пользователя через `InsertTx` и ждёт события `job_failed`, затем `job_completed` (первая отправка падает) через `Subscribe` с таймаутом. Тест видит только то, что заметно снаружи (воркер зарегистрирован, очередь `email` работает): `MaxWorkers` и `SoftStopTimeout` проверь сам, задачей, которая ждёт `ctx`, и вызовом `Stop`.

```bash
cd homework && go test -race -v ./m09/l02-river-basics/ -run TestNewClientEndToEnd
```

</Hw>

<Hw id="ex4" title="River руками" level="база">

Без тестов. Поставь CLI (`go install github.com/riverqueue/river/cmd/river@latest`), создай базу `createdb -U goschool rivertry` в контейнере и выполни `river migrate-up --line main --database-url postgres://goschool:goschool@localhost:55432/rivertry`. В `psql` посмотри `\d river_job` и индексы. Напиши программу, которая вставляет 5 задач: одну через `Insert`, две через `InsertTx` с коммитом, две через `InsertTx` с откатом. Проверь `SELECT id, kind, state, queue, priority FROM river_job`. Потом выгрузи SQL: `river migrate-get --line main --all --exclude-version 1 --up` и найди в нём частичный уникальный индекс по `unique_key`.

</Hw>
